package hetznerrobot

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HETZNERROBOT_USERNAME", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HETZNERROBOT_PASSWORD", nil),
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HETZNERROBOT_URL", "https://robot-ws.your-server.de"),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"hetznerrobot_boot":     resourceBoot(),
			"hetznerrobot_firewall": resourceFirewall(),
			"hetznerrobot_vswitch":  resourceVSwitch(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hetznerrobot_boot":    dataBoot(),
			"hetznerrobot_server":  dataServer(),
			"hetznerrobot_vswitch": dataVSwitch(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	username, ok := d.Get("username").(string)
	if !ok {
		return nil, diag.Errorf("username must be a string")
	}

	password, ok := d.Get("password").(string)
	if !ok {
		return nil, diag.Errorf("password must be a string")
	}

	url, ok := d.Get("url").(string)
	if !ok {
		return nil, diag.Errorf("url must be a string")
	}

	if username == "" {
		return nil, diag.Errorf("username is required for Hetzner Robot authentication")
	}
	if password == "" {
		return nil, diag.Errorf("password is required for Hetzner Robot authentication")
	}

	client := NewHetznerRobotClient(username, password, url)

	var diags diag.Diagnostics
	return client, diags
}
