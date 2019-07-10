variable "state" {}

terraform {
  backend "local" {}
}

data "terraform_remote_state" "aaa" {
  backend = "local"

  config = {
    path = "${var.state}/aaa.tfstate"
  }
}

data "yaml_map_of_strings" "doc" {
  input = data.terraform_remote_state.aaa.outputs.result
}

output "result" { value=data.yaml_map_of_strings.doc.output["foo"] }