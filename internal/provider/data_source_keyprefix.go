package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Etcd_key struct {
	Key   string
	Value string
}

func dataSourceKeyPrefix() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeyPrefixRead,
		Schema: map[string]*schema.Schema{
			"prefix": {
				Type:     schema.TypeString,
				Required: true,
			},
			"entries": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"last_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceKeyPrefixRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var requestTimeout = 5 * time.Second

	cli := m.(*clientv3.Client)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	prefix := fmt.Sprintf("%v", d.Get("prefix"))
	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return append(diag.FromErr(err), diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error reading data from etcd",
			Detail:   "Failed calling cli.Get() from dataSourceKeyRead()",
		})
	}
	if resp.Count == 0 {
		return append(diag.FromErr(err), diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error reading data from etcd",
			Detail:   "Etcd returns no answer. It is suppose to have at least one empty value.",
		})
	}

	entries := make([]interface{}, len(resp.Kvs), len(resp.Kvs))

	for i, ev := range resp.Kvs {
		if ev == nil {
			return append(diag.FromErr(err), diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Error reading data from etcd",
				Detail:   "Etcd prefix returns no answer. It is suppose to have at least one empty value.",
			})
		}

		entry := make(map[string]interface{})

		entry["key"] = string(ev.Key)
		entry["value"] = string(ev.Value)

		entries[i] = entry
	}

	if err := d.Set("entries", entries); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(uuidGenerator())

	return diags
}
