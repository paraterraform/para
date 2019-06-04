package app

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"

	"strings"
	"syscall"
)

const (
	pathPluginDirLocal = "terraform.d/plugins"
	pathPluginDirUser  = "~/.terraform.d/plugins"
)

var pluginDirCandidates = []string{
	pathPluginDirLocal,
	pathPluginDirUser,
}

func Execute(args []string, indices []string, customCachePath string) {
	var pluginDir string
	var mountpoint *string
	var stat os.FileInfo
	var err error

	log.SetFlags(0)

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

	log.Printf("Para is being initialized...")

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
		log.Printf("  - Plugin Dir: %s", pluginDir)
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

	cacheDir, err := discoverCacheDir(customCachePath)
	if err != nil {
		fmt.Printf(
			" * Error: Para requires a writable cache dir for operation but failed discovering one: %s\n",
			err,
		)
		os.Exit(1)
	}
	log.Printf("  - Cache Dir: %s", cacheDir)

	index, location, err := DiscoverIndex(indices, cacheDir)
	if err != nil {
		fmt.Printf(" * Error: cannnot decode primary index at '%s' as a valid YAML map: %s\n", location, err)
	}
	log.Printf("  - Primary Index: %s", location)
	log.Printf("  - Command: %s", strings.Join(args, " "))
	log.Printf("")
	log.Printf(strings.Repeat("-", 72))
	log.Printf("")

	// Init sub-process
	cmd := exec.Command(args[0], args[1:]...)

	// Setup signal handlers and cleanup
	var signalChan = make(chan os.Signal, 100)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range signalChan {
			if err := cmd.Process.Signal(sig); err != nil {
				log.Printf("Unable to forward signal: %s", err) // TODO trace instead of log
			}
		}
	}()
	//
	ready, err := mountPluginsDir(*index, *mountpoint)
	if err != nil {
		log.Fatalf("Para was unable to mount plugin FS over '%s': %s", pluginDir, err)
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
		log.Fatalf(
			"Para was not able to execute <%s> and failed with an error: %s",
			strings.Join(args, " "),
			err,
		)
	}
}

func discoverCacheDir(customPath string) (string, error) {
	if len(customPath) > 0 {
		return customPath, nil
	}
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		userCacheDirForPara := filepath.Join(userCacheDir, "para")
		if pathExists(userCacheDirForPara) {
			// TODO verify it's writable?
			return userCacheDirForPara, nil
		}
	}
	path := filepath.Join(os.TempDir(), fmt.Sprintf("para-%v", os.Geteuid()))
	err = os.MkdirAll(path, 0744)
	return path, err
}

func pathExists(path string) bool {
	if len(path) == 0 {
		return false
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		// TODO verify it's writable?
		return true
	}
	return false
}

func mountPluginsDir(index Index, mountpoint string) (<-chan struct{}, error) {
	c, err := fuse.Mount(
		mountpoint,
		fuse.VolumeName("Terraform Plugins"),
		fuse.FSName("terraform-plugins"),
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

func fuseRun(index Index, c *fuse.Conn) {
	defer func() {
		if err := c.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	err := fs.Serve(c, FS{index: &index})
	if err != nil {
		log.Fatal(err)
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
}
