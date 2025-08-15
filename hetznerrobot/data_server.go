package hetznerrobot

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataServer() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServerRead,
		Description: "Provides details about a Hetzner Robot server",
		Schema: map[string]*schema.Schema{
			"server_number": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Server number",
			},
			"server_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server name",
			},
			"server_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server IP",
			},
			"server_ipv6": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server IPv6 Net",
			},
			"datacenter": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data center",
			},
			"is_canceled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Status of server cancellation",
			},
			"paid_until": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Paid until date",
			},
			"product": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server product name",
			},
			"ip_addresses": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Array of assigned single IP addresses",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"server_subnets": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Array of assigned subnets",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:     schema.TypeString,
							Required: true,
						},
						"mask": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server status (\"ready\" or \"in process\")",
			},
			"traffic": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Free traffic quota, 'unlimited' in case of unlimited traffic",
			},
			"linked_storagebox": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Linked Storage Box ID",
			},
			"reset": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag of reset system availability",
			},
			"rescue": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag of Rescue System availability",
			},
			"vnc": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag of VNC installation availability",
			},
			"windows": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag of Windows installation availability",
			},
			"plesk": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag of Plesk installation availability",
			},
			"cpanel": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag of cPanel installation availability",
			},
			"wol": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag of Wake On Lan availability",
			},
			"hot_swap": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag of Hot Swap availability",
			},
		},
	}
}

func dataSourceServerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, ok := meta.(HetznerRobotClient)
	if !ok {
		return diag.Errorf("Unable to cast meta to HetznerRobotClient")
	}

	serverNumber, ok := d.Get("server_number").(int)
	if !ok {
		return diag.Errorf("Unable to get server_number as int")
	}

	server, err := c.getServer(ctx, serverNumber)
	if err != nil {
		return diag.Errorf("Unable to find Server with number %d:\n\t %q", serverNumber, err)
	}
	if err := d.Set("datacenter", server.DataCenter); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_canceled", server.Canceled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("paid_until", server.PaidUntil); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("product", server.Product); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("server_ip_addresses", server.IPs); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("server_ip", server.ServerIP); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("server_ip_v6_net", server.ServerIPv6); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("server_name", server.ServerName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("server_subnets", server.Subnets); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", server.Status); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("traffic", server.Traffic); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.Itoa(server.ServerNumber))

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}
