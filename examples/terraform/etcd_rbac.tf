module "etcd_rbac" {
  source = "../_modules/etcd-rbac"

  rbac_rules  = {
    "user1" = {
      "roles": {
        "user1_role1": {
          "key": "your/path/here",
          "permission": "READWRITE"
        },
        "user1_role2": {
          "key": "your/path/here",
          "permission": "READ"
        },
      },
    },
    "user2" = {
      "roles": {
        "user2_role1": {
          "key": "your/path/here",
          "permission": "READWRITE"
        }
      }
    }
  }
}

output "rbac_helper" {
  value = module.etcd_rbac.rbac_helper
}
