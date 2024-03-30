locals {
  function_name = "go-swim"
  src_path      = "${path.module}/../"

  binary_path  = "${path.module}/tf_generated/bootstrap"
  archive_path = "${path.module}/tf_generated/bootstrap.zip"
}

output "binary_path" {
  value = local.binary_path
}
