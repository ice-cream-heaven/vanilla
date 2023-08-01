package adapter

import (
	"errors"
	"fmt"
	"github.com/Dreamacro/clash/constant"
	"github.com/bytedance/sonic"
	"github.com/elliotchance/pie/v2"
	"github.com/go-resty/resty/v2"
	"github.com/ice-cream-heaven/log"
	"github.com/ice-cream-heaven/vanilla/dns"
	"net/netip"
	"time"
)

type DnsMode int

const (
	DnsDisable DnsMode = iota
	DnsDirect
	DnsRemote
)

type Adapter struct {
	constant.ProxyAdapter

	uniqueId string

	client *resty.Client

	opt map[string]any

	// 一些特殊配置
	dnsMode   DnsMode
	resolvers []dns.Resolver
}

func NewAdapter(c constant.ProxyAdapter, o map[string]any) (*Adapter, error) {
	p := &Adapter{
		opt:          o,
		ProxyAdapter: c,
		client: resty.New().
			SetTimeout(time.Minute).
			SetRetryWaitTime(time.Second).
			SetRetryCount(3).
			SetRedirectPolicy(resty.FlexibleRedirectPolicy(10)),
	}

	switch c.Type() {
	case constant.Direct, constant.Reject:
		// do nothing
	default:
		err := p.validateAddr()
		if err != nil {
			return nil, err
		}
	}

	err := p.updateUniqueId()
	if err != nil {
		return nil, err
	}

	{
		p.client.JSONUnmarshal = sonic.Unmarshal
		p.client.JSONMarshal = sonic.Marshal
		p.client.SetTransport(p.Transport()).SetLogger(log.Clone().SetPrefixMsg(fmt.Sprintf("vanilla[%s]", p.ShortId())))
	}

	return p, nil
}

func (p *Adapter) validateAddr() error {
	ip, err := netip.ParseAddr(p.Addr())
	if err != nil {
		return nil
	}

	if ip.IsPrivate() {
		return errors.New("private addr")
	}

	if ip.IsLoopback() {
		return errors.New("loopback addr")
	}

	if ip.IsMulticast() {
		return errors.New("multicast addr")
	}

	if ip.IsUnspecified() {
		return errors.New("unspecified addr")
	}

	if ip.IsLinkLocalUnicast() {
		return errors.New("link local unicast addr")
	}

	if ip.IsLinkLocalMulticast() {
		return errors.New("link local multicast addr")
	}

	if ip.IsInterfaceLocalMulticast() {
		return errors.New("interface local multicast addr")
	}

	if ip.IsInterfaceLocalMulticast() {
		return errors.New("interface local multicast addr")
	}

	if ip.IsGlobalUnicast() {
		return errors.New("global unicast addr")
	}

	if ip.String() == "8.8.8.8" {
		return errors.New("google dns")
	}

	if ip.String() == "1.1.1.1" {
		return errors.New("cloudflare dns")
	}

	return nil
}

func (p *Adapter) DnsMode(m DnsMode, nameservers ...string) *Adapter {
	p.dnsMode = m

	if len(nameservers) > 0 {
		p.resolvers = append([]dns.Resolver{}, pie.Map(nameservers, func(addr string) dns.Resolver {
			res, err := dns.NewResolverWithProxy(addr, p.DialForDns)
			if err != nil {
				log.Panicf("err:%v", err)
			}
			return res
		})...)
	}

	return p
}

func (p *Adapter) ToClash() map[string]any {
	mapping := p.opt
	mapping["name"] = p.Name()

	if p.SupportUDP() {
		mapping["udp"] = true
	} else {
		delete(mapping, "udp")
	}

	if p.SupportXUDP() {
		mapping["xudp"] = true
	} else {
		delete(mapping, "xudp")
	}

	if p.SupportTFO() {
		mapping["tfo"] = true
	} else {
		delete(mapping, "tfo")
	}

	return mapping
}
