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

	for idx, rule := range firewall.Rules.Input {
		ipVersion := rule.IPVersion
		if ipVersion == "" {
			ipVersion = "ipv4"
		}
		data.Set(fmt.Sprintf("rules[input][%d][%s]", idx, "ip_version"), ipVersion)
		if rule.Name != "" {
			data.Set(fmt.Sprintf("rules[input][%d][%s]", idx, "name"), rule.Name)
		}
		if rule.DstIP != "" {
			data.Set(fmt.Sprintf("rules[input][%d][%s]", idx, "dst_ip"), rule.DstIP)
		}
		if rule.DstPort != "" {
			data.Set(fmt.Sprintf("rules[input][%d][%s]", idx, "dst_port"), rule.DstPort)
		}
		if rule.SrcIP != "" {
			data.Set(fmt.Sprintf("rules[input][%d][%s]", idx, "src_ip"), rule.SrcIP)
		}
		if rule.SrcPort != "" {
			data.Set(fmt.Sprintf("rules[input][%d][%s]", idx, "src_port"), rule.SrcPort)
		}
		if rule.Protocol != "" {
			data.Set(fmt.Sprintf("rules[input][%d][%s]", idx, "protocol"), rule.Protocol)
		}
		if rule.TCPFlags != "" {
			data.Set(fmt.Sprintf("rules[input][%d][%s]", idx, "tcp_flags"), rule.TCPFlags)
		}
		data.Set(fmt.Sprintf("rules[input][%d][%s]", idx, "action"), rule.Action)
	}

	_, err := c.makeAPICall(ctx, "POST", fmt.Sprintf("%s/firewall/%s", c.url, firewall.IP), data, []int{http.StatusOK, http.StatusAccepted})
	if err != nil {
		return err
	}

	return nil
}
