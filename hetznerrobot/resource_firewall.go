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
							Type:     schema.TypeString,
							Optional: true,
							Default:  "ipv4",
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
	c, ok := m.(HetznerRobotClient)
	if !ok {
		return nil, fmt.Errorf("unable to cast meta to HetznerRobotClient")
	}

	firewallID := d.Id()

	firewall, err := c.getFirewall(ctx, firewallID)
	if err != nil {
		return nil, fmt.Errorf("could not find firewall with ID %s: %w", firewallID, err)
	}

	active := firewall.Status == "active"

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

	_ = d.Set("active", active)
	_ = d.Set("rule", rules)
	_ = d.Set("server_ip", firewall.IP)
	_ = d.Set("whitelist_hos", firewall.WhitelistHetznerServices)
	d.SetId(firewall.IP)

	results := make([]*schema.ResourceData, 1)
	results[0] = d
	return results, nil
}

func resourceFirewallCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c, ok := m.(HetznerRobotClient)
	if !ok {
		return diag.Errorf("Unable to cast meta to HetznerRobotClient")
	}

	serverIP, _ := d.Get("server_ip").(string)

	status := "disabled"
	if active, _ := d.Get("active").(bool); active {
		status = "active"
	}

	var diags diag.Diagnostics
	rules := make([]HetznerRobotFirewallRule, 0)
	rules_data, _ := d.Get("rule").([]any)
	for _, ruleMap := range rules_data {
		ruleProperties, ok := ruleMap.(map[string]any)
		if !ok {
			continue
		}
		ipVersion := "ipv4"
		if v, ok := ruleProperties["ip_version"].(string); ok && v != "" {
			ipVersion = v
		}
		name, _ := ruleProperties["name"].(string)
		srcIP, _ := ruleProperties["src_ip"].(string)
		srcPort, _ := ruleProperties["src_port"].(string)
		dstIP, _ := ruleProperties["dst_ip"].(string)
		dstPort, _ := ruleProperties["dst_port"].(string)
		protocol, _ := ruleProperties["protocol"].(string)
		tcpFlags, _ := ruleProperties["tcp_flags"].(string)
		action, _ := ruleProperties["action"].(string)

		// Warn about IPv6 restrictions
		if ipVersion == "ipv6" {
			if srcIP != "" {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  fmt.Sprintf("IPv6 rule '%s': src_ip field ignored", name),
					Detail:   "Hetzner Robot API does not support source IP filtering for IPv6 rules. The src_ip field will be ignored.",
				})
			}
			if dstIP != "" {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  fmt.Sprintf("IPv6 rule '%s': dst_ip field ignored", name),
					Detail:   "Hetzner Robot API does not support destination IP filtering for IPv6 rules. The dst_ip field will be ignored.",
				})
			}
		}

		rules = append(rules, HetznerRobotFirewallRule{
			Name:      name,
			SrcIP:     srcIP,
			SrcPort:   srcPort,
			DstIP:     dstIP,
			DstPort:   dstPort,
			Protocol:  protocol,
			TCPFlags:  tcpFlags,
			Action:    action,
			IPVersion: ipVersion,
		})
	}

	if err := c.setFirewall(ctx, HetznerRobotFirewall{
		IP:                       serverIP,
		WhitelistHetznerServices: func() bool { val, _ := d.Get("whitelist_hos").(bool); return val }(),
		Status:                   status,
		Rules:                    HetznerRobotFirewallRules{Input: rules},
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(serverIP)

	return diags
}

func resourceFirewallRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c, ok := m.(HetznerRobotClient)
	if !ok {
		return diag.Errorf("Unable to cast meta to HetznerRobotClient")
	}

	serverIP := d.Id()

	firewall, err := c.getFirewall(ctx, serverIP)
	if err != nil {
		return diag.FromErr(err)
	}

	active := firewall.Status == "active"

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
	_ = d.Set("active", active)
	_ = d.Set("rule", rules)
	_ = d.Set("server_ip", firewall.IP)
	_ = d.Set("whitelist_hos", firewall.WhitelistHetznerServices)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func resourceFirewallUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	c, ok := m.(HetznerRobotClient)
	if !ok {
		return diag.Errorf("Unable to cast meta to HetznerRobotClient")
	}

	serverIP, _ := d.Get("server_ip").(string)

	status := "disabled"
	if active, _ := d.Get("active").(bool); active {
		status = "active"
	}

	var diags diag.Diagnostics
	rules := make([]HetznerRobotFirewallRule, 0)
	rules_data, _ := d.Get("rule").([]any)
	for _, ruleMap := range rules_data {
		ruleProperties, ok := ruleMap.(map[string]any)
		if !ok {
			continue
		}
		ipVersion := "ipv4"
		if v, ok := ruleProperties["ip_version"].(string); ok && v != "" {
			ipVersion = v
		}
		name, _ := ruleProperties["name"].(string)
		srcIP, _ := ruleProperties["src_ip"].(string)
		srcPort, _ := ruleProperties["src_port"].(string)
		dstIP, _ := ruleProperties["dst_ip"].(string)
		dstPort, _ := ruleProperties["dst_port"].(string)
		protocol, _ := ruleProperties["protocol"].(string)
		tcpFlags, _ := ruleProperties["tcp_flags"].(string)
		action, _ := ruleProperties["action"].(string)

		// Warn about IPv6 restrictions
		if ipVersion == "ipv6" {
			if srcIP != "" {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  fmt.Sprintf("IPv6 rule '%s': src_ip field ignored", name),
					Detail:   "Hetzner Robot API does not support source IP filtering for IPv6 rules. The src_ip field will be ignored.",
				})
			}
			if dstIP != "" {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  fmt.Sprintf("IPv6 rule '%s': dst_ip field ignored", name),
					Detail:   "Hetzner Robot API does not support destination IP filtering for IPv6 rules. The dst_ip field will be ignored.",
				})
			}
		}

		rules = append(rules, HetznerRobotFirewallRule{
			Name:      name,
			SrcIP:     srcIP,
			SrcPort:   srcPort,
			DstIP:     dstIP,
			DstPort:   dstPort,
			Protocol:  protocol,
			TCPFlags:  tcpFlags,
			Action:    action,
			IPVersion: ipVersion,
		})
	}

	if err := c.setFirewall(ctx, HetznerRobotFirewall{
		IP:                       serverIP,
		WhitelistHetznerServices: func() bool { val, _ := d.Get("whitelist_hos").(bool); return val }(),
		Status:                   status,
		Rules:                    HetznerRobotFirewallRules{Input: rules},
	}); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceFirewallDelete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}
