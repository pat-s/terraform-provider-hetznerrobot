package hetznerrobot

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVSwitch() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVSwitchCreate,
		ReadContext:   resourceVSwitchRead,
		UpdateContext: resourceVSwitchUpdate,
		DeleteContext: resourceVSwitchDelete,
		Description:   "Manages vSwitch configuration for Hetzner Robot servers",

		Importer: &schema.ResourceImporter{
			StateContext: resourceVSwitchImportState,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "vSwitch name",
			},
			"vlan": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "VLAN ID",
			},
			// computed / read-only fields
			"is_canceled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Cancellation status",
			},
			"servers": {
				Type:        schema.TypeList,
				Description: "Attached server list",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"server_number": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"server_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"server_ipv6_net": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"subnets": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Attached subnet list",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:     schema.TypeString,
							Required: true,
						},
						"mask": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"gateway": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"cloud_networks": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Attached cloud network list",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Required: true,
						},
						"mask": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"gateway": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceVSwitchImportState(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	c, ok := meta.(HetznerRobotClient)
	if !ok {
		return nil, fmt.Errorf("unable to cast meta to HetznerRobotClient")
	}

	vSwitchID := d.Id()
	vSwitch, err := c.getVSwitch(ctx, vSwitchID)
	if err != nil {
		return nil, fmt.Errorf("unable to find VSwitch with ID %s: %w", vSwitchID, err)
	}

	_ = d.Set("name", vSwitch.Name)
	_ = d.Set("vlan", vSwitch.Vlan)
	_ = d.Set("is_canceled", vSwitch.Canceled)
	_ = d.Set("servers", vSwitch.Server)
	_ = d.Set("subnets", vSwitch.Subnet)
	_ = d.Set("cloud_networks", vSwitch.CloudNetwork)

	results := make([]*schema.ResourceData, 1)
	results[0] = d
	return results, nil
}

func resourceVSwitchCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, ok := meta.(HetznerRobotClient)
	if !ok {
		return diag.Errorf("Unable to cast meta to HetznerRobotClient")
	}

	name, _ := d.Get("name").(string)
	vlan, _ := d.Get("vlan").(int)
	vSwitch, err := c.createVSwitch(ctx, name, vlan)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to create VSwitch: %w", err))
	}

	_ = d.Set("is_canceled", vSwitch.Canceled)
	_ = d.Set("servers", vSwitch.Server)
	_ = d.Set("subnets", vSwitch.Subnet)
	_ = d.Set("cloud_networks", vSwitch.CloudNetwork)
	d.SetId(strconv.Itoa(vSwitch.ID))

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func resourceVSwitchRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, ok := meta.(HetznerRobotClient)
	if !ok {
		return diag.Errorf("Unable to cast meta to HetznerRobotClient")
	}

	vSwitchID := d.Id()
	vSwitch, err := c.getVSwitch(ctx, vSwitchID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to find VSwitch with ID %s: %w", vSwitchID, err))
	}

	_ = d.Set("name", vSwitch.Name)
	_ = d.Set("vlan", vSwitch.Vlan)
	_ = d.Set("canceled", vSwitch.Canceled)
	_ = d.Set("servers", vSwitch.Server)
	_ = d.Set("subnets", vSwitch.Subnet)
	_ = d.Set("cloud_networks", vSwitch.CloudNetwork)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func resourceVSwitchUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, ok := meta.(HetznerRobotClient)
	if !ok {
		return diag.Errorf("Unable to cast meta to HetznerRobotClient")
	}

	vSwitchID := d.Id()
	name, _ := d.Get("name").(string)
	vlan, _ := d.Get("vlan").(int)
	err := c.updateVSwitch(ctx, vSwitchID, name, vlan)
	if err != nil {
		return diag.Errorf("Unable to update VSwitch:\n\t %q", err)
	}

	if d.HasChange("servers") {
		o, n := d.GetChange("servers")

		oldServers, _ := o.([]any)
		newServers, _ := n.([]any)

		mb := make(map[int]struct{}, len(newServers))
		for _, x := range newServers {
			srv, ok := x.(map[string]any)
			if !ok {
				continue
			}
			if serverNum, ok := srv["server_number"].(int); ok {
				mb[serverNum] = struct{}{}
			}
		}
		var serversToRemove []HetznerRobotVSwitchServer
		for _, x := range oldServers {
			srv, ok := x.(map[string]any)
			if !ok {
				continue
			}
			srvNum, ok := srv["server_number"].(int)
			if !ok {
				continue
			}
			if _, found := mb[srvNum]; !found {
				serversToRemove = append(serversToRemove, HetznerRobotVSwitchServer{ServerNumber: srvNum})
			}
		}

		if err := c.removeVSwitchServers(ctx, vSwitchID, serversToRemove); err != nil {
			diag.Errorf("Unable to remove servers from VSwitch:\n\t %q", err)
		}

		ma := make(map[int]struct{}, len(oldServers))
		for _, x := range oldServers {
			srv, ok := x.(map[string]any)
			if !ok {
				continue
			}
			if serverNum, ok := srv["server_number"].(int); ok {
				ma[serverNum] = struct{}{}
			}
		}
		var serversToAdd []HetznerRobotVSwitchServer
		for _, x := range newServers {
			srv, ok := x.(map[string]any)
			if !ok {
				continue
			}
			srvNum, ok := srv["server_number"].(int)
			if !ok {
				continue
			}
			if _, found := ma[srvNum]; !found {
				serversToAdd = append(serversToAdd, HetznerRobotVSwitchServer{ServerNumber: srvNum})
			}
		}

		if err := c.addVSwitchServers(ctx, vSwitchID, serversToAdd); err != nil {
			diag.Errorf("Unable to add servers to VSwitch:\n\t %q", err)
		}
	}

	return resourceVSwitchRead(ctx, d, meta)
}

func resourceVSwitchDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, ok := meta.(HetznerRobotClient)
	if !ok {
		return diag.Errorf("Unable to cast meta to HetznerRobotClient")
	}

	vSwitchID := d.Id()
	err := c.deleteVSwitch(ctx, vSwitchID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to find VSwitch with ID %s: %w", vSwitchID, err))
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}
