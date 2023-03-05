package server

import (
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/yaskoo/provider-hetzner/apis/cloud/v1alpha1"
)

func toServerCreateOpts(name string, sp v1alpha1.ServerParameters) hcloud.ServerCreateOpts {
	opts := hcloud.ServerCreateOpts{
		Name: name,
		ServerType: &hcloud.ServerType{
			ID:   int(sp.ServerType.IntVal),
			Name: sp.ServerType.StrVal,
		},
		Image: &hcloud.Image{
			ID:   int(sp.Image.IntVal),
			Name: sp.Image.StrVal,
		},
		StartAfterCreate: sp.StartAfterCreate,
		Automount:        sp.Automount,
	}

	if sp.SSHKeys != nil {
		var sshKeys []*hcloud.SSHKey
		for _, key := range *sp.SSHKeys {
			sshKeys = append(sshKeys, &hcloud.SSHKey{
				ID:   int(key.IntVal),
				Name: key.StrVal,
			})
		}
		opts.SSHKeys = sshKeys
	}

	if sp.Location != nil {
		opts.Location = &hcloud.Location{
			ID:   int(sp.Location.IntVal),
			Name: sp.Location.StrVal,
		}
	}

	if sp.Datacenter != nil {
		opts.Datacenter = &hcloud.Datacenter{
			ID:   int(sp.Datacenter.IntVal),
			Name: sp.Datacenter.StrVal,
		}
	}

	if sp.UserData != nil {
		opts.UserData = *sp.UserData
	}

	if sp.Labels != nil {
		opts.Labels = sp.Labels
	}

	if sp.Volumes != nil {
		var volumes []*hcloud.Volume
		for _, v := range *sp.Volumes {
			volumes = append(volumes, &hcloud.Volume{
				ID: v,
			})
		}
		opts.Volumes = volumes
	}

	if sp.Networks != nil {
		var nets []*hcloud.Network
		for _, v := range *sp.Networks {
			nets = append(nets, &hcloud.Network{
				ID: v,
			})
		}
		opts.Networks = nets
	}

	if sp.Firewalls != nil {
		var firewalls []*hcloud.ServerCreateFirewall
		for _, v := range *sp.Firewalls {
			firewalls = append(firewalls, &hcloud.ServerCreateFirewall{
				Firewall: hcloud.Firewall{
					ID: v,
				},
			})
		}
		opts.Firewalls = firewalls
	}

	if sp.PlacementGroup != nil {
		opts.PlacementGroup = &hcloud.PlacementGroup{
			ID: *sp.PlacementGroup,
		}
	}

	if sp.PublicNet != nil {
		opts.PublicNet = &hcloud.ServerCreatePublicNet{
			EnableIPv4: *sp.PublicNet.EnableIPv4,
			EnableIPv6: *sp.PublicNet.EnableIPv6,
			IPv4: &hcloud.PrimaryIP{
				ID: *sp.PublicNet.IPv4,
			},
			IPv6: &hcloud.PrimaryIP{
				ID: *sp.PublicNet.IPv6,
			},
		}
	}
	return opts
}
