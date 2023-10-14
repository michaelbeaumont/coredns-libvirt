package libvirt

import (
	"fmt"
	"net"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

const pluginName = "libvirt"

func init() {
	plugin.Register(pluginName, setup)
}

func setup(c *caddy.Controller) error {
	c.Next()
	if !c.NextArg() {
		return plugin.Error(pluginName, c.ArgErr())
	}
	var h handler
	switch c.Val() {
	case "guest":
		h = handler{}
	default:
		return plugin.Error(pluginName, fmt.Errorf("expected 'guest' as argument"))
	}

	var rules []subnetRules

	for c.NextBlock() {
		var kind ruleKind
		switch c.Val() {
		case "keep":
			kind = keep
		default:
			return plugin.Error(pluginName, fmt.Errorf("unexpected argument: %s", c.Val()))
		}
		if !c.NextArg() {
			return plugin.Error(pluginName, c.ArgErr())
		}
		cidr := c.Val()
		_, net, err := net.ParseCIDR(cidr)
		if err != nil {
			return plugin.Error(pluginName, err)
		}
		rules = append(rules, subnetRules{
			kind: kind,
			cidr: *net,
		})
		if len(c.RemainingArgs()) > 0 {
			return plugin.Error(pluginName, fmt.Errorf("unexpected arguments on line %d", c.Line()))
		}
	}

	h.rules = rules

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		fmt.Println(rules)
		h.Next = next
		return h
	})

	return nil
}
