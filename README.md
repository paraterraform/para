# Para

> Para - the missing community plugin manager for Terraform.
> A "swiss army knife" for [Terraform] and [Terragrunt] - just 1 tool to facilitate all your workflows.
 
## Overview

<a href="https://en.wiktionary.org/wiki/paraterraform#English" target="_blank"><img align="right" src="docs/para.png"></a>
Para, together with Terraform, is a reference to the concept of [paraterraforming](https://en.wikipedia.org/wiki/Terraforming#Paraterraforming).

As paraterraforming is an option until terraforming is possible - Para takes care of distributing community plugins
for Terraform until it's implemented in Terraform.

Para uses FUSE to mount a virtual file system over well-known Terraform plugin locations (such as `terraform.d/plugins`
and `~/.terraform.d/plugins` - see [official docs](https://www.terraform.io/docs/extend/how-terraform-works.html#plugin-locations) for
details) and downloads them on demand (with optional caching) using a [curated index](https://github.com/paraterraform/index) (or your own).

Please note that FUSE must be available (macOS requires [OSXFUSE](https://osxfuse.github.io)).

## Capabilities

* Download [community plugins](https://www.terraform.io/docs/providers/type/community-index.html) for [Terraform] on demand using a [curated default index](https://github.com/paraterraform/index) or your own
* Download [Terraform] on demand (just run it as though it's there `para terraform ...`)
* Download [Terragrunt] on demand (just run it as though it's there `para terragrunt ...`)

## Examples

Please see [examples](./examples) for complete setups showcasing Para's usage.  

## Usage

### Install

#### Automatic

There is an automatic [launcher script](https://github.com/paraterraform/para/blob/master/para) that in about 80 lines of
Bash would download (and cache!) the right (or the latest) version of `para` whenever you need it (and check for updates).
It's suggested way of installing `para` - just download the launcher script to your Terraform/Terragrunt config dir
(dont' forget to check it into your version control system) and make it executable:
```bash
curl --location --output para https://raw.githubusercontent.com/paraterraform/para/master/para
chmod +x para 
``` 
From there on just always invoke Para as `./para`:
```bash
$ ./para 
Para Launcher Activated!
- Checking para.cfg.yaml in current directory for 'version: X.Y.Z'
- Desired version: latest (latest is used when no version specified)
- Downloading Para checksums for version 'latest' to '$TMPDIR/para-501/para/latest'
- Downloading Para 'darwin' binary for version 'latest' to '$TMPDIR/para-501/para/latest'
- Starting Para from '$TMPDIR/para-501/para/latest/para_v0.3.1_darwin-amd64'

------------------------------------------------------------------------

Para is being initialized...
<rest is omitted for brevity>
```

You can request a specific version of Para by adding a `version` field to `para.cfg.yaml` next to the launcher script.

#### Manual

Just download the binary for your platform from the [latest release page](https://github.com/github.com/paraterraform/para/releases/latest) with `curl`:
```bash
curl -Lo para "https://github.com/paraterraform/para/releases/latest/download/$(curl -L https://github.com/paraterraform/para/releases/latest/download/SHA256SUMS | grep -i $(uname -s) | awk '{ print $2 }')"
chmod +x para
```

It's advised to make sure that `para` is on your `$PATH` for convenience.

### Run

Para serves as a process wrapper so just prepend it to every `terraforrm` (or `terragrunt`) command.

For instance, let's say we want to work with YAML using the community plugin [terraform-provider-yaml](https://github.com/ashald/terraform-provider-yaml):
```bash
$ cat > example.tf << HCL2
data "yaml_list_of_strings" "doc" {
  input = <<YAML
 - foo
 - 123
 - bar: 456
YAML
}

output "result" { value=data.yaml_list_of_strings.doc.output }
HCL2
``` 

As you can imagine, that's not going to work that easy:  

```bash
$ terraform init

Initializing the backend...

Initializing provider plugins...
- Checking for available provider plugins...

Provider "yaml" not available for installation.

A provider named "yaml" could not be found in the Terraform Registry.

This may result from mistyping the provider name, or the given provider may
be a third-party provider that cannot be installed automatically.

In the latter case, the plugin must be installed manually by locating and
downloading a suitable distribution package and placing the plugin's executable
file in the following directory:
    terraform.d/plugins/darwin_amd64

Terraform detects necessary plugins by inspecting the configuration and state.
To view the provider versions requested by each module, run
"terraform providers".

```

Now, let's try `para`!
```bash
$ ./para terraform init
Para is being initialized...
- Plugin Dir:
* Para is humble but it won't let itself be ignored! Please make sure that at least one of the following dirs exists: terraform.d/plugins, ~/.terraform.d/plugins.
```

Ooops, but now you get the idea!
Let's fix that really quickly:
```bash
$ mkdir -p terraform.d/plugins
```

And try again:
```bash
$ ./para terraform init
para terraform init
Para is being initialized...
- Cache Dir: $TMPDIR/para-501
- Terraform: downloading to $TMPDIR/para-501/terraform/0.12.2/darwin_amd64
- Plugin Dir: terraform.d/plugins
- Primary Index: https://raw.githubusercontent.com/paraterraform/index/master/para.idx.yaml as of 2019-06-17T23:54:22-04:00 (providers: 8)
- Index Extensions: para.idx.d (0/0), ~/.para/para.idx.d (0/0), /etc/para/para.idx.d (0/0)
- Command: terraform init

------------------------------------------------------------------------


Initializing the backend...

Initializing provider plugins...
- Para provides 3rd-party Terraform provider plugin 'yaml' version 'v2.1.0' for 'darwin_amd64' (cached)


The following providers do not have any version constraints in configuration,
so the latest version was installed.

To prevent automatic upgrades to new major versions that may contain breaking
changes, it is recommended to add version = "..." constraints to the
corresponding provider blocks in configuration, with the constraint strings
suggested below.

* provider.yaml: version = "~> 2.1"

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
```

Voilà!

To be honest, there is one more detail to. Terraform doesn't create copies of the plugins so once you start using `para` you have to use it all the time:
```bash
$ terraform apply

Error: Could not satisfy plugin requirements


Plugin reinitialization required. Please run "terraform init".

Plugins are external binaries that Terraform uses to access and manipulate
resources. The configuration provided requires plugins which can't be located,
don't satisfy the version constraints, or are otherwise incompatible.

Terraform automatically discovers provider requirements from your
configuration, including providers used in child modules. To see the
requirements and constraints from each module, run "terraform providers".



Error: provider.yaml: no suitable version installed
  version requirements: "(any version)"
  versions installed: none
```

But!

```bash
$ ./para terraform apply
Para is being initialized...
- Cache Dir: $TMPDIR/para-501
- Terraform: downloading to $TMPDIR/para-501/terraform/0.12.2/darwin_amd64
- Plugin Dir: terraform.d/plugins
- Primary Index: https://raw.githubusercontent.com/paraterraform/index/master/para.idx.yaml as of 2019-06-17T23:54:22-04:00 (providers: 8)
- Index Extensions: para.idx.d (0/0), ~/.para/para.idx.d (0/0), /etc/para/para.idx.d (0/0)
- Command: terraform apply

------------------------------------------------------------------------

- Para provides 3rd-party Terraform provider plugin 'yaml' version 'v2.1.0' for 'darwin_amd64' (cached)

data.yaml_list_of_strings.doc: Refreshing state...

Apply complete! Resources: 0 added, 0 changed, 0 destroyed.

Outputs:

result = [
  "foo",
  "123",
  "{bar: 456}",
]
```

For the rest, check the [short help output](./docs/help/short.md) or [long help output](./docs/help/long.md) by running `para` or `para -h` respectively!

### Index

Para relies heavily on a special plugin index for discovery of 3rd party plugins.

### Primary

It's the main source of truth for Para. Always just 1 file is used. Unless overridden, Para uses 1st available from
the list of pre-defined locations:
* `para.idx.yaml`
* `~/.para/para.idx.yaml`
* `/etc/para/para.idx.yaml`
* `https://raw.githubusercontent.com/paraterraform/index/master/para.idx.yaml`

As you can see, even by default `para` will always prefer your local index if there is one before falling back on the
[default one]((https://github.com/paraterraform/index)). Alternatively, you can override it via a config, environment
variable or a command line flag. 

Primary index file should be a YAML in the format of:

```yaml
<kind>:
  <name>:
   <vX.Y.Z>:
     <platform>:
       url: <file://...|http://...|https://...>
       size: <size of the provider binary in bytes>
       digest: <md5|sha1|sha256|sha512>:<hash of the file that will be download - verified before extraction>
```

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

### Extensions

You may like the primary index in general but what if:
* something is missing from the primary index?
* you would rather use some other version for one particular plugin?
* there is a plugin that is only available in your corporate network?
* you want to have some special plugins for this particular project?
* you'd rather curate your plugins explicitly?

Those and many other use-cases can be addressed with index extensions!

By default (which you can change to your liking!) they are loaded from one of the following dirs:
* `para.idx.d`
* `~/.para/para.idx.d`
* `/etc/para/para.idx.d`

Both file names and file content are used when processing index extensions
Only files matching the pattern `<kind>.<name>.yaml` are loaded.
Each index extension file should be one of the following:
* a valid single-document YAMLs with the following structure
```yaml
<vX.Y.Z>:
   <platform>:
     url: <file://...|http://...|https://...>
     size: <size of the provider binary in bytes>
     digest: <md5|sha1|sha256|sha512>:<hash of the file that will be download - verified before extraction>
```
* a single line with a URL like `file://...|http://...|https://...` pointing to a file with the content as described above.

The former may come handy if you'd rather have your index in-place and the latter may be useful for plugin developers if
they want to maintain their own feed of plugin versions as seen in [terraform-provider-yaml](https://github.com/ashald/terraform-provider-yaml/blob/master/provider.yaml.yaml).

Alternatively, it can be used to block certain plugins as putting an empty file like `provider.foo.yaml` would wipe out
all known versions of the `prrovider` plugin named `foo` from the primary index.  

## Development

### Roadmap

Future is not set in stone but might as well look like:
* testing ¯\\\_(ツ)_/¯
* logging/tracing for debugging
* helper commands to analyze/verify indices
* helper commands to control cache state  
* serve as an adaptor for arbitrary state management backends (leverage `http` backend to convert API calls into command line calls)
* you name it!

### Guidelines

Para is written and maintained by [Borys Pierov](https://github.com/ashald).
Contributions are welcome and should follow [development guidelines](./docs/development.md).
All contributors are honored in [CONTRIBUTORS.md](./CONTRIBUTORS.md).

## License

See [LICENSE.txt](./LICENSE.txt)

[Terraform]: https://www.terraform.io
[Terragrunt]: https://github.com/gruntwork-io/terragrunt