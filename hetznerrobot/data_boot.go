package hetznerrobot

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataBoot() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBootRead,
		Description: "Provides details about a Hetzner Robot server boot configuration",
		Schema: map[string]*schema.Schema{
			"server_ip": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Server IP address",
			},
			// read-only / computed
			"active_profile": {
				Type:        schema.TypeString, // Enum should be better (linux/rescue/...)
				Computed:    true,
				Description: "Active boot profile",
			},
			"architecture": {
				Type:        schema.TypeString, // Enum should be better (amd64/...)
				Computed:    true,
				Description: "Active Architecture",
			},
			"ipv4_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server main IPv4 address",
			},
			"ipv6_network": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server main IPv6 net address",
			},
			"language": {
				Type:        schema.TypeString, // Enum should be better (amd64/...)
				Computed:    true,
				Description: "Language",
			},
			"operating_system": {
				Type:        schema.TypeString, // Enum should be better (ubuntu_20.04/...)
				Computed:    true,
				Description: "Active Operating System / Distribution",
			},
			"password": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current Rescue System root password / Linux installation password or null",
				Sensitive:   true,
			},
		},
		/*
			AuthorizedKeys []string		    authorized_key (Array)	Authorized public SSH keys
			HostKeys []string				host_key (Array)	Host keys
		*/
	}
}

func dataSourceBootRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, ok := meta.(HetznerRobotClient)
	if !ok {
		return diag.Errorf("Unable to cast meta to HetznerRobotClient")
	}

	serverIP, ok := d.Get("server_ip").(string)
	if !ok {
		return diag.Errorf("Unable to get server_ip as string")
	}
	boot, err := c.getBoot(ctx, serverIP)
	if err != nil {
		return diag.Errorf("Unable to find Boot Profile for server IP %s:\n\t %q", serverIP, err)
	}

	if err := d.Set("active_profile", boot.ActiveProfile); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("architecture", boot.Architecture); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ipv4_address", boot.ServerIPv4); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ipv6_network", boot.ServerIPv6); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("language", boot.Language); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("operating_system", boot.OperatingSystem); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("password", boot.Password); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(serverIP)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}
