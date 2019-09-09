# Help Output - Long

```bash
$ para -h
Para is being initialized...

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

Flags:
  -f, --config string       config file (default - first available from: para.cfg.yaml, ~/.para/para.cfg.yaml, /etc/para/para.cfg.yaml)
  -i, --index string        index location (default - first available from: para.idx.yaml, ~/.para/para.idx.yaml, /etc/para/para.idx.yaml, https://raw.githubusercontent.com/paraterraform/index/master/para.idx.yaml)
  -x, --extensions string   index extensions directory (default - union from: para.idx.d, ~/.para/para.idx.d, /etc/para/para.idx.d)
  -c, --cache string        cache dir (default - ~/.cache/para if exists or /tmp/para-$UID)
  -r, --refresh duration    attempt to refresh remote indices every given interval (default 1h0m0s)
  -t, --terraform string    Terraform version to download (default - latest)
  -g, --terragrunt string   Terragrunt version to download (default - latest)
  -u, --unmount string      force unmount dir (just unmount the given dir and exit, all other flags and arguments ignored)
  -h, --help                help for this command
``` 
