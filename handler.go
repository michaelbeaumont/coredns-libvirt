package libvirt

import (
	"context"
	"net"
	"slices"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
)

type handler struct {
	ttl   uint32
	rules []subnetRules
	Next  plugin.Handler
}

type ruleKind int

const (
	keep ruleKind = iota
)

type subnetRules struct {
	kind ruleKind
	cidr net.IPNet
}

var _ plugin.Handler = handler{}

func (h handler) Name() string {
	return pluginName
}

func (h handler) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	if state.QClass() != dns.ClassINET {
		return plugin.NextOrFailure(pluginName, h.Next, ctx, w, r)
	}

	vmName := strings.TrimSuffix(state.QName(), ".")

	var ips []net.IP
	switch state.QType() {
	case dns.TypeA, dns.TypeAAAA:
		var err error
		ips, err = findGuestIPs(vmName)
		if err != nil {
			log.Error(plugin.Error(pluginName, errors.Wrap(err, "couldn't lookup IPs of libvirt domains")))
			return dns.RcodeServerFailure, nil
		}
	default:
		return plugin.NextOrFailure(pluginName, h.Next, ctx, w, r)
	}

	// Only if we have _no_ IPs at all of any family do we go to the next plugin
	if len(ips) == 0 {
		return plugin.NextOrFailure(pluginName, h.Next, ctx, w, r)
	}

	if len(h.rules) > 0 {
		ips = slices.DeleteFunc(ips, func(ip net.IP) bool {
			for _, rule := range h.rules {
				switch rule.kind {
				case keep:
					if rule.cidr.Contains(ip) {
						return false
					}
				}
			}
			return true
		})
	}

	var answer []dns.RR
	switch state.QType() {
	case dns.TypeA:
		slices.DeleteFunc(ips, func(ip net.IP) bool {
			return ip.To4() == nil
		})
		answer = a(state.QName(), h.ttl, ips)
	case dns.TypeAAAA:
		slices.DeleteFunc(ips, func(ip net.IP) bool {
			return ip.To4() != nil
		})
		answer = aaaa(state.QName(), h.ttl, ips)
	}

	resp := dns.Msg{}
	resp.SetReply(r)
	resp.Authoritative = true
	resp.Answer = answer

	if err := w.WriteMsg(&resp); err != nil {
		log.Error(plugin.Error(pluginName, errors.Wrap(err, "couldn't write message")))
		return dns.RcodeServerFailure, nil
	}
	return dns.RcodeSuccess, nil
}

func a(zone string, ttl uint32, ips []net.IP) []dns.RR {
	answers := make([]dns.RR, len(ips))
	for i, ip := range ips {
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl}
		r.A = ip
		answers[i] = r
	}
	return answers
}

func aaaa(zone string, ttl uint32, ips []net.IP) []dns.RR {
	answers := make([]dns.RR, len(ips))
	for i, ip := range ips {
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl}
		r.AAAA = ip
		answers[i] = r
	}
	return answers
}
