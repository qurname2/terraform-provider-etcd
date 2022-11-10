package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/pkg/errors"
	"math/rand"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	lowerCharSet   = "abcdefghijklmnopqrstuvwxyz"
	upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialCharSet = "!@#$%&*+-_?.,"
	numberSet      = "0123456789"
	allCharSet     = lowerCharSet + upperCharSet + specialCharSet + numberSet
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func generatePassword(passwdLen, nSpecialChar, nNum, nUpperCase int) string {
	var password strings.Builder

	//Set special character
	for i := 0; i < nSpecialChar; i++ {
		random := rand.Intn(len(specialCharSet))
		password.WriteString(string(specialCharSet[random]))
	}

	//Set numeric
	for i := 0; i < nNum; i++ {
		random := rand.Intn(len(numberSet))
		password.WriteString(string(numberSet[random]))
	}

	//Set uppercase
	for i := 0; i < nUpperCase; i++ {
		random := rand.Intn(len(upperCharSet))
		password.WriteString(string(upperCharSet[random]))
	}

	remainingLength := passwdLen - nSpecialChar - nNum - nUpperCase
	for i := 0; i < remainingLength; i++ {
		random := rand.Intn(len(allCharSet))
		password.WriteString(string(allCharSet[random]))
	}
	inRune := []rune(password.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})
	return string(inRune)
}

func resourceUserCreate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second

	cli := meta.(*clientv3.Client)

	name := d.Get("name").(string)
	password := d.Get("password").(string)
	if password == "" {
		password = generatePassword(24, 3, 3, 3)
		d.Set("password", password)
	}
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	_, err := cli.UserAdd(ctx, name, password)
	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("A problem occurred with user creation %s", name)))
	}

	d.SetId(uuidGenerator())

	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second

	cli := meta.(*clientv3.Client)

	name := d.Get("name").(string)
	if name == "" {
		name = d.Id()
		d.Set("name", name)
	}
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err := cli.UserGet(ctx, name)
	defer cancel()
	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("The user %v doesn't exist. Maybe someone removed that manually.", name)))
	}
	d.SetId(uuidGenerator())

	return nil
}

func resourceUserUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second
	var name string
	cli := meta.(*clientv3.Client)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	oldValueName, newValueName := d.GetChange("name")
	_, newValuePassword := d.GetChange("password")

	tflog.Debug(ctx, fmt.Sprintf("oldvalue is: %v, newValue is: %v", oldValueName, newValueName))

	if oldValueName == newValueName {
		name = fmt.Sprintf("%s", oldValueName)
		_, err := cli.UserChangePassword(ctx, name, fmt.Sprintf("%s", newValuePassword))
		if err != nil {
			return diag.FromErr(errors.Wrap(err, fmt.Sprintf("A problem occurred with password changing for user: %s", name)))
		}
	} else {
		name = fmt.Sprintf("%s", newValueName)
		tflog.Debug(ctx, fmt.Sprintf("Going to remove user: %s", oldValueName))
		_, errUserDelete := cli.UserDelete(ctx, fmt.Sprintf("%s", oldValueName))
		if errUserDelete != nil {
			return diag.FromErr(errors.Wrap(errUserDelete, fmt.Sprintf("A problem occurred with user deletion %s", oldValueName)))
		}
		_, errUserAdd := cli.UserAdd(ctx, fmt.Sprintf("%s", newValueName), fmt.Sprintf("%s", newValuePassword))
		if errUserAdd != nil {
			return diag.FromErr(errors.Wrap(errUserAdd, fmt.Sprintf("A problem occurred with user update %s", newValueName)))
		}
	}

	d.SetId(uuidGenerator())
	return resourceUserRead(ctx, d, meta)
}

func resourceUserDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var requestTimeout = 5 * time.Second

	cli := meta.(*clientv3.Client)
	name := d.Get("name").(string)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err := cli.UserDelete(ctx, name)
	defer cancel()

	tflog.Info(ctx, fmt.Sprintf("Going to remove user: %s", name))

	if err != nil {
		return diag.FromErr(errors.Wrap(err, fmt.Sprintf("A problem occurred with user deletion %s", name)))

	}

	return nil
}
