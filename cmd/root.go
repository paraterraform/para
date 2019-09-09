package cmd

import (
	"bazil.org/fuse"
	"fmt"
	"github.com/paraterraform/para/app"
	"github.com/paraterraform/para/utils"
	"github.com/spf13/viper"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	flagConfig  = "config"
	flagUnmount = "unmount"

	flagIndex      = "index"
	flagExtensions = "extensions"
	flagCache      = "cache"
	flagRefresh    = "refresh"

	flagTerraform = "terraform"
	flagTerragrunt = "terragrunt"
)

const usageTemplate = `Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
`

const helpShort = `
Para - the missing community plugin manager for Terraform.
A "swiss army knife" for Terraform and Terragrunt - just 1 tool to facilitate all your workflows.

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

Use "para -h/--help" for more information.
`

var defaultConfigCandidates = []string{
	"para.cfg.yaml",
	"~/.para/para.cfg.yaml",
	"/etc/para/para.cfg.yaml",
}

var defaultIndexCandidates = []string{
	"para.idx.yaml",
	"~/.para/para.idx.yaml",
	"/etc/para/para.idx.yaml",
	"https://raw.githubusercontent.com/paraterraform/index/master/para.idx.yaml",
}

var defaultExtensionsCandidates = []string{
	"para.idx.d",
	"~/.para/para.idx.d",
	"/etc/para/para.idx.d",
}

var optionConfig string
var optionUnmount string

var rootCmd = &cobra.Command{

	Long: `
Para - the missing community plugin manager for Terraform.
A "swiss army knife" for Terraform and Terragrunt - just 1 tool to facilitate all your workflows.

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
               digest: <md5|sha1|sha256|sha512>:<hash of the file that will be download - verified before extraction>

    All strings (key & values, except for URLs) must be lowercase. All fields are required (url, size, digest).

    URLs may point to archives and they will be automatically extracted (size MUST be always derived from the actual
    plugin binary and digest MUST be derived from the archive in such cases) if supported (determined by the extension
    at the end of the Url):
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

    Both file names and file content are used when processing index extensions. Only files matching the pattern 
    '<kind>.<name>.yaml' are loaded and they should be valid single-document YAMLs with the following structure:

	     <vX.Y.Z>:
		   <platform>:
		     url: <file://...|http://...|https://...>
		     size: <size of the provider binary in bytes>
		     digest: <md5|sha1|sha256|sha512>:<hash of the file that will be download - verified before extraction>

    Alternatively it can be a single line with a Url like <file://...|http://...|https://...> pointing to a file with
    the content as described above. An empty file would wipe out all known version for the given plugin from the primary
    index.

    By default Para loads all extensions from all pre-defined locations but if an explicit location is specified then
    it's the only one used.

  Cache Dir
    When Para fetches remote files it stores them briefly in the $TMPDIR but then caches them in the designated cache
    dir. As per the well-known joke, cache invalidation is too ambitious challenge so Para doesn't do anything about it.
    By default cache is stored in $TMPDIR so that it will be cleared on reboots. It's possible to configure Para to
    store cache elsewhere but then it's user's responsibility to manage it in case it grows too big.
    Cache dir facilitates offline operation. 

  Config File
    Any of the flags below (except for config itself as well as help and unmount flags) can be provided via a config
    file. It's if value is not provided via a flag, config file is discovered from one of pre-defined locations.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(optionUnmount) > 0 {
			fmt.Printf("Force-unmounting: '%s'\n", optionUnmount)
			err := fuse.Unmount(optionUnmount)
			if err != nil {
				fmt.Printf("* Error: %s", err)
				os.Exit(1)
			}
			os.Exit(0)
		}

		if len(args) == 0 {
			fmt.Print(helpShort)
			os.Exit(1)
		}

		var indexCandidates []string
		optionIndex := viper.GetString(flagIndex)
		if len(optionIndex) > 0 {
			indexCandidates = append(indexCandidates, optionIndex)
		} else {
			indexCandidates = defaultIndexCandidates
		}

		var extensionsCandidates []string
		optionExtensions := viper.GetString(flagExtensions)
		if len(optionExtensions) > 0 {
			extensionsCandidates = append(extensionsCandidates, optionExtensions)
		} else {
			extensionsCandidates = defaultExtensionsCandidates
		}

		optionCachePath := viper.GetString(flagCache)
		optionRefresh := viper.GetDuration(flagRefresh)

		optionTerraform := viper.GetString(flagTerraform)
		optionTerragrunt := viper.GetString(flagTerragrunt)
		app.Execute(
			args, indexCandidates, extensionsCandidates, optionCachePath, optionRefresh, optionTerraform, optionTerragrunt,
		)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	fmt.Println("Para is being initialized...")

	cobra.OnInitialize(initConfig)
	rootCmd.Flags().SetInterspersed(false)
	rootCmd.Flags().SortFlags = false
	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.Flags().StringVarP(
		&optionConfig,
		flagConfig,
		"f",
		"",
		fmt.Sprintf(
			"config file (default - first available from: %s)",
			strings.Join(defaultConfigCandidates, ", "),
		),
	)
	rootCmd.Flags().StringP(
		flagIndex,
		"i",
		"",
		fmt.Sprintf(
			"index location (default - first available from: %s)",
			strings.Join(defaultIndexCandidates, ", "),
		),
	)
	rootCmd.Flags().StringP(
		flagExtensions,
		"x",
		"",
		fmt.Sprintf(
			"index extensions directory (default - union from: %s)",
			strings.Join(defaultExtensionsCandidates, ", "),
		),
	)
	rootCmd.Flags().StringP(
		flagCache,
		"c",
		"",
		"cache dir (default - ~/.cache/para if exists or /tmp/para-$UID)",
	)
	rootCmd.Flags().DurationP(
		flagRefresh,
		"r",
		time.Hour,
		"attempt to refresh remote indices every given interval",
	)

	// Downloadables
	rootCmd.Flags().StringP(
		flagTerraform,
		"t",
		"",
		"Terraform version to download (default - latest)",
	)

	rootCmd.Flags().StringP(
		flagTerragrunt,
		"g",
		"",
		"Terragrunt version to download (default - latest)",
	)

	// Flags that change behavior
	rootCmd.Flags().StringVarP(
		&optionUnmount,
		flagUnmount,
		"u",
		"",
		"force unmount dir (just unmount the given dir and exit, all other flags and arguments ignored)",
	)

	_ = viper.BindPFlag(flagIndex, rootCmd.Flags().Lookup(flagIndex))
	_ = viper.BindPFlag(flagExtensions, rootCmd.Flags().Lookup(flagExtensions))
	_ = viper.BindPFlag(flagCache, rootCmd.Flags().Lookup(flagCache))
	_ = viper.BindPFlag(flagRefresh, rootCmd.Flags().Lookup(flagRefresh))
	_ = viper.BindPFlag(flagTerraform, rootCmd.Flags().Lookup(flagTerraform))
	_ = viper.BindPFlag(flagTerragrunt, rootCmd.Flags().Lookup(flagTerragrunt))
}

func initConfig() {
	var candidates []string
	var selected string

	// Don't forget to read config either from cfgFile or from home directory!
	if optionConfig != "" {
		candidates = []string{optionConfig}
		// Use config file from the flag.
		//
	} else {
		candidates = defaultConfigCandidates
	}

	// For now just sticking with YAML
	// TODO consider using viper to parse all configs and therefore support more formats
	for _, path := range candidates {
		expanded := utils.PathExpand(path)
		if utils.PathExists(expanded) {
			selected = expanded
		}
	}

	if selected != "" {
		fmt.Println("- Config File:", utils.PathSimplify(selected))
		viper.SetConfigFile(selected)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("* Can't read config:", err)
			os.Exit(1)
		}
	}
}
