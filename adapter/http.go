package adapter

import (
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/ice-cream-heaven/log"
	"github.com/ice-cream-heaven/vanilla/dns"
	"github.com/metacubex/mihomo/component/dialer"
	"github.com/metacubex/mihomo/constant"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"time"
)

func dialContext2Metadata(network, addr string) *constant.Metadata {
	meta := &constant.Metadata{}
	switch network {
	case "tcp", "tcp4", "tcp6":
		meta.NetWork = constant.TCP
	case "udp", "udp4", "udp6":
		meta.NetWork = constant.UDP
	default:
		meta.NetWork = constant.InvalidNet
	}

	var portStr string
	meta.Host, portStr, _ = net.SplitHostPort(addr)

	port, _ := strconv.ParseUint(portStr, 10, 16)
	meta.DstPort = uint16(port)

	return meta
}

func (p *Adapter) dnsQuery(host string) (ip netip.Addr) {
	switch p.dnsMode {
	case DnsDisable:
		// do nothing
	case DnsDirect:
		ip, _ = netip.ParseAddr(host)
		if !ip.IsValid() {
			_ip, _ := dns.DefaultResolver.LookupHost(host)
			if _ip != nil {
				ip = netip.AddrFrom4([4]byte(_ip))
				log.Debugf("use default dns:%v", ip)
			}
		}
	case DnsRemote:
		ip, _ = netip.ParseAddr(host)
		if !ip.IsValid() {
			for _, resolver := range p.resolvers {
				ips, err := resolver.LookupIPv4(host)
				if err != nil {
					log.Errorf("err:%v", err)
					continue
				}

				if len(ips) == 0 {
					continue
				}

				ip = netip.AddrFrom4([4]byte(ips[0]))
				break
			}
		}

		log.Debugf("use remote dns:%v", ip)
	}

	return
}

func (p *Adapter) HttpDial(network, addr string) (net.Conn, error) {
	return p.HttpDialContext(context.Background(), network, addr)
}

func (p *Adapter) DialForDns(network, addr string) (net.Conn, error) {
	meta := dialContext2Metadata(network, addr)
	meta.DstIP = p.dnsQuery(meta.Host)
	if meta.DstIP.IsValid() {
		meta.Host = ""
		meta.DNSMode = constant.DNSFakeIP
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	return p.ProxyAdapter.DialContext(
		ctx, meta,
		dialer.WithNetDialer(&net.Dialer{
			Timeout:   time.Second,
			KeepAlive: 30 * time.Second,
		}),
	)
}

func (p *Adapter) HttpDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	meta := dialContext2Metadata(network, addr)
	meta.DstIP = p.dnsQuery(meta.Host)
	if meta.DstIP.IsValid() {
		meta.Host = ""
		meta.DNSMode = constant.DNSFakeIP
	}

	return p.ProxyAdapter.DialContext(ctx, meta, dialer.WithPreferIPv4())
}

func (p *Adapter) HttpDialDialer(ctx context.Context, network, addr string, opts ...dialer.Option) (net.Conn, error) {
	meta := dialContext2Metadata(network, addr)
	meta.DstIP = p.dnsQuery(meta.Host)
	if meta.DstIP.IsValid() {
		meta.Host = ""
		meta.DNSMode = constant.DNSFakeIP
	}

	return p.ProxyAdapter.DialContext(ctx, meta, opts...)
}

func (p *Adapter) Transport() http.RoundTripper {
	return &http.Transport{
		DialContext: p.HttpDialContext,
	}
}

func (p *Adapter) GetClient() *http.Client {
	return p.client.GetClient()
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
