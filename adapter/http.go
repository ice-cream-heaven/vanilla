package adapter

import (
	"context"
	"github.com/Dreamacro/clash/component/dialer"
	"github.com/Dreamacro/clash/constant"
	"github.com/go-resty/resty/v2"
	"github.com/ice-cream-heaven/log"
	"github.com/ice-cream-heaven/vanilla/dns"
	"net"
	"net/http"
	"net/netip"
	"time"
)

func (p *Adapter) HttpDial(network, addr string) (net.Conn, error) {
	return p.HttpDialContext(context.Background(), network, addr)
}

func (p *Adapter) HttpDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	meta := &constant.Metadata{}
	switch network {
	case "tcp", "tcp4", "tcp6":
		meta.NetWork = constant.TCP
	case "udp", "udp4", "udp6":
		meta.NetWork = constant.UDP
	default:
		meta.NetWork = constant.InvalidNet
	}

	meta.Host, meta.DstPort, _ = net.SplitHostPort(addr)

	switch p.dnsMode {
	case DnsDisable:
	// do nothing
	case DnsDirect:
		ip, err := netip.ParseAddr(meta.Host)
		if err == nil {
			// host是一个ip
			meta.DstIP = ip
		} else {
			ip, err := dns.DefaultResolver.LookupHost(meta.Host)
			if err == nil {
				meta.DstIP = netip.AddrFrom4([4]byte(ip))
				meta.Host = ""
				meta.DNSMode = constant.DNSMapping
			}
		}
	case DnsRemote:
		// TODO: 优化
		ip, err := netip.ParseAddr(meta.Host)
		if err == nil {
			meta.DstIP = ip
		} else {
			for _, resolver := range p.resolvers {
				ips, err := resolver.LookupIPv4(meta.Host)
				if err != nil {
					log.Errorf("err:%v", err)
					continue
				}

				if len(ips) == 0 {
					continue
				}

				meta.DstIP = netip.AddrFrom4([4]byte(ips[0]))
			}
		}
	}

	return p.ProxyAdapter.DialContext(ctx, meta, dialer.WithPreferIPv4())
}

func (p *Adapter) Transport() http.RoundTripper {
	return &http.Transport{
		DialContext: p.HttpDialContext,
	}
}

func (p *Adapter) GetClient() *http.Client {
	return &http.Client{
		Transport: p.Transport(),
	}
}

func (p *Adapter) GetClientWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: p.HttpDialContext,
		},
		Timeout: timeout,
	}
}

func (p *Adapter) R() *resty.Request {
	return p.client.R()
}
