package hetznerrobot

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceFirewall() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFirewallCreate,
		ReadContext:   resourceFirewallRead,
		UpdateContext: resourceFirewallUpdate,
		DeleteContext: resourceFirewallDelete,
		Description:   "Manages firewall configuration for a Hetzner Robot server",
		Importer: &schema.ResourceImporter{
			StateContext: resourceFirewallImportState,
		},
		Schema: map[string]*schema.Schema{
			"server_ip": {
				Type:     schema.TypeString,
				Required: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"whitelist_hos": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dst_ip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dst_port": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"src_ip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"src_port": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tcp_flags": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"action": {
							Type: schema.TypeString,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
								"accept",
								"discard",
							}, false)),
							Required: true,
						},
						"ip_version": {
							Type: schema.TypeString,
							Optional: true,
							Default: "ipv4",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
								"ipv4",
								"ipv6",
							}, false)),
						},
					},
				},
			},
		},
	}
}

func resourceFirewallImportState(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
	c := m.(HetznerRobotClient)

	firewallID := d.Id()

	firewall, err := c.getFirewall(ctx, firewallID)
	if err != nil {
		return nil, fmt.Errorf("could not find firewall with ID %s: %s", firewallID, err)
	}

	active := false
	if firewall.Status == "active" {
		active = true
	}

	rules := make([]map[string]any, 0)
	for _, rule := range firewall.Rules.Input {
		r := map[string]any{
			"name":       rule.Name,
			"src_ip":     rule.SrcIP,
			"src_port":   rule.SrcPort,
			"dst_ip":     rule.DstIP,
			"dst_port":   rule.DstPort,
			"protocol":   rule.Protocol,
			"tcp_flags":  rule.TCPFlags,
			"action":     rule.Action,
			"ip_version": rule.IPVersion,
		}
		rules = append(rules, r)
	}

	d.Set("active", active)
	d.Set("rule", rules)
	d.Set("server_ip", firewall.IP)
	d.Set("whitelist_hos", firewall.WhitelistHetznerServices)
	d.SetId(firewall.IP)

	results := make([]*schema.ResourceData, 1)
	results[0] = d
	return results, nil
}

func resourceFirewallCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(HetznerRobotClient)

	serverIP := d.Get("server_ip").(string)

	status := "disabled"
	if d.Get("active").(bool) {
		status = "active"
	}

	rules := make([]HetznerRobotFirewallRule, 0)
	for _, ruleMap := range d.Get("rule").([]any) {
		ruleProperties := ruleMap.(map[string]any)
		ipVersion := "ipv4"
		if v, ok := ruleProperties["ip_version"].(string); ok && v != "" {
			ipVersion = v
		}
		rules = append(rules, HetznerRobotFirewallRule{
			Name:      ruleProperties["name"].(string),
			SrcIP:     ruleProperties["src_ip"].(string),
			SrcPort:   ruleProperties["src_port"].(string),
			DstIP:     ruleProperties["dst_ip"].(string),
			DstPort:   ruleProperties["dst_port"].(string),
			Protocol:  ruleProperties["protocol"].(string),
			TCPFlags:  ruleProperties["tcp_flags"].(string),
			Action:    ruleProperties["action"].(string),
			IPVersion: ipVersion,
		})
	}

	if err := c.setFirewall(ctx, HetznerRobotFirewall{
		IP:                       serverIP,
		WhitelistHetznerServices: d.Get("whitelist_hos").(bool),
		Status:                   status,
		Rules:                    HetznerRobotFirewallRules{Input: rules},
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(serverIP)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func resourceFirewallRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(HetznerRobotClient)

	serverIP := d.Id()

	firewall, err := c.getFirewall(ctx, serverIP)
	if err != nil {
		return diag.FromErr(err)
	}

	active := false
	if firewall.Status == "active" {
		active = true
	}

	rules := make([]map[string]any, 0)
	for _, rule := range firewall.Rules.Input {
		r := map[string]any{
			"name":       rule.Name,
			"src_ip":     rule.SrcIP,
			"src_port":   rule.SrcPort,
			"dst_ip":     rule.DstIP,
			"dst_port":   rule.DstPort,
			"protocol":   rule.Protocol,
			"tcp_flags":  rule.TCPFlags,
			"action":     rule.Action,
			"ip_version": rule.IPVersion,
		}
		rules = append(rules, r)
	}
	d.Set("active", active)
	d.Set("rule", rules)
	d.Set("server_ip", firewall.IP)
	d.Set("whitelist_hos", firewall.WhitelistHetznerServices)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func resourceFirewallUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c := m.(HetznerRobotClient)

	serverIP := d.Get("server_ip").(string)

	status := "disabled"
	if d.Get("active").(bool) {
		status = "active"
	}

	rules := make([]HetznerRobotFirewallRule, 0)
	for _, ruleMap := range d.Get("rule").([]any) {
		ruleProperties := ruleMap.(map[string]any)
		ipVersion := "ipv4"
		if v, ok := ruleProperties["ip_version"].(string); ok && v != "" {
			ipVersion = v
		}
		rules = append(rules, HetznerRobotFirewallRule{
			Name:      ruleProperties["name"].(string),
			SrcIP:     ruleProperties["src_ip"].(string),
			SrcPort:   ruleProperties["src_port"].(string),
			DstIP:     ruleProperties["dst_ip"].(string),
			DstPort:   ruleProperties["dst_port"].(string),
			Protocol:  ruleProperties["protocol"].(string),
			TCPFlags:  ruleProperties["tcp_flags"].(string),
			Action:    ruleProperties["action"].(string),
			IPVersion: ipVersion,
		})
	}

	if err := c.setFirewall(ctx, HetznerRobotFirewall{
		IP:                       serverIP,
		WhitelistHetznerServices: d.Get("whitelist_hos").(bool),
		Status:                   status,
		Rules:                    HetznerRobotFirewallRules{Input: rules},
	}); err != nil {
		return diag.FromErr(err)
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func resourceFirewallDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}
