package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/client/v3"
)

func resourceRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceRoleCreate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second

	cli := meta.(*clientv3.Client)

	role := d.Get("name").(string)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err := cli.RoleGet(ctx, role)
	defer cancel()
	if err == nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("The role %v already exist and it is not managed by this terraform.", role)))
	}
	_, err = cli.RoleAdd(ctx, role)

	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Error with RoleAdd function")))
	}

	d.SetId(uuidGenerator())

	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second

	cli := meta.(*clientv3.Client)

	role := d.Get("name").(string)
	if role == "" {
		role = d.Id()
		d.Set("name", role)
	}
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err := cli.RoleGet(ctx, role)
	defer cancel()
	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("The role %s doesn't exist. Maybe someone removed it manually.", role)))
	}

	d.SetId(uuidGenerator())

	return nil
}
func resourceRoleUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second

	oldValue, newValue := d.GetChange("name")
	cli := meta.(*clientv3.Client)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	tflog.Info(ctx, fmt.Sprintf("oldvalue is: %v, newValue is: %v", oldValue, newValue))
	role, err := cli.RoleGet(ctx, fmt.Sprintf("%v", oldValue))
	defer cancel()
	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("The original role %v doesn't exist.", oldValue)))
	}
	_, err = cli.RoleGet(ctx, fmt.Sprintf("%v", newValue))
	if err == nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("The new role name %v already exist.", newValue)))
	}

	// Creating new role
	_, err = cli.RoleAdd(ctx, fmt.Sprintf("%v", newValue))
	if err != nil {
		return diag.FromErr(err)
	}
	for _, p := range role.Perm {
		if fmt.Sprintf("%v", p.PermType) == "READWRITE" {
			_, err = cli.RoleGrantPermission(ctx, fmt.Sprintf("%v", newValue), string(p.Key), string(p.RangeEnd), clientv3.PermissionType(clientv3.PermReadWrite))
		} else {
			_, err = cli.RoleGrantPermission(ctx, fmt.Sprintf("%v", newValue), string(p.Key), string(p.RangeEnd), clientv3.PermissionType(clientv3.PermRead))
		}
		if err != nil {
			return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Failed copying grants from old role %v to new role %v.", oldValue, newValue)))
		}
	}
	_, err = cli.RoleDelete(ctx, fmt.Sprintf("%v", oldValue))
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("name", fmt.Sprintf("%v", newValue))
	if err != nil {
		return nil
	}
	d.SetId(uuidGenerator())
	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second

	cli := meta.(*clientv3.Client)

	name := d.Get("name").(string)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err := cli.RoleGet(ctx, name)
	cancel()
	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("The role %s doesn't exist", name)))
	}
	ctx, cancel = context.WithTimeout(context.Background(), requestTimeout)
	_, err = cli.RoleDelete(ctx, name)
	cancel()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
