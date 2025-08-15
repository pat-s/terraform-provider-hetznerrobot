package hetznerrobot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type HetznerRobotFirewallResponse struct {
	Firewall HetznerRobotFirewall `json:"firewall"`
}

type HetznerRobotFirewall struct {
	IP                       string                    `json:"server_ip"`
	WhitelistHetznerServices bool                      `json:"whitelist_hos"`
	Status                   string                    `json:"status"`
	Rules                    HetznerRobotFirewallRules `json:"rules"`
}

type HetznerRobotFirewallRules struct {
	Input []HetznerRobotFirewallRule `json:"input"`
}

type HetznerRobotFirewallRule struct {
	Name      string `json:"name"`
	DstIP     string `json:"dst_ip"`
	DstPort   string `json:"dst_port"`
	SrcIP     string `json:"src_ip"`
	SrcPort   string `json:"src_port"`
	Protocol  string `json:"protocol"`
	TCPFlags  string `json:"tcp_flags"`
	Action    string `json:"action"`
	IPVersion string `json:"ip_version"`
}

func (c *HetznerRobotClient) getFirewall(ctx context.Context, ip string) (*HetznerRobotFirewall, error) {
	bytes, err := c.makeAPICall(ctx, "GET", fmt.Sprintf("%s/firewall/%s", c.url, ip), nil, []int{http.StatusOK, http.StatusAccepted})
	if err != nil {
		return nil, err
	}

	firewall := HetznerRobotFirewallResponse{}
	if err = json.Unmarshal(bytes, &firewall); err != nil {
		return nil, err
	}
	return &firewall.Firewall, nil
}

func (c *HetznerRobotClient) setFirewall(ctx context.Context, firewall HetznerRobotFirewall) error {
	data := url.Values{}

	whitelistHOS := "false"
	if firewall.WhitelistHetznerServices {
		whitelistHOS = "true"
	}

	data.Set("whitelist_hos", whitelistHOS)
	data.Set("status", firewall.Status)

	// Process all rules using the working format
	for idx, rule := range firewall.Rules.Input {
		ipVersion := rule.IPVersion
		if ipVersion == "" {
			ipVersion = "ipv4"
		}

		// Basic fields that are always set
		data.Set(fmt.Sprintf("rules[input][%d][name]", idx), rule.Name)
		data.Set(fmt.Sprintf("rules[input][%d][ip_version]", idx), ipVersion)
		data.Set(fmt.Sprintf("rules[input][%d][action]", idx), rule.Action)

		// For IPv6 rules, src_ip and dst_ip CANNOT be set according to API restrictions
		if ipVersion != "ipv6" {
			// Only set IP addresses for IPv4 rules
			data.Set(fmt.Sprintf("rules[input][%d][src_ip]", idx), rule.SrcIP)
			if rule.DstIP != "" {
				data.Set(fmt.Sprintf("rules[input][%d][dst_ip]", idx), rule.DstIP)
			}
		}

		// Port fields can be set for both IPv4 and IPv6
		data.Set(fmt.Sprintf("rules[input][%d][dst_port]", idx), rule.DstPort)
		if rule.SrcPort != "" {
			data.Set(fmt.Sprintf("rules[input][%d][src_port]", idx), rule.SrcPort)
		}

		// Protocol and TCP flags
		if rule.Protocol != "" {
			data.Set(fmt.Sprintf("rules[input][%d][protocol]", idx), rule.Protocol)
		}
		if rule.TCPFlags != "" {
			data.Set(fmt.Sprintf("rules[input][%d][tcp_flags]", idx), rule.TCPFlags)
		}
	}

	// Add default output rule - required by API
	data.Set("rules[output][0][name]", "Allow all")
	data.Set("rules[output][0][action]", "accept")

	_, err := c.makeAPICall(ctx, "POST", fmt.Sprintf("%s/firewall/%s", c.url, firewall.IP), data, []int{http.StatusOK, http.StatusAccepted})
	if err != nil {
		return err
	}

	return nil
}
