package cmd

import (
	"fmt"
	"github.com/paraterraform/para/app"
	"github.com/spf13/viper"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const usageTemplate = `Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
`

var defaultConfigCandidates = []string{
	"para.cfg.yaml",
	"~/.para.cfg.yaml",
	"/etc/para.cfg.yaml",
}

var defaultIndexCandidates = []string{
	"para.idx.yaml",
	"~/.para.idx.yaml",
	"/etc/para.idx.yaml",
	"https://raw.githubusercontent.com/paraterraform/index/master/para.idx.yaml",
}

var defaultExtensionsCandidates = []string{
	"para.d",
	"~/.para.d",
	"/etc/para.d",
}

var optionCachePath string
var optionIndex string
var optionConfig string
var optionExtensions string

var rootCmd = &cobra.Command{

	Long: `Para - the missing 3rd-party plugin manager for Terraform

Concepts
  Primary Index
    It's the main source of truth for Para. Always just 1 file is used. Unless overridden, Para uses 1st available from
    the list of pre-defined locations. Should be YAML in the format of:
      
        <kind>:
          <name>:
           <vX.Y.Z>:
             <platform>:
               url: <file://...|http://...|https://...>
               size: <size of the provider binary in bytes>
               digest: <md5|sha1|sha256|sha512>:<hash of the provider binary>

    All strings (key & values, except for URLs) must be lowercase. All fields are required (url, size, digest).

    URLs may point to archives and they will be automatically extracted (size and digest should be derived from the
    actual binaries rather than archives) if supported (determined by the extension at the end of the URL):
      * .zip
      * .tar
      * .tar.gz  or .tgz
      * .tar.bz2 or .tbz2
      * .tar.xz  or .txz
      * .tar.lz4 or .tlz4
      * .tar.sz  or .tsz
      * .rar

  Index Extensions
    Can be used to add or override entries in the primary index. May come handy when one is happy with the remote index
    but needs some extra plugins or if one needs to use an alternative implementation for a give plugin.

    Both file names and file context is used when processing index extensions. Only files matching the pattern 
    '<kind>.<name>.yaml' are loaded and they should be valid single-document YAMLs with the following structure:

	     <vX.Y.Z>:
		   <platform>:
		     url: <file://...|http://...|https://...>
		     size: <size of the provider binary in bytes>
		     digest: <md5|sha1|sha256|sha512>:<hash of the provider binary>

    By default Para loads all extensions from all pre-defined locations but if an explicit location is specified then
    it's the only one used.

  Cache Dir
    When Para fetches remote files it stores them briefly in the $TMPDIR but then caches them in the designated cache
    dir. As per the well-known joke, cache invalidation is too ambitious challenge so Para doesn't do anything about it.
    By default cache is stored in $TMPDIR so that it will be cleared on reboots. It's possible to configure Para to
    store cache elsewhere but then it's user's responsibility to manage it in case it grows too big.
    Cache dir facilitates offline operation. 

  Config File
    Any of the flags below (except for config itself and help flag) can be provided via a config file. It's if value is
    not provided via a flag, config file is discovered from one of pre-defined locations.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println(
				`Para - the missing 3rd-party plugin manager for Terraform.

Overview
  Para, together with Terraform, is a reference to the concept of paraterraforming.

  As paraterraforming is an option until terraforming is possible - Para takes care of distributing 3rd party plugins
  for Terraform until it's implemented in Terraform.

  Para uses FUSE to mount a virtual file system over well-known Terraform plugin locations (such as terraform.d/plugins
  and ~/.terraform.d/plugins - see https://www.terraform.io/docs/extend/how-terraform-works.html#plugin-locations for
  details) and downloads them on demand (with optional caching) using a curated index (or your own).
	
  Please note that FUSE must be available (macOS requires OSXFUSE - https://osxfuse.github.io).

Usage:
  para terraform [flags] <command> [args]

Examples:
  para terraform init
  para terraform plan
  para terraform apply

Use "para -h/--help" for more information.`)
			os.Exit(1)
		}

		var indexCandidates []string

		if len(optionConfig) > 0 {
			indexCandidates = append(indexCandidates, optionConfig)
		} else {
			indexCandidates = defaultIndexCandidates
		}

		app.Execute(args, indexCandidates, optionCachePath)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	//cobra.OnInitialize(initConfig) // TODO - reactivate
	rootCmd.Flags().SetInterspersed(false)
	rootCmd.Flags().SortFlags = false
	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.Flags().StringVarP(
		&optionConfig,
		"config",
		"f",
		"",
		fmt.Sprintf(
			"config file (default - first available from: %s)",
			strings.Join(defaultConfigCandidates, ", "),
		),
	)
	rootCmd.Flags().StringVarP(
		&optionIndex,
		"index",
		"i",
		"",
		fmt.Sprintf(
			"index location (default - first available from: %s)",
			strings.Join(defaultIndexCandidates, ", "),
		),
	)
	rootCmd.Flags().StringVarP(
		&optionExtensions,
		"extensions",
		"x",
		"",
		fmt.Sprintf(
			"index extensions directory (default - union from: %s)",
			strings.Join(defaultExtensionsCandidates, ", "),
		),
	)
	rootCmd.Flags().StringVarP(
		&optionCachePath,
		"cache",
		"c",
		"",
		"cache dir (default - ~/.cache/para if exists or /tmp/para-$UID)",
	)

	_ = viper.BindPFlag("index", rootCmd.Flags().Lookup("index"))
	_ = viper.BindPFlag("cache", rootCmd.Flags().Lookup("cache"))
}

func initConfig() {
	//// Don't forget to read config either from cfgFile or from home directory!
	//if cfgFile != "" {
	//	// Use config file from the flag.
	//	viper.SetConfigFile(cfgFile)
	//} else {
	//	// Find home directory.
	//	home, err := homedir.Dir()
	//	if err != nil {
	//		fmt.Println(err)
	//		os.Exit(1)
	//	}
	//
	//	// Search config in home directory with name ".cobra" (without extension).
	//	viper.AddConfigPath(home)
	//	viper.SetConfigName("para.cfg.yaml")
	//}
	//
	//if err := viper.ReadInConfig(); err != nil {
	//	fmt.Println("Can't read config:", err)
	//	os.Exit(1)
	//}
}
