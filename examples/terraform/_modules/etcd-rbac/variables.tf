variable "rbac_rules" {
  default = {
    "user1" = {
      "roles": {
        "user1_role1": {
          "key": "your/path/here",
          "permission": "READWRITE"
        },
        "user1_role2": {
          "key": "your/path/here",
          "permission": "READWRITE"
        },
      },
    }
  }
}

locals {
  rbac_helpers = flatten([
    for user, roles_info in var.rbac_rules: [
      for role, role_descr in roles_info.roles: {
        user = user
        role = role
        key  = role_descr.key
        permission = role_descr.permission
      }
    ]
  ])
}
