package app

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/paraterraform/para/app/index"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	pathPluginDirLocal = "terraform.d/plugins"
	pathPluginDirUser  = "~/.terraform.d/plugins"
)

var pluginDirCandidates = []string{
	pathPluginDirLocal,
	pathPluginDirUser,
}

func Execute(args []string, primaryIndexCandidates, indexExtensions []string, customCachePath string, refresh time.Duration) {
	var pluginDir string
	var mountpoint *string
	var stat os.FileInfo
	var err error

	fmt.Printf("Para is being initialized:\n")

	// Plugin Dir
	for _, pluginDir = range pluginDirCandidates {
		expandedPath, err := homedir.Expand(pluginDir)
		if err != nil {
			continue
		}
		if stat, err = os.Stat(expandedPath); !os.IsNotExist(err) {
			mountpoint = &expandedPath
			break
		}
	}

	if mountpoint == nil {
		fmt.Printf(
			"  Para is humble but it won't let itself be ignored! Please make sure that at least one of the "+
				"following dirs exists: %s.\n",
			strings.Join(pluginDirCandidates, ", "),
		)
		os.Exit(1)
	}

	if !stat.IsDir() {
		fmt.Printf(
			" * Error: the '%s' path exists but does not appear to be a directory - please see "+
				"https://www.terraform.io/docs/extend/how-terraform-works.html#plugin-locations\n",
			*mountpoint,
		)
		os.Exit(1)
	}
	// Check if plugin dir is in use
	pidBytes, err := ioutil.ReadFile(filepath.Join(*mountpoint, FileMeta))
	if os.IsNotExist(err) {
		fmt.Printf("  - Plugin Dir: %s\n", pluginDir)
	} else {
		if err != nil {
			fmt.Printf(
				" * Error: the previous instance of Para failed and left '%s' in a bad shape - please "+
					"run following command to clean up: para -u %s\n", pluginDir, pluginDir,
			)
		} else {
			fmt.Printf(" * Error: another instance of Para (PID: %s) uses '%s' right now - "+
				"please wait until it will finish.\n", strings.TrimSpace(string(pidBytes)), pluginDir,
			)
			if pluginDir == pathPluginDirUser {
				fmt.Printf(
					"   If the other instance is running from another Terraform configuration - please "+
						"consider creating '%s' within Terraform configuration dir to avoid contention over '%s'.\n",
					pathPluginDirLocal, pathPluginDirUser,
				)
			}
		}
		os.Exit(1)
	}

	// Cache Dir
	cacheDir, err := discoverCacheDir(customCachePath)
	if err != nil {
		fmt.Printf(
			" * Error: Para requires a writable cache dir for operation but failed discovering one: %s\n",
			err,
		)
		os.Exit(1)
	}
	fmt.Printf("  - Cache Dir: %s\n", simplifyPath(cacheDir))

	// Primary Index
	kindNameIndex, location, err := index.DiscoverIndex(primaryIndexCandidates, cacheDir, refresh)
	if err != nil {
		fmt.Printf(" * Error: cannnot decode primary index at '%s' as a valid YAML map: %s\n", location, err)
		os.Exit(1)
	}
	fmt.Printf("  - Primary Index: %s\n", location)

	// Index Extensions
	fmt.Printf("  - Index Extensions:\n")
	loadedExtensions, failedExtensions := loadExtensions(kindNameIndex, indexExtensions, refresh)
	for _, ext := range indexExtensions {
		countLoaded := loadedExtensions[ext]
		countFailed := failedExtensions[ext]
		fmt.Printf("     %s: loaded %d, errors %d\n", ext, countLoaded, countFailed)
	}

	// Command
	fmt.Printf("  - Command: %s\n", strings.Join(args, " "))

	// Footer
	fmt.Println()
	fmt.Println(strings.Repeat("-", 72))
	fmt.Println()

	// Init sub-process
	cmd := exec.Command(args[0], args[1:]...)

	// Setup signal handlers and cleanup
	var signalChan = make(chan os.Signal, 100)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range signalChan {
			if err := cmd.Process.Signal(sig); err != nil {
				fmt.Printf("* Para is unable to forward signal: %s", err) // TODO trace instead of log
			}
		}
	}()
	//
	ready, err := mountPluginsDir(kindNameIndex.BuildPlatformIndex(), *mountpoint)
	if err != nil {
		fmt.Printf("* Para was unable to mount plugin FS over '%s': %s", pluginDir, err)
		os.Exit(1)
	}
	<-ready

	// spawn sub-process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	_ = fuse.Unmount(*mountpoint)
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		fmt.Printf(
			"Para was not able to execute <%s> and failed with an error: %s",
			strings.Join(args, " "),
			err,
		)
		os.Exit(1)
	}
}

func loadExtensions(index *index.LoadingIndex, extensions []string, refresh time.Duration) (loaded map[string]uint64, failed map[string]uint64) {
	loaded = make(map[string]uint64)
	failed = make(map[string]uint64)

	for idx := len(extensions) - 1; idx >= 0; idx-- {
		path := extensions[idx]
		expandedPath, err := homedir.Expand(path)
		if err != nil {
			continue
		}
		matches, _ := ioutil.ReadDir(expandedPath)
		for _, ext := range matches {
			if ext.IsDir() { // TODO trace
				failed[path] += 1
				continue
			}

			err := index.LoadExtension(filepath.Join(expandedPath, ext.Name()), refresh)
			if err != nil {
				failed[path] += 1
			} else {
				loaded[path] += 1
			}
		}
	}
	return
}

func discoverCacheDir(customPath string) (string, error) {
	if len(customPath) > 0 {
		return customPath, nil
	}
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		userCacheDirForPara := filepath.Join(userCacheDir, "para")
		if pathExists(userCacheDirForPara) {
			return userCacheDirForPara, nil // TODO verify it's writable?
		}
	}
	path := filepath.Join(os.TempDir(), fmt.Sprintf("para-%v", os.Geteuid()))
	err = os.MkdirAll(path, 0744)
	return path, err
}

func mountPluginsDir(index *index.RuntimeIndex, mountpoint string) (<-chan struct{}, error) {
	c, err := fuse.Mount(
		mountpoint,
		fuse.VolumeName("Terraform Plugins"),
		fuse.FSName("terraform-platformToPlugins"),
		fuse.Subtype("para"),
		fuse.LocalVolume(),
		fuse.ReadOnly(),
	)

	if err != nil {
		return nil, err
	}
	go fuseRun(index, c)

	return c.Ready, nil
}

func fuseRun(index *index.RuntimeIndex, c *fuse.Conn) {
	defer func() {
		if err := c.Close(); err != nil {
			fmt.Printf("* [ASYNC] Para encountered an error: %s", err)
			os.Exit(1)
		}
	}()

	err := fs.Serve(c, FS{index: index})
	if err != nil {
		fmt.Printf("* [ASYNC] Para encountered an error: %s", err)
		os.Exit(1)
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		fmt.Printf("* [ASYNC] Para encountered an error: %s", err)
		os.Exit(1)
	}
}
