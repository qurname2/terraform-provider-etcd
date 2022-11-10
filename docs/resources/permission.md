---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "etcd_permission Resource - terraform-provider-etcd"
subcategory: ""
description: |-
  
---

# etcd_permission (Resource)



## Example Usage

```terraform
resource "etcd_permission" "test_permission" {
  role       = "terraform_test_role"
  key        = "/test/terraform/"
  withprefix = true
  permission = "READWRITE"  # The options are "READ" or "READWRITE".
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **key** (String)
- **permission** (String)
- **role** (String)
- **withprefix** (Boolean)

### Optional

- **endrange** (String)
- **id** (String) The ID of this resource.

