resource "etcd_permission" "test_permission" {
  role       = "terraform_test_role"
  key        = "/test/terraform/"
  withprefix = true
  permission = "READWRITE"  # The options are "READ" or "READWRITE".
}
