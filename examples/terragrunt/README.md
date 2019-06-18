# Para and Terragrunt Multi-Module Setup

## Problem

In this example setup Para is used together with Terragrunt configured to execute several
interdependent modules. Because Terragrunt dynamically changes sub-directories to module
sub-dirs, Para cannot predict all possible paths where Terragrunt will run Terraform.
Supporting a use-case like that would require a deep integration between Para and Terragrunt
or a simple trick with relative symlinks.

## Solution

While this example is extremely simple, it is complex enough to demonstrate the problem and 
proposed solution expected to work in similar setups of arbitrary complexity.

Here we have 2 modules - `aaa` and `bbb` with latter depending onn the former one.
Both of them use [now obsolete but still perfect for such an example] community provider 
[`terraform-providerr-yaml`](https://github.com/ashald/terraform-provider-yaml).
First module would parse a YAML document containing a sequence of maps and expose one of them
via ann output that will be consumed by the second module.

Obviously, a problem like this can be solved by just creating `~/.terraform.d/plugins` but
this may not be acceptable in some cases (such as when several concurrent Terraform and/or
Terragrunt processes may be executed or when extra isolation is desired).

All it takes to make it work is just create `terraform.d/plugins` in the root dir and then just
setup symlinks in each of the module dirs to the `terraform.d` at the top level.

```bash
$ tree
.
├── README.md
├── aaa
│   ├── root.tf
│   ├── terraform.d -> ../terraform.d
│   └── terragrunt.hcl
├── bbb
│   ├── root.tf
│   ├── terraform.d -> ../terraform.d
│   └── terragrunt.hcl
├── state
├── terraform.d
│   └── plugins
└── terragrunt.hcl
``` 

For portability (so that your root dir does not need to be on the same absolute path everywhere
where it's used) it's advised to create relative symlinks like this:
```bash
$ cd aaa
$ ln -sF "../terraform.d" terraform.d

```

## Bonus

Don't forget that Para will download both Terragrunt and Terraform for you if they are not present
on your path. Just use Para as though `terragrunt` is on your `$PATH` and Para will take care of
the rest. By default the most recent versions will be downloaded but specific versions can be 
requested via Para config file at one of the following paths:
* para.cfg.yaml
* ~/.para/para.cfg.yaml
* /etc/para/para.cfg.yaml

as simple as:
```yaml
terraform: 0.12.2
terragrunt: 0.19.4
```

## Example

```bash
$ para terragrunt apply-all
Para is being initialized...
- Cache Dir: $TMPDIR/para-501
- Terraform: downloading to $TMPDIR/para-501/terraform/0.12.2/darwin_amd64
- Terrragrunt: downloading to $TMPDIR/para-501/terragrunt/v0.19.4/darwin_amd64
- Plugin Dir: terraform.d/plugins
- Primary Index: https://raw.githubusercontent.com/paraterraform/index/master/para.idx.yaml as of 2019-06-17T23:20:38-04:00 (providers: 8)
- Index Extensions: para.idx.d (0/0), ~/.para/para.idx.d (0/0), /etc/para/para.idx.d (0/0)
- Command: terragrunt apply-all

------------------------------------------------------------------------

[terragrunt] [/Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt] 2019/06/17 23:21:24 Running command: terraform --version
[terragrunt] 2019/06/17 23:21:24 Setting download directory for module /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt to /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/.terragrunt-cache
[terragrunt] 2019/06/17 23:21:24 Module /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt does not have an associated terraform configuration and will be skipped.
[terragrunt] 2019/06/17 23:21:24 Setting download directory for module /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/aaa to /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/aaa/.terragrunt-cache
[terragrunt] 2019/06/17 23:21:24 Setting download directory for module /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb to /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb/.terragrunt-cache
[terragrunt] 2019/06/17 23:21:24 Stack at /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt:
  => Module /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/aaa (excluded: false, dependencies: [])
  => Module /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb (excluded: false, dependencies: [/Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/aaa])
[terragrunt] 2019/06/17 23:21:24 [terragrunt]  Are you sure you want to run 'terragrunt apply' in each folder of the stack described above? (y/n)
y
[terragrunt] [/Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb] 2019/06/17 23:21:26 Module /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb must wait for 1 dependencies to finish
[terragrunt] [/Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/aaa] 2019/06/17 23:21:26 Module /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/aaa must wait for 0 dependencies to finish
[terragrunt] [/Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/aaa] 2019/06/17 23:21:26 Running module /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/aaa now
[terragrunt] [/Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/aaa] 2019/06/17 23:21:26 Reading Terragrunt config file at /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/aaa/terragrunt.hcl
[terragrunt] [/Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/aaa] 2019/06/17 23:21:26 Running command: terraform apply -input=false -auto-approve
- Para provides 3rd-party Terraform provider plugin 'yaml' version 'v2.1.0' for 'darwin_amd64' (cached)

data.yaml_list_of_strings.doc: Refreshing state...

Apply complete! Resources: 0 added, 0 changed, 0 destroyed.

Outputs:

result = {foo: xxx}
[terragrunt] [/Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/aaa] 2019/06/17 23:21:27 Module /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/aaa has finished successfully!
[terragrunt] [/Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb] 2019/06/17 23:21:27 Dependency /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/aaa of module /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb just finished successfully. Module /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb must wait on 0 more dependencies.
[terragrunt] [/Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb] 2019/06/17 23:21:27 Running module /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb now
[terragrunt] [/Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb] 2019/06/17 23:21:27 Reading Terragrunt config file at /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb/terragrunt.hcl
[terragrunt] [/Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb] 2019/06/17 23:21:27 Running command: terraform apply -input=false -auto-approve
data.terraform_remote_state.aaa: Refreshing state...
data.yaml_map_of_strings.doc: Refreshing state...

Apply complete! Resources: 0 added, 0 changed, 0 destroyed.

Outputs:

result = xxx
[terragrunt] [/Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb] 2019/06/17 23:21:28 Module /Users/ashald/workspace/oss/github.com/paraterraform/para/examples/terragrunt/bbb has finished successfully!
```