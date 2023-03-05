package firewall

import (
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/yaskoo/provider-hetzner/apis/cloud/v1alpha1"
	"net"
)

func toIPNets(cidrs []string) ([]net.IPNet, error) {
	if cidrs == nil || len(cidrs) == 0 {
		return nil, nil
	}

	parsed := make([]net.IPNet, len(cidrs))
	for idx, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}

		parsed[idx] = *ipNet
	}
	return parsed, nil
}

func toFirewallRules(rules []v1alpha1.FirewallRule) ([]hcloud.FirewallRule, error) {
	if rules == nil || len(rules) == 0 {
		return nil, nil
	}

	mapped := make([]hcloud.FirewallRule, len(rules))
	for idx, rule := range rules {
		sources, err := toIPNets(rule.SourceIPs)
		if err != nil {
			return nil, err
		}

		destinations, err := toIPNets(rule.DestinationIPs)
		if err != nil {
			return nil, err
		}

		mapped[idx] = hcloud.FirewallRule{
			Direction:      hcloud.FirewallRuleDirection(rule.Direction),
			SourceIPs:      sources,
			DestinationIPs: destinations,
			Protocol:       hcloud.FirewallRuleProtocol(rule.Protocol),
			Port:           rule.Port,
			Description:    rule.Description,
		}
	}
	return mapped, nil
}

func toFirewallResources(fromSpec []v1alpha1.FirewallResource) ([]hcloud.FirewallResource, error) {
	if fromSpec == nil || len(fromSpec) == 0 {
		return nil, nil
	}

	mapped := make([]hcloud.FirewallResource, len(fromSpec))
	for idx, res := range fromSpec {
		firewallResource := hcloud.FirewallResource{
			Type: hcloud.FirewallResourceType(res.Type),
		}

		if res.LabelSelector != nil {
			firewallResource.LabelSelector = &hcloud.FirewallResourceLabelSelector{
				Selector: *res.LabelSelector,
			}
		}

		if res.Server != nil {
			firewallResource.Server = &hcloud.FirewallResourceServer{
				ID: *res.Server,
			}
		}
		mapped[idx] = firewallResource
	}
	return mapped, nil
}
