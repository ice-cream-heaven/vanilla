package adapter

import (
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/go-resty/resty/v2"
	"github.com/ice-cream-heaven/log"
	"github.com/ice-cream-heaven/utils/json"
	"github.com/ice-cream-heaven/vanilla/dns"
	"github.com/metacubex/mihomo/constant"
	"golang.org/x/exp/maps"
	"net"
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

	opt map[string]any

	uniqueId string

	client *resty.Client

	// 一些特殊配置
	dnsMode   DnsMode
	resolvers []dns.Resolver
}

func NewAdapter(c constant.ProxyAdapter, o any) (*Adapter, error) {
	p := &Adapter{
		ProxyAdapter: c,
		client: resty.New().
			SetTimeout(time.Second * 10).
			SetRetryWaitTime(time.Second).
			SetRetryCount(3).
			SetRedirectPolicy(resty.FlexibleRedirectPolicy(10)),
	}

	//switch c.Type() {
	//case constant.Direct, constant.Reject:
	//	// do nothing
	//default:
	//	err := p.validateAddr()
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	err := p.updateUniqueId(o)
	if err != nil {
		return nil, err
	}

	{
		p.client.JSONUnmarshal = json.Unmarshal
		p.client.JSONMarshal = json.Marshal
		p.client.SetTransport(p.Transport()).
			SetLogger(log.Clone().SetPrefixMsg(fmt.Sprintf("vanilla[%s]", p.ShortId())))
	}

	return p, nil
}

//func (p *Adapter) validateAddr() error {
//	host, _, err := net.SplitHostPort(p.Addr())
//	if err != nil {
//		return err
//	}
//
//	ip, err := netip.ParseAddr(host)
//	if err != nil {
//		return nil
//	}
//
//	if ip.IsPrivate() {
//		return errors.New("private addr")
//	}
//
//	if ip.IsLoopback() {
//		return errors.New("loopback addr")
//	}
//
//	if ip.String() == "8.8.8.8" {
//		return errors.New("google dns")
//	}
//
//	if ip.String() == "1.1.1.1" {
//		return errors.New("cloudflare dns")
//	}
//
//	return nil
//}

func (p *Adapter) Hostname() string {
	host, _, _ := net.SplitHostPort(p.Addr())
	return host
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
	mapping := map[string]any{}
	maps.Copy(mapping, p.opt)
	mapping["name"] = p.Name()
	mapping["unique_id"] = p.uniqueId
	return mapping
}
