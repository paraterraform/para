terraform {
  backend "local" {}
}

data "yaml_list_of_strings" "doc" {
  input = <<YAML
- foo: xxx
- bar: yyy
YAML
}

output "result" { value=data.yaml_list_of_strings.doc.output[0] }