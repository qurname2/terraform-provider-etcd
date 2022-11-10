package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func resourcePermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePermissionCreate,
		ReadContext:   resourcePermissionRead,
		UpdateContext: resourcePermissionUpdate,
		DeleteContext: resourcePermissionDelete,
		Schema: map[string]*schema.Schema{
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"withprefix": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"endrange": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"permission": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if !contains([]string{"READ", "READWRITE"}, v) {
						errs = append(errs, fmt.Errorf("%q must be READ or READWRITE, got: %v", key, v))
					}
					return
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourcePermissionCreate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second
	var rangeEnd string
	var permissionType clientv3.PermissionType

	cli := meta.(*clientv3.Client)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	role := d.Get("role").(string)
	key := d.Get("key").(string)
	withPrefix := d.Get("withprefix").(bool)

	if withPrefix == true {
		rangeEnd = clientv3.GetPrefixRangeEnd(key)
	} else {
		rangeEnd = d.Get("endrange").(string)
		if rangeEnd == "" {
			return diag.Errorf("'endrange' is a mandatory argument when you define 'withprefix' == false")
		}
	}
	if d.Get("permission").(string) == "READWRITE" {
		permissionType = clientv3.PermissionType(clientv3.PermReadWrite)
	} else {
		permissionType = clientv3.PermissionType(clientv3.PermRead)
	}
	_, err := cli.RoleGrantPermission(ctx, role, key, rangeEnd, permissionType)
	//cancel()
	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Failed creating permission: %v to key: %v into role: %v", permissionType, key, role)))
	}

	d.SetId(uuidGenerator())

	return resourcePermissionRead(ctx, d, meta)
}

func resourcePermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second

	cli := meta.(*clientv3.Client)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

	role := d.Get("role").(string)
	key := d.Get("key").(string)
	resp, err := cli.RoleGet(ctx, role)
	cancel()
	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Failed getting role: %v", role)))
	}
	for _, p := range resp.Perm {
		if string(p.Key) != key {
			continue
		}
		if string(p.RangeEnd) == clientv3.GetPrefixRangeEnd(key) {
			d.Set("withPrefix", true)
		} else {
			d.Set("withPrefix", false)
		}
		d.Set("permission", fmt.Sprintf("%v", p.PermType))
		d.SetId(uuidGenerator())
	}

	return nil
}

func resourcePermissionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second
	var rangeEnd string
	var permission clientv3.PermissionType

	cli := meta.(*clientv3.Client)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	role := d.Get("role").(string)
	key := d.Get("key").(string)
	withPrefix := d.Get("withprefix").(bool)
	if withPrefix == true {
		rangeEnd = clientv3.GetPrefixRangeEnd(key)
	} else {
		rangeEnd = d.Get("endrange").(string)
		if rangeEnd == "" {
			return diag.Errorf("endrange is a mandatory argument when you define 'withprefix' == false.")
		}
	}
	if d.Get("role").(string) == "READWRITE" {
		permission = clientv3.PermissionType(clientv3.PermReadWrite)
	} else {
		permission = clientv3.PermissionType(clientv3.PermRead)
	}
	_, err := cli.RoleGrantPermission(ctx, role, key, rangeEnd, permission)
	cancel()
	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Failed creating permission: %v to key: %v into role: %v", permission, key, role)))
	}

	d.SetId(uuidGenerator())

	return resourcePermissionRead(ctx, d, meta)
}

func resourcePermissionDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second

	cli := meta.(*clientv3.Client)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	role := d.Get("role").(string)
	key := d.Get("key").(string)
	rangeEnd := d.Get("endrange").(string)

	resp, err := cli.RoleGet(ctx, role)
	//cancel()
	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Failed getting role: %v", role)))
	}
	for _, p := range resp.Perm {
		if string(p.Key) != key || string(p.RangeEnd) != rangeEnd {
			continue
		}
		_, err = cli.RoleRevokePermission(ctx, role, key, rangeEnd)
		if err != nil {
			return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Failed revoking permission to key: %v from role: %v", key, role)))
		}
	}
	return nil
}
