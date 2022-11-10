package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var requestTimeout = 10 * time.Second

func resourceGrantRoleUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGrantRoleUserCreate,
		ReadContext:   resourceGrantRoleUserRead,
		UpdateContext: resourceGrantRoleUserUpdate,
		DeleteContext: resourceGrantRoleUserRevoke,
		Schema: map[string]*schema.Schema{
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceGrantRoleUserCreate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cli := meta.(*clientv3.Client)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	user := d.Get("user_name").(string)
	role := d.Get("role").(string)

	tflog.Info(ctx, "some test message for Create function")

	_, errUserGet := cli.UserGet(ctx, user)

	if errUserGet != nil {
		return diag.FromErr(errors.Wrap(errUserGet, fmt.Sprintf("The user %s doesn't exist, please create it first.", user)))
	}

	_, errRoleGet := cli.RoleGet(ctx, role)

	if errRoleGet != nil {
		return diag.FromErr(errors.Wrap(errRoleGet, fmt.Sprintf("The role %s doesn't exist, please create it first.", role)))
	}

	_, err := cli.UserGrantRole(ctx, user, role)

	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Failed granting role %s to user %s", role, user)))
	}

	d.SetId(uuidGenerator())

	return resourceGrantRoleUserRead(ctx, d, meta)
}

func resourceGrantRoleUserRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cli := meta.(*clientv3.Client)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	user := d.Get("user_name").(string)
	role := d.Get("role").(string)

	respRoleGet, err := cli.RoleGet(ctx, role)
	if err != nil {
		tflog.Error(ctx, "respRoleGet", map[string]interface{}{
			"header": respRoleGet.Header,
			"perm":   respRoleGet.Perm,
		})
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Failed getting role %s", role)))
	}

	if err := d.Set("user_name", user); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error occurred with user_name setting"))
	}
	d.SetId(uuidGenerator())
	return nil
}

func resourceGrantRoleUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cli := meta.(*clientv3.Client)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	tflog.Debug(ctx, "some test message for Update function")

	user := d.Get("user_name").(string)
	role := d.Get("role").(string)

	_, errUserGet := cli.UserGet(ctx, user)
	defer cancel()
	if errUserGet != nil {
		return diag.FromErr(errors.Wrap(errUserGet, fmt.Sprintf("The user %s doesn't exist, please create it first.", user)))
	}

	_, errRoleGet := cli.RoleGet(ctx, role)
	if errRoleGet != nil {
		return diag.FromErr(errors.Wrap(errUserGet, fmt.Sprintf("The role %s doesn't exist, please create it first.", role)))
	}

	_, err := cli.UserGrantRole(ctx, user, role)

	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Failed granting role: %v to user: %v", role, user)))
	}

	d.SetId(uuidGenerator())

	return resourceGrantRoleUserRead(ctx, d, meta)
}

func resourceGrantRoleUserRevoke(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clientv3.Client)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	tflog.Info(ctx, "some test message for Revoke function")

	user := d.Get("user_name").(string)
	role := d.Get("role").(string)

	_, errRevokeRole := cli.UserRevokeRole(ctx, user, role)
	defer cancel()
	if errRevokeRole != nil {
		return diag.FromErr(errors.Wrap(errRevokeRole, fmt.Sprintf("Failed revoke role: %v from user: %v", role, user)))
	}

	return nil
}
