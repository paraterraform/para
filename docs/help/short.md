# Help Output - Short

```bash
$ para
Para is being initialized...

Para - the missing 3rd-party plugin manager for Terraform.

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
``` 
