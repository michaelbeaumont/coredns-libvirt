package libvirt

import (
	"fmt"

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
	switch c.Val() {
	case "guest":
		dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
			return handler{Next: next}
		})
	default:
		return plugin.Error(pluginName, fmt.Errorf("expected 'guest' as argument"))
	}

	return nil
}
