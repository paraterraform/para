package app

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/paraterraform/para/app/index"
	"github.com/paraterraform/para/utils"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
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

func Execute(
	args []string,
	primaryIndexCandidates, indexExtensions []string,
	customCachePath string, refresh time.Duration,
	versionTerraform string,
) {
	var pluginDir string
	var mountpoint *string
	var stat os.FileInfo
	var err error

	// Cache Dir
	fmt.Printf("- Cache Dir: ")
	cacheDir, err := discoverCacheDir(customCachePath)
	if err != nil {
		fmt.Printf(
			"\n* Error: Para requires a writable cache dir for operation but failed discovering one: %s\n",
			err,
		)
		os.Exit(1)
	}
	fmt.Println(utils.PathSimplify(cacheDir))

	cmd := args[0]
	if cmd == terraformExec || cmd == terragruntExec {
		fmt.Printf("- Terraform: ")
		terraformExisting, err := exec.LookPath(terraformExec)
		if err != nil {
			// No terraform - need to download it
			fmt.Print("downloading")
			terraformDir, err := downloadTerraform(versionTerraform, cacheDir, refresh)
			if err != nil {
				fmt.Printf("\n* Error: Para was unable to download Terraform: %s\n", err)
				os.Exit(1)
			}
			err = appendToPath(terraformDir)
			if err != nil {
				fmt.Printf("\n* Error: Para was unable to add Terraform to $PATH: %s\n", err)
				os.Exit(1)
			}
			fmt.Printf(" to %s\n", utils.PathSimplify(terraformDir))
		} else {
			fmt.Printf("found at %s\n", utils.PathSimplify(terraformExisting))
		}
	}
	if cmd == terragruntExec {
		terragruntExisting, err := exec.LookPath(terragruntExec)
		fmt.Printf("- Terrragrunt: ")
		if err != nil {
			// No terragrunt - need to download it
			fmt.Print("downloading")
			terragruntDir, err := downloadTerragrunt(versionTerraform, cacheDir, refresh)
			if err != nil {
				fmt.Printf("\n* Error: Para was unable to download Terragrunt: %s\n", err)
				os.Exit(1)
			}
			err = appendToPath(terragruntDir)
			if err != nil {
				fmt.Printf("\n* Error: Para was unable to add Terragrunt to $PATH: %s\n", err)
				os.Exit(1)
			}
			fmt.Printf(" to %s\n", utils.PathSimplify(terragruntDir))
		} else {
			fmt.Printf("found at %s\n", utils.PathSimplify(terragruntExisting))
		}
	}

	// Plugin Dir
	fmt.Printf("- Plugin Dir: ")
	for _, pluginDir = range pluginDirCandidates {
		expandedPath := utils.PathExpand(pluginDir)

		if stat, err = os.Stat(expandedPath); !os.IsNotExist(err) {
			mountpoint = &expandedPath
			break
		}
	}

	if mountpoint == nil {
		fmt.Printf(
			"\n* Para is humble but it won't let itself be ignored! Please make sure that at least one of the "+
				"following dirs exists: %s.\n",
			strings.Join(pluginDirCandidates, ", "),
		)
		os.Exit(1)
	}

	if !stat.IsDir() {
		fmt.Printf(
			"\n* Error: the '%s' path exists but does not appear to be a directory - please see "+
				"https://www.terraform.io/docs/extend/how-terraform-works.html#plugin-locations\n",
			*mountpoint,
		)
		os.Exit(1)
	}
	// Check if plugin dir is in use
	pidBytes, err := ioutil.ReadFile(filepath.Join(*mountpoint, FileMeta))
	if os.IsNotExist(err) {
		fmt.Println(pluginDir)
	} else {
		if err != nil {
			fmt.Printf(
				"\n* Error: the previous instance of Para failed and left '%s' in a bad shape - please "+
					"run following command to clean up: para -u %s\n", pluginDir, pluginDir,
			)
		} else {
			fmt.Printf("\n* Error: another instance of Para (PID: %s) uses '%s' right now - "+
				"please wait until it will finish.\n", strings.TrimSpace(string(pidBytes)), pluginDir,
			)
			if pluginDir == pathPluginDirUser {
				fmt.Printf(
					" If the other instance is running from another Terraform configuration - please "+
						"consider creating '%s' within Terraform configuration dir to avoid contention over '%s'.\n",
					pathPluginDirLocal, pathPluginDirUser,
				)
			}
		}
		os.Exit(1)
	}

	// Primary Index
	fmt.Printf("- Primary Index: ")
	loadingIndex, err := index.DiscoverIndex(primaryIndexCandidates, cacheDir, refresh)
	if err != nil {
		fmt.Printf("\n* Error: cannnot decode primary index as a valid YAML map: %s\n", err)
		os.Exit(1)
	}
	var indexStats []string
	for kind, nameToPlugins := range loadingIndex.KindToNameToPlugins {
		indexStats = append(indexStats, fmt.Sprintf("%ss: %d", kind, len(nameToPlugins)))
	}
	sort.Strings(indexStats)
	fmt.Printf(
		"%s as of %s (%s)\n",
		loadingIndex.Location,
		loadingIndex.Timestamp.Format(time.RFC3339),
		strings.Join(indexStats, ", "),
	)

	// Index Extensions
	fmt.Printf("- Index Extensions: ")
	loadedExtensions, failedExtensions := loadExtensions(loadingIndex, indexExtensions)
	var extensionsStats []string
	for _, ext := range indexExtensions {
		countLoaded := loadedExtensions[ext]
		countFailed := failedExtensions[ext]
		extensionsStats = append(
			extensionsStats,
			fmt.Sprintf("%s (%d/%d)", ext, countLoaded, countLoaded+countFailed),
		)
	}
	fmt.Printf("%s\n", strings.Join(extensionsStats, ", "))

	// Command
	fmt.Printf("- Command: %s\n", strings.Join(args, " "))

	// Footer
	fmt.Println()
	fmt.Println(strings.Repeat("-", 72))
	fmt.Println()

	// Init sub-process
	subprocess := exec.Command(cmd, args[1:]...)

	// Setup signal handlers and cleanup
	var signalChan = make(chan os.Signal, 100)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range signalChan {
			if err := subprocess.Process.Signal(sig); err != nil {
				fmt.Printf("* Para is unable to forward signal: %s", err) // TODO trace instead of log
			}
		}
	}()
	//
	ready, err := mountPluginsDir(loadingIndex.BuildRuntimeIndex(), *mountpoint)
	if err != nil {
		fmt.Printf("* Para was unable to mount plugin FS over '%s': %s", pluginDir, err)
		os.Exit(1)
	}
	<-ready

	// spawn sub-process
	subprocess.Stdin = os.Stdin
	subprocess.Stdout = os.Stdout
	subprocess.Stderr = os.Stderr

	err = subprocess.Run()
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

func loadExtensions(index *index.LoadingIndex, extensions []string) (loaded map[string]uint64, failed map[string]uint64) {
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

			err := index.LoadExtension(filepath.Join(expandedPath, ext.Name()))
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
		if utils.PathExists(userCacheDirForPara) {
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
		fuse.AsyncRead(),
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

func appendToPath(new string) error {
	name := "PATH"
	current, _ := os.LookupEnv(name)
	value := strings.Join(append(strings.Split(current, ":"), new), ":")
	return os.Setenv(name, value)
}
