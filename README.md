# Para

> Para - community plugin manager for Terraform.
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

## Installation

### Automatic

There is an automatic [launcher script](https://github.com/paraterraform/para/blob/master/para) that in about 80 lines of
Bash would download (and cache!) the right (or the latest) version of `para` whenever you need it (and check for updates).
It's suggested way of installing `para` - just download the launcher script to your Terraform/Terragrunt config dir
(dont' forget to check it into your version control system) and make it executable:
```bash
curl -Lo para https://raw.githubusercontent.com/paraterraform/para/master/para && chmod +x para 
``` 
From there on just always invoke Para as `./para`:
```bash
$ ./para terraform init
Para Launcher Activated!
- Checking para.cfg.yaml in current directory for 'version: X.Y.Z'
- Desired version: latest (latest is used when no version specified)
- Downloading Para checksums for version 'latest' to '$TMPDIR/para-501/para/latest'
- Downloading Para 'darwin' binary for version 'latest' to '$TMPDIR/para-501/para/latest'
- Starting Para from '$TMPDIR/para-501/para/latest/para_v0.3.2_darwin-amd64'

------------------------------------------------------------------------

Para is being initialized...
<rest is omitted for brevity>
```

You can request a specific version of Para by adding a `version` field to `para.cfg.yaml` next to the launcher script.

### Manual

Just download the binary for your platform from the [latest release page](https://github.com/github.com/paraterraform/para/releases/latest) with `curl`:
```bash
curl -Lo para "https://github.com/paraterraform/para/releases/latest/download/$(curl -L https://github.com/paraterraform/para/releases/latest/download/SHA256SUMS | grep -i $(uname -s) | awk '{ print $2 }')"
chmod +x para
```

It's advised to make sure that `para` is on your `$PATH` for convenience.

## Usage

Para serves as a process wrapper so just prepend it to every `terraforrm` (or `terragrunt`) command.

<div align="center">
  <img src="docs/demo.svg">
</div>

The step-by-step instruction for the demo are available [here](./docs/demo.md). 

For the rest, check the [short help output](./docs/help/short.md) or [long help output](./docs/help/long.md) by running `para` or `para -h` respectively!

## Index

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