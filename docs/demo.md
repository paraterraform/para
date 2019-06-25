# Para Demo

Let's say we want to work with YAML using the community plugin [terraform-provider-yaml](https://github.com/ashald/terraform-provider-yaml):
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