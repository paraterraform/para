remote_state {
  backend = "local"
  config = {
    path = format("%s/state/%s.tfstate", get_parent_terragrunt_dir(), path_relative_to_include())
  }
}

terraform {
  extra_arguments "common_var" {
    commands  = get_terraform_commands_that_need_vars()

    env_vars = {
      TF_VAR_state = format("%s/state", get_parent_terragrunt_dir())
    }
  }
}