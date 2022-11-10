package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func resourceKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeyCreate,
		ReadContext:   resourceKeyRead,
		UpdateContext: resourceKeyUpdate,
		DeleteContext: resourceKeyDelete,
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "Etcd key",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"value": {
				Description: "Etcd value",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceKeyCreate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second
	cli := meta.(*clientv3.Client)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	key := d.Get("key").(string)
	value := d.Get("value").(string)
	resp, err := cli.Get(ctx, key)

	tflog.Debug(ctx, fmt.Sprintf("cli.Get response: %s, kvs: %s, count: %v", resp.Kvs, resp.Kvs, resp.Count))

	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Error Calling Get() funcion from resourceKeyCreate for key: %s", key)))
	}

	_, putErr := cli.Put(ctx, key, value)
	if putErr != nil {
		return diag.FromErr(errors.Wrap(putErr, "Error writing key/value into etcd server"))
	}

	d.SetId(uuidGenerator())

	return resourceKeyRead(ctx, d, meta)
}

func resourceKeyRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second
	cli := meta.(*clientv3.Client)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	key := d.Get("key").(string)
	if key == "" {
		key = d.Id()
		d.Set("key", key)
	}
	resp, err := cli.Get(ctx, key)

	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("Failed calling cli.Get() from resourceKeyRead(). key=/%v/", key)))
	}

	if resp.Count == 0 {
		return diag.Errorf("Etcd returns no answer. It is suppose to have at least one empty value")
	}

	for _, ev := range resp.Kvs {
		tflog.Debug(ctx, fmt.Sprintf("here is the resp.kvs %v", resp.Kvs))
		if err := d.Set("value", string(ev.Value)); err != nil {
			return diag.FromErr(errors.Wrap(err, "Failed saving data into 'value'."))
		}
		err := d.Set("key", ev.Key)
		if err != nil {
			return nil
		}
	}

	d.SetId(uuidGenerator())

	return nil
}

func resourceKeyUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second
	if d.HasChange("value") {

		cli := meta.(*clientv3.Client)

		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()
		key := d.Get("key").(string)
		value := d.Get("value").(string)

		resp, err := cli.Get(ctx, key)

		if err != nil {
			return diag.FromErr(errors.Wrap(err, "Error Calling Get() function from resourceKeyUpdate."))
		}
		if resp.Count == 0 {
			return diag.Errorf("The Key already exists")
		}
		_, putErr := cli.Put(ctx, key, value)

		if putErr != nil {
			return diag.FromErr(errors.Wrap(putErr, "Error writing key/value in etcd server"))
		}

		return resourceKeyRead(ctx, d, meta)
	}
	return nil
}

func resourceKeyDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second
	cli := meta.(*clientv3.Client)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	key := d.Get("key").(string)

	_, errDelete := cli.Delete(ctx, key)

	if errDelete != nil {
		return diag.FromErr(errors.Wrap(errDelete, "Error Calling Delete() function from resourceKeyDelete."))
	}
	return nil
}
