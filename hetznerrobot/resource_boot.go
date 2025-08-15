package hetznerrobot

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBoot() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBootCreate,
		ReadContext:   resourceBootRead,
		UpdateContext: resourceBootUpdate,
		DeleteContext: resourceBootDelete,
		Description:   "Manages boot configuration for a Hetzner Robot server",

		Importer: &schema.ResourceImporter{
			StateContext: resourceBootImportState,
		},

		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Server ID",
			},
			// optional
			"active_profile": {
				Type:        schema.TypeString, // Enum should be better (linux/rescue/...)
				Optional:    true,
				Description: "Active boot profile",
			},
			"architecture": {
				Type:        schema.TypeString, // Enum should be better (amd64/...)
				Optional:    true,
				Description: "Active Architecture",
			},
			"language": {
				Type:        schema.TypeString, // Enum should be better (amd64/...)
				Optional:    true,
				Description: "Language",
			},
			"operating_system": {
				Type:        schema.TypeString, // Enum should be better (ubuntu_20.04/...)
				Optional:    true,
				Description: "Active Operating System / Distribution",
			},
			"authorized_keys": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "One or more SSH key fingerprints",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// read-only / computed
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
			"password": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current Rescue System root password / Linux installation password or null",
				Sensitive:   true,
			},
		},
	}
}

func resourceBootImportState(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	c, ok := meta.(HetznerRobotClient)
	if !ok {
		return nil, fmt.Errorf("unable to cast meta to HetznerRobotClient")
	}

	serverID := d.Id()

	boot, err := c.getBoot(ctx, serverID)
	if err != nil {
		return nil, err
	}

	_ = d.Set("active_profile", boot.ActiveProfile)
	_ = d.Set("architecture", boot.Architecture)
	_ = d.Set("ipv4_address", boot.ServerIPv4)
	_ = d.Set("ipv6_network", boot.ServerIPv6)
	_ = d.Set("language", boot.Language)
	_ = d.Set("operating_system", boot.OperatingSystem)
	_ = d.Set("password", boot.Password)
	_ = d.Set("server_id", serverID)

	results := make([]*schema.ResourceData, 1)
	results[0] = d
	return results, nil
}

func resourceBootCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, ok := meta.(HetznerRobotClient)
	if !ok {
		return diag.Errorf("Unable to cast meta to HetznerRobotClient")
	}

	serverID := d.Id()
	activeBootProfile, _ := d.Get("active_profile").(string)
	arch, _ := d.Get("architecture").(string)
	os, _ := d.Get("operating_system").(string)
	lang, _ := d.Get("language").(string)
	authorizedKeys := make([]string, 0)
	if input := d.Get("authorized_keys"); input != nil {
		keys, ok := input.([]any)
		if !ok {
			keys = []any{}
		}
		for _, key := range keys {
			keyStr, ok := key.(string)
			if !ok {
				continue
			}
			authorizedKeys = append(authorizedKeys, keyStr)
		}
	}

	bootProfile, err := c.setBootProfile(ctx, serverID, activeBootProfile, arch, os, lang, authorizedKeys)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("ipv4_address", bootProfile.ServerIPv4)
	_ = d.Set("ipv6_network", bootProfile.ServerIPv6)
	_ = d.Set("password", bootProfile.Password)
	d.SetId(serverID)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func resourceBootRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, ok := meta.(HetznerRobotClient)
	if !ok {
		return diag.Errorf("Unable to cast meta to HetznerRobotClient")
	}

	serverID := d.Id()
	boot, err := c.getBoot(ctx, serverID)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("active_profile", boot.ActiveProfile)
	_ = d.Set("architecture", boot.Architecture)
	_ = d.Set("ipv4_address", boot.ServerIPv4)
	_ = d.Set("ipv6_network", boot.ServerIPv6)
	_ = d.Set("language", boot.Language)
	_ = d.Set("operating_system", boot.OperatingSystem)
	_ = d.Set("password", boot.Password)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func resourceBootUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, ok := meta.(HetznerRobotClient)
	if !ok {
		return diag.Errorf("Unable to cast meta to HetznerRobotClient")
	}

	serverID := d.Id()
	activeBootProfile, _ := d.Get("active_profile").(string)
	arch, _ := d.Get("architecture").(string)
	os, _ := d.Get("operating_system").(string)
	lang, _ := d.Get("language").(string)
	authorizedKeys := make([]string, 0)
	if input := d.Get("authorized_keys"); input != nil {
		keys, ok := input.([]any)
		if !ok {
			keys = []any{}
		}
		for _, key := range keys {
			keyStr, ok := key.(string)
			if !ok {
				continue
			}
			authorizedKeys = append(authorizedKeys, keyStr)
		}
	}

	bootProfile, err := c.setBootProfile(ctx, serverID, activeBootProfile, arch, os, lang, authorizedKeys)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("ipv4_address", bootProfile.ServerIPv4)
	_ = d.Set("ipv6_network", bootProfile.ServerIPv6)
	_ = d.Set("password", bootProfile.Password)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func resourceBootDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}
