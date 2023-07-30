package adapter

import (
	"fmt"
	"github.com/Dreamacro/clash/constant"
	"github.com/bytedance/sonic"
	"github.com/elliotchance/pie/v2"
	"github.com/go-resty/resty/v2"
	"github.com/ice-cream-heaven/log"
	"github.com/ice-cream-heaven/vanilla/dns"
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
