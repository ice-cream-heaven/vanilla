package adapter

import (
	"encoding/base64"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/go-resty/resty/v2"
	"github.com/ice-cream-heaven/log"
	"github.com/ice-cream-heaven/utils/json"
	"github.com/ice-cream-heaven/vanilla/dns"
	"github.com/metacubex/mihomo/adapter/outbound"
	"github.com/metacubex/mihomo/constant"
	"golang.org/x/exp/maps"
	"io"
	"net"
	"net/url"
	"strconv"
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

	opt      map[string]any
	clashOpt any

	uniqueId    string
	vanillaLink string

	client *resty.Client

	// 一些特殊配置
	dnsMode   DnsMode
	resolvers []dns.Resolver
}

func NewAdapter(c constant.ProxyAdapter, o any) (*Adapter, error) {
	p := &Adapter{
		ProxyAdapter: c,
		client: resty.New().
			SetTimeout(time.Minute * 10).
			SetRetryWaitTime(time.Second).
			SetRetryCount(3).
			SetLogger(log.Clone().SetOutput(io.Discard)).
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

func (p *Adapter) ToV2ray() string {
	switch p.Type() {
	case constant.Shadowsocks:
		var opt *outbound.ShadowSocksOption
		switch x := p.clashOpt.(type) {
		case outbound.ShadowSocksOption:
			opt = &x
		case *outbound.ShadowSocksOption:
			opt = x
		default:
			return ""
		}

		u := &url.URL{
			Scheme: "ss",
			Host:   net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port)),
		}

		user := url.UserPassword(opt.Cipher, opt.Password)
		u.User = url.User(base64.StdEncoding.EncodeToString([]byte(user.String())))
		u.Fragment = p.Name()

		return u.String()

		//case constant.ShadowsocksR:
		//	var opt *outbound.ShadowSocksROption
		//	switch x := p.clashOpt.(type) {
		//	case outbound.ShadowSocksROption:
		//		opt = &x
		//	case *outbound.ShadowSocksROption:
		//		opt = x
		//	default:
		//		return ""
		//	}

		//case constant.Snell:
		//	var opt *outbound.SnellOption
		//	switch x := p.clashOpt.(type) {
		//	case outbound.SnellOption:
		//		opt = &x
		//	case *outbound.SnellOption:
		//		opt = x
		//	default:
		//		return ""
		//	}
		//
		//	u := &url.URL{
		//		Scheme: "snell",
		//	}
		//
		//	query := u.Query()
		//
		//	if opt.Port > 0 {
		//		u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		//	} else {
		//		u.Host = opt.Server
		//	}
		//
		//	if opt.Psk != "" {
		//		u.User = url.User(opt.Psk)
		//	}
		//
		//	if opt.Version != 0 {
		//		query.Set("version", strconv.Itoa(opt.Version))
		//	}
		//
		//	if len(opt.ObfsOpts) > 0 {
		//		for key, value := range opt.ObfsOpts {
		//			query.Set("obfs-"+key, anyx.ToString(value))
		//		}
		//	}
		//
		//	u.RawQuery = urlx.SortQuery(query).Encode()
		//
		//	p.vanillaLink = u.String()
		//	p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

		//case constant.Socks5:
		//	var opt *outbound.Socks5Option
		//	switch x := p.clashOpt.(type) {
		//	case outbound.Socks5Option:
		//		opt = &x
		//	case *outbound.Socks5Option:
		//		opt = x
		//	default:
		//		return ""
		//	}
		//
		//	u := &url.URL{
		//		Scheme: "socks5",
		//	}
		//
		//	query := u.Query()
		//
		//	if opt.Port > 0 {
		//		u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		//	} else {
		//		u.Host = opt.Server
		//	}
		//
		//	if opt.UserName != "" && opt.Password != "" {
		//		u.User = url.UserPassword(opt.UserName, opt.Password)
		//	} else if opt.UserName != "" {
		//		u.User = url.User(opt.UserName)
		//	} else if opt.Password != "" {
		//		u.User = url.UserPassword("", opt.Password)
		//	}
		//
		//	u.RawQuery = urlx.SortQuery(query).Encode()
		//
		//	p.vanillaLink = u.String()
		//	p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

		//case constant.Http:
		//	var opt *outbound.HttpOption
		//	switch x := p.clashOpt.(type) {
		//	case outbound.HttpOption:
		//		opt = &x
		//	case *outbound.HttpOption:
		//		opt = x
		//	default:
		//		return ""
		//	}
		//
		//	u := &url.URL{
		//		Scheme: "http",
		//	}
		//
		//	query := u.Query()
		//
		//	if opt.TLS {
		//		u.Scheme = "https"
		//	}
		//
		//	if opt.Port > 0 {
		//		u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		//	} else {
		//		u.Host = opt.Server
		//	}
		//
		//	if opt.UserName != "" && opt.Password != "" {
		//		u.User = url.UserPassword(opt.UserName, opt.Password)
		//	} else if opt.UserName != "" {
		//		u.User = url.User(opt.UserName)
		//	} else if opt.Password != "" {
		//		u.User = url.UserPassword("", opt.Password)
		//	}
		//
		//	if opt.SNI != "" {
		//		query.Set("sni", opt.SNI)
		//	}
		//
		//	if len(opt.Headers) > 0 {
		//		for key, value := range opt.Headers {
		//			query.Set("header-"+key, value)
		//		}
		//	}
		//
		//	u.RawQuery = urlx.SortQuery(query).Encode()
		//
		//	p.vanillaLink = u.String()
		//	p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

		//case constant.Vmess:
		//	var opt *outbound.VmessOption
		//	switch x := p.clashOpt.(type) {
		//	case outbound.VmessOption:
		//		opt = &x
		//	case *outbound.VmessOption:
		//		opt = x
		//	default:
		//		return ""
		//	}
		//
		//	u := &url.URL{
		//		Scheme: "vmess",
		//	}
		//
		//	if opt.Port > 0 {
		//		u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		//	} else {
		//		u.Host = opt.Server
		//	}
		//
		//	if opt.UUID != "" {
		//		u.User = url.User(opt.UUID)
		//	}
		//
		//	query := url.Values{}
		//
		//	if opt.ServerName != "" {
		//		query.Set("sni", opt.ServerName)
		//	}
		//
		//	if len(opt.HTTPOpts.Headers) > 0 {
		//		if value, ok := opt.HTTPOpts.Headers["Host"]; ok && len(value) > 0 {
		//			query.Set("http-host", strings.Join(pie.Sort(value), ","))
		//		}
		//	}
		//
		//	if len(opt.HTTPOpts.Path) > 0 {
		//		query.Set("http-path", strings.Join(pie.Sort(opt.HTTPOpts.Path), ","))
		//	}
		//
		//	if len(opt.HTTP2Opts.Host) > 0 {
		//		query.Set("h2-host", strings.Join(pie.Sort(opt.HTTP2Opts.Host), ","))
		//	}
		//
		//	if opt.HTTP2Opts.Path != "" {
		//		query.Set("h2-path", opt.HTTP2Opts.Path)
		//	}
		//
		//	if len(opt.WSOpts.Headers) > 0 {
		//		if value, ok := opt.WSOpts.Headers["Host"]; ok && value != "" {
		//			query.Set("ws-host", value)
		//		}
		//	}
		//
		//	if opt.WSOpts.Path != "" {
		//		query.Set("ws-path", opt.WSOpts.Path)
		//	}
		//
		//	if opt.GrpcOpts.GrpcServiceName != "" {
		//		query.Set("grpc-service-name", opt.GrpcOpts.GrpcServiceName)
		//	}
		//
		//	u.RawQuery = urlx.SortQuery(query).Encode()
		//
		//	p.vanillaLink = u.String()
		//	p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

		//case constant.Vless:
		//	var opt *outbound.VlessOption
		//	switch x := p.clashOpt.(type) {
		//	case outbound.VlessOption:
		//		opt = &x
		//	case *outbound.VlessOption:
		//		opt = x
		//	default:
		//		return ""
		//	}
		//
		//	u := &url.URL{
		//		Scheme: "vless",
		//	}
		//
		//	query := u.Query()
		//
		//	if opt.Port > 0 {
		//		u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		//	} else {
		//		u.Host = opt.Server
		//	}
		//
		//	if opt.UUID != "" {
		//		u.User = url.User(opt.UUID)
		//	}
		//
		//	if opt.Flow != "" {
		//		query.Set("flow", opt.Flow)
		//	}
		//
		//	if len(opt.ALPN) > 0 {
		//		query.Set("alpn", strings.Join(pie.Sort(opt.ALPN), ";"))
		//	}
		//
		//	if opt.PacketEncoding != "" {
		//		query.Set("packet", opt.PacketEncoding)
		//	}
		//
		//	if opt.Network != "" {
		//		query.Set("net", opt.Network)
		//	}
		//
		//	if opt.RealityOpts.ShortID != "" {
		//		query.Set("short-id", opt.RealityOpts.ShortID)
		//	}
		//
		//	if opt.RealityOpts.PublicKey != "" {
		//		query.Set("public-key", opt.RealityOpts.PublicKey)
		//	}
		//
		//	if opt.HTTPOpts.Method != "" {
		//		query.Set("http-method", opt.HTTPOpts.Method)
		//	}
		//
		//	if len(opt.HTTPOpts.Path) > 0 {
		//		query.Set("http-path", strings.Join(pie.Sort(opt.HTTPOpts.Path), ","))
		//	}
		//
		//	if len(opt.HTTPOpts.Headers) > 0 {
		//		if value, ok := opt.HTTPOpts.Headers["Host"]; ok && len(value) > 0 {
		//			query.Set("http-host", strings.Join(pie.Sort(value), ","))
		//		}
		//	}
		//
		//	if len(opt.HTTP2Opts.Host) > 0 {
		//		query.Set("h2-host", strings.Join(pie.Sort(opt.HTTP2Opts.Host), ","))
		//	}
		//
		//	if opt.HTTP2Opts.Path != "" {
		//		query.Set("h2-path", opt.HTTP2Opts.Path)
		//	}
		//
		//	if opt.GrpcOpts.GrpcServiceName != "" {
		//		query.Set("grpc-service-name", opt.GrpcOpts.GrpcServiceName)
		//	}
		//
		//	if opt.WSOpts.Path != "" {
		//		query.Set("ws-path", opt.WSOpts.Path)
		//	}
		//
		//	if len(opt.WSOpts.Headers) > 0 {
		//		if value, ok := opt.WSOpts.Headers["Host"]; ok && value != "" {
		//			query.Set("ws-host", value)
		//		}
		//	}
		//
		//	if opt.WSOpts.V2rayHttpUpgradeFastOpen {
		//		query.Set("v2ray-http-upgrade-fast-open", "true")
		//	}
		//
		//	if opt.WSOpts.EarlyDataHeaderName != "" {
		//		query.Set("early-data-header-name", opt.WSOpts.EarlyDataHeaderName)
		//	}
		//
		//	if opt.WSOpts.V2rayHttpUpgrade {
		//		query.Set("v2ray-http-upgrade", "true")
		//	}
		//
		//	if opt.WSPath != "" {
		//		query.Set("ws-path", opt.WSPath)
		//	}
		//
		//	if len(opt.WSHeaders) > 0 {
		//		if value, ok := opt.WSHeaders["Host"]; ok && value != "" {
		//			query.Set("ws-host", value)
		//		}
		//	}
		//
		//	u.RawQuery = urlx.SortQuery(query).Encode()
		//
		//	p.vanillaLink = u.String()
		//	p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

		//case constant.Trojan:
		//	var opt *outbound.TrojanOption
		//	switch x := p.clashOpt.(type) {
		//	case outbound.TrojanOption:
		//		opt = &x
		//	case *outbound.TrojanOption:
		//		opt = x
		//	default:
		//		return ""
		//	}
		//
		//	u := &url.URL{
		//		Scheme: "trojan",
		//	}
		//	query := u.Query()
		//
		//	if opt.Port > 0 {
		//		u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		//	} else {
		//		u.Host = opt.Server
		//	}
		//
		//	if opt.Password != "" {
		//		u.User = url.User(opt.Password)
		//	}
		//
		//	if opt.Network != "" {
		//		query.Set("net", opt.Network)
		//	}
		//
		//	if opt.GrpcOpts.GrpcServiceName != "" {
		//		query.Set("grpc-service-name", opt.GrpcOpts.GrpcServiceName)
		//	}
		//
		//	if len(opt.WSOpts.Headers) > 0 {
		//		if value, ok := opt.WSOpts.Headers["Host"]; ok && value != "" {
		//			query.Set("ws-host", value)
		//		}
		//	}
		//
		//	if opt.WSOpts.Path != "" {
		//		query.Set("ws-path", opt.WSOpts.Path)
		//	}
		//
		//	if opt.RealityOpts.ShortID != "" {
		//		query.Set("short-id", opt.RealityOpts.ShortID)
		//	}
		//
		//	if opt.RealityOpts.PublicKey != "" {
		//		query.Set("public-key", opt.RealityOpts.PublicKey)
		//	}
		//
		//	u.RawQuery = urlx.SortQuery(query).Encode()
		//
		//	p.vanillaLink = u.String()
		//	p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

		//case constant.Hysteria:
		//	var opt *outbound.HysteriaOption
		//	switch x := p.clashOpt.(type) {
		//	case outbound.HysteriaOption:
		//		opt = &x
		//	case *outbound.HysteriaOption:
		//		opt = x
		//	default:
		//		return ""
		//	}
		//
		//	u := &url.URL{
		//		Scheme: "hysteria",
		//	}
		//
		//	query := u.Query()
		//
		//	if opt.Port > 0 {
		//		u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		//	} else {
		//		u.Host = opt.Server
		//	}
		//
		//	if opt.AuthString != "" {
		//		u.User = url.User(opt.AuthString)
		//	}
		//
		//	if opt.Protocol != "" {
		//		query.Set("protocol", opt.Protocol)
		//	}
		//
		//	if opt.ObfsProtocol != "" {
		//		query.Set("obfs-protocol", opt.ObfsProtocol)
		//	}
		//
		//	if opt.Obfs != "" {
		//		query.Set("obfs", opt.Obfs)
		//	}
		//
		//	if opt.SNI != "" {
		//		query.Set("sni", opt.SNI)
		//	}
		//
		//	if len(opt.ALPN) > 0 {
		//		query.Set("alpn", strings.Join(pie.Sort(opt.ALPN), ";"))
		//	}
		//
		//	if opt.CustomCAString != "" {
		//		query.Set("ca", opt.CustomCAString)
		//	}
		//
		//	if opt.ReceiveWindowConn != 0 {
		//		query.Set("recv-window-conn", strconv.Itoa(opt.ReceiveWindowConn))
		//	}
		//
		//	if opt.ReceiveWindow != 0 {
		//		query.Set("recv-window", strconv.Itoa(opt.ReceiveWindow))
		//	}
		//
		//	if opt.DisableMTUDiscovery {
		//		query.Set("disable-mtu-discovery", "true")
		//	}
		//
		//	if opt.FastOpen {
		//		query.Set("fast-open", "true")
		//	}
		//
		//	if opt.HopInterval != 0 {
		//		query.Set("hop-interval", strconv.Itoa(opt.HopInterval))
		//	}
		//
		//	u.RawQuery = urlx.SortQuery(query).Encode()
		//
		//	p.vanillaLink = u.String()
		//	p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

		//case constant.WireGuard:
		//	var opt *outbound.WireGuardOption
		//	switch x := p.clashOpt.(type) {
		//	case outbound.WireGuardOption:
		//		opt = &x
		//	case *outbound.WireGuardOption:
		//		opt = x
		//	default:
		//		return ""
		//	}
		//
		//	u := &url.URL{
		//		Scheme: "wireguard",
		//	}
		//
		//	query := u.Query()
		//
		//	if opt.Port > 0 {
		//		u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		//	} else {
		//		u.Host = opt.Server
		//	}
		//
		//	if opt.PublicKey != "" {
		//		u.User = url.User(opt.PublicKey)
		//	}
		//
		//	u.RawQuery = urlx.SortQuery(query).Encode()
		//
		//	p.vanillaLink = u.String()
		//	p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

		//case constant.Tuic:
		//	var opt *outbound.TuicOption
		//	switch x := p.clashOpt.(type) {
		//	case outbound.TuicOption:
		//		opt = &x
		//	case *outbound.TuicOption:
		//		opt = x
		//	default:
		//		return ""
		//	}
		//
		//	u := &url.URL{
		//		Scheme: "tuic",
		//	}
		//
		//	query := u.Query()
		//
		//	if opt.Port > 0 {
		//		u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		//	} else {
		//		u.Host = opt.Server
		//	}
		//
		//	if opt.UUID != "" && opt.Password != "" {
		//		u.User = url.UserPassword(opt.UUID, opt.Password)
		//	} else if opt.UUID != "" {
		//		u.User = url.User(opt.UUID)
		//	} else if opt.Password != "" {
		//		u.User = url.UserPassword("", opt.Password)
		//	}
		//
		//	if len(opt.ALPN) > 0 {
		//		query.Set("alpn", strings.Join(pie.Sort(opt.ALPN), ";"))
		//	}
		//
		//	if opt.ReduceRtt {
		//		query.Set("reduce-rtt", "true")
		//	}
		//
		//	if opt.UDPOverStream {
		//		query.Set("udp-over-stream", "true")
		//	}
		//
		//	if opt.UdpRelayMode != "" {
		//		query.Set("udp-relay-mode", opt.UdpRelayMode)
		//	}
		//
		//	if opt.CongestionController != "" {
		//		query.Set("congestion-controller", opt.CongestionController)
		//	}
		//
		//	if opt.DisableSni {
		//		query.Set("disable-sniffing", "true")
		//	}
		//
		//	if opt.FastOpen {
		//		query.Set("fast-open", "true")
		//	}
		//
		//	if opt.CWND != 0 {
		//		query.Set("cwnd", strconv.Itoa(opt.CWND))
		//	}
		//
		//	if opt.CustomCAString != "" {
		//		query.Set("ca", opt.CustomCAString)
		//	}
		//
		//	if opt.SNI != "" {
		//		query.Set("sni", opt.SNI)
		//	}
		//
		//	if opt.ReceiveWindow != 0 {
		//		query.Set("recv-window", strconv.Itoa(opt.ReceiveWindow))
		//	}
		//
		//	if opt.ReceiveWindowConn != 0 {
		//		query.Set("recv-window-conn", strconv.Itoa(opt.ReceiveWindowConn))
		//	}
		//
		//	if opt.UDPOverStream {
		//		query.Set("udp-over-stream", "true")
		//	}
		//
		//	if opt.DisableMTUDiscovery {
		//		query.Set("disable-mtu-discovery", "true")
		//	}
		//
		//	if opt.UDPOverStreamVersion != 0 {
		//		query.Set("udp-over-stream-version", strconv.Itoa(opt.UDPOverStreamVersion))
		//	}
		//
		//	u.RawQuery = urlx.SortQuery(query).Encode()
		//
		//	p.vanillaLink = u.String()
		//	p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

	}

	return ""
}

func (p *Adapter) TypeString() string {
	switch p.Type() {
	case constant.Direct:
		return "direct"
	case constant.Reject:
		return "reject"

	case constant.Shadowsocks:
		return "ss"
	case constant.ShadowsocksR:
		return "ssr"
	case constant.Snell:
		return "snell"
	case constant.Socks5:
		return "socks5"
	case constant.Http:
		return "http"
	case constant.Vmess:
		return "vmess"
	case constant.Vless:
		return "vless"
	case constant.Trojan:
		return "trojan"
	case constant.Hysteria:
		return "hysteria"
	case constant.Hysteria2:
		return "hysteria2"
	case constant.WireGuard:
		return "wireguard"
	case constant.Tuic:
		return "tuic"
	default:
		panic("unknown type")
	}
}
