resource "etcd_role" "backend-role" {
  for_each = {
    for users in local.rbac_helpers : "${users.user}-${users.role}" => users
  }
  name     = each.value.role
}

resource "etcd_permission" "backend_permission" {
  for_each = {
    for users in local.rbac_helpers : "${users.user}-${users.role}" => users
  }
  depends_on  = [
    etcd_role.backend-role
  ]
  role       = each.value.role
  key        = each.value.key
  withprefix = true
  permission = each.value.permission # The options are "READ" or "READWRITE".
}

resource "etcd_role_user" "backend_role_grant_user" {
  for_each = {
    for users in local.rbac_helpers : "${users.user}-${users.role}" => users
  }
  depends_on  = [
    etcd_role.backend-role
  ]
  user_name = each.value.user
  role      = each.value.role
}

output "rbac_helper" {
  value = local.rbac_helpers
}
