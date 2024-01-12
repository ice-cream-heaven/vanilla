package adapter

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/ice-cream-heaven/log"
	"github.com/ice-cream-heaven/utils/anyx"
	"github.com/ice-cream-heaven/utils/cryptox"
	"github.com/ice-cream-heaven/utils/urlx"
	"github.com/metacubex/mihomo/adapter/outbound"
	"github.com/metacubex/mihomo/common/structure"
	"github.com/metacubex/mihomo/constant"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func (p *Adapter) UniqueId() string {
	return p.uniqueId
}

func (p *Adapter) Md5() string {
	return cryptox.Md5(p.String())
}

func (p *Adapter) ShortId() string {
	if len(p.uniqueId) < 8 {
		return p.uniqueId
	}
	return p.uniqueId[:8]
}

func (p *Adapter) String() string {
	b := log.GetBuffer()
	defer log.PutBuffer(b)

	switch p.Type() {
	case constant.Direct:
		b.WriteString("direct://localhost")
	case constant.Reject:
		b.WriteString("reject://black-hole")
	default:
		b.WriteString(strings.ToLower(p.Type().String()))
		b.WriteString("://")
		b.WriteString(p.Addr())
	}

	return b.String()
}

func (p *Adapter) updateUniqueId(src any) error {
	switch p.Type() {
	case constant.Direct:
		p.uniqueId = "direct"
		p.opt = map[string]any{
			"type": "direct",
		}
		return nil
	case constant.Reject:
		p.uniqueId = "reject"
		p.opt = map[string]any{
			"type": "reject",
		}
		return nil
	}

	var o any
	if val, ok := src.(map[string]any); ok {
		p.opt = val
		typ, ok := val["type"]
		if !ok {
			return errors.New("miss type")
		}

		var opt any
		switch typ {
		case "ss", "shadowsocks":
			opt = &outbound.ShadowSocksOption{}

		case "ssr", "shadowsocksr":
			opt = &outbound.ShadowSocksROption{}

		case "snell":
			opt = &outbound.SnellOption{}

		case "socks", "socks5", "socks4":
			opt = &outbound.Socks5Option{}

		case "http", "https":
			opt = &outbound.HttpOption{}

		case "vmess":
			opt = &outbound.VmessOption{}

		case "vless":
			opt = &outbound.VlessOption{}

		case "trojan":
			opt = &outbound.TrojanOption{}

		case "hysteria":
			opt = &outbound.HysteriaOption{}

		case "wireguard":
			opt = &outbound.WireGuardOption{}

		case "tuic":
			opt = &outbound.TuicOption{}

		case "direct":
			opt = &outbound.Direct{}

		case "reject":
			opt = &outbound.Reject{}

		default:
			log.Errorf("invalid type: %s", typ)
			return errors.New("invalid type")
		}

		err := decoder.Decode(val, opt)
		if err != nil {
			return err
		}

		o = opt
	} else {
		o = src

		var err error
		p.opt, err = encode(src)
		if err != nil {
			return err
		}

		if p.SupportUDP() {
			p.opt["udp"] = true
		} else {
			delete(p.opt, "udp")
		}

		if p.SupportXUDP() {
			p.opt["xudp"] = true
		} else {
			delete(p.opt, "xudp")
		}

		if p.SupportTFO() {
			p.opt["tfo"] = true
		} else {
			delete(p.opt, "tfo")
		}
		p.opt["type"] = strings.ToLower(p.Type().String())
	}

	p.clashOpt = o

	switch p.Type() {
	case constant.Shadowsocks:
		var opt *outbound.ShadowSocksOption
		switch x := o.(type) {
		case outbound.ShadowSocksOption:
			opt = &x
		case *outbound.ShadowSocksOption:
			opt = x
		default:
			return fmt.Errorf("invalid option type: %T", o)
		}

		u := &url.URL{
			Scheme: "ss",
		}
		query := u.Query()

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.Password != "" {
			u.User = url.User(opt.Password)
		}

		if opt.Plugin != "" {
			query.Set("plugin", opt.Plugin)

			decoder := structure.NewDecoder(structure.Option{TagName: "obfs", WeaklyTypedInput: true})

			switch opt.Plugin {
			case "obfs":
				o := simpleObfsOption{Host: "bing.com"}
				err := decoder.Decode(opt.PluginOpts, &o)
				if err != nil {
					log.Errorf("err:%v", err)
					return err
				}

				query.Set("obfs-mode", o.Mode)
				query.Set("obfs-host", o.Host)

			case "v2ray-plugin":
				o := v2rayObfsOption{Host: "bing.com", Mux: true}
				err := decoder.Decode(opt.PluginOpts, &o)
				if err != nil {
					log.Errorf("err:%v", err)
					return err
				}

				query.Set("obfs-mode", o.Mode)
				query.Set("obfs-host", o.Host)
				query.Set("obfs-path", o.Path)
				if o.TLS {
					query.Set("obfs-tls", "true")
				}
				if o.Mux {
					query.Set("mux", "true")
				}
				if len(o.Headers) > 0 {
					for k, v := range o.Headers {
						query.Set("obfs-header."+k, v)
					}
				}
			}
		}

		u.RawQuery = urlx.SortQuery(query).Encode()

		p.vanillaLink = u.String()
		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

	case constant.ShadowsocksR:
		var opt *outbound.ShadowSocksROption
		switch x := o.(type) {
		case outbound.ShadowSocksROption:
			opt = &x
		case *outbound.ShadowSocksROption:
			opt = x
		default:
			return fmt.Errorf("invalid option type: %T", o)
		}

		u := &url.URL{
			Scheme: "ssr",
		}
		query := u.Query()

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.Password != "" {
			u.User = url.User(opt.Password)
		}

		if opt.Cipher != "" {
			query.Set("cipher", opt.Cipher)
		}

		if opt.Obfs != "" {
			query.Set("obfs", opt.Obfs)
		}

		if opt.ObfsParam != "" {
			query.Set("obfs-param", opt.ObfsParam)
		}

		if opt.Protocol != "" {
			query.Set("protocol", opt.Protocol)
		}

		if opt.ProtocolParam != "" {
			query.Set("protocol-param", opt.ProtocolParam)
		}

		u.RawQuery = urlx.SortQuery(query).Encode()

		p.vanillaLink = u.String()
		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

	case constant.Snell:
		var opt *outbound.SnellOption
		switch x := o.(type) {
		case outbound.SnellOption:
			opt = &x
		case *outbound.SnellOption:
			opt = x
		default:
			return fmt.Errorf("invalid option type: %T", o)
		}

		u := &url.URL{
			Scheme: "snell",
		}

		query := u.Query()

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.Psk != "" {
			u.User = url.User(opt.Psk)
		}

		if opt.Version != 0 {
			query.Set("version", strconv.Itoa(opt.Version))
		}

		if len(opt.ObfsOpts) > 0 {
			for key, value := range opt.ObfsOpts {
				query.Set("obfs-"+key, anyx.ToString(value))
			}
		}

		u.RawQuery = urlx.SortQuery(query).Encode()

		p.vanillaLink = u.String()
		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

	case constant.Socks5:
		var opt *outbound.Socks5Option
		switch x := o.(type) {
		case outbound.Socks5Option:
			opt = &x
		case *outbound.Socks5Option:
			opt = x
		default:
			return fmt.Errorf("invalid option type: %T", o)
		}

		u := &url.URL{
			Scheme: "socks5",
		}

		query := u.Query()

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.UserName != "" && opt.Password != "" {
			u.User = url.UserPassword(opt.UserName, opt.Password)
		} else if opt.UserName != "" {
			u.User = url.User(opt.UserName)
		} else if opt.Password != "" {
			u.User = url.UserPassword("", opt.Password)
		}

		u.RawQuery = urlx.SortQuery(query).Encode()

		p.vanillaLink = u.String()
		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

	case constant.Http:
		var opt *outbound.HttpOption
		switch x := o.(type) {
		case outbound.HttpOption:
			opt = &x
		case *outbound.HttpOption:
			opt = x
		default:
			return fmt.Errorf("invalid option type: %T", o)
		}

		u := &url.URL{
			Scheme: "http",
		}

		query := u.Query()

		if opt.TLS {
			u.Scheme = "https"
		}

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.UserName != "" && opt.Password != "" {
			u.User = url.UserPassword(opt.UserName, opt.Password)
		} else if opt.UserName != "" {
			u.User = url.User(opt.UserName)
		} else if opt.Password != "" {
			u.User = url.UserPassword("", opt.Password)
		}

		if opt.SNI != "" {
			query.Set("sni", opt.SNI)
		}

		if len(opt.Headers) > 0 {
			for key, value := range opt.Headers {
				query.Set("header-"+key, value)
			}
		}

		u.RawQuery = urlx.SortQuery(query).Encode()

		p.vanillaLink = u.String()
		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

	case constant.Vmess:
		var opt *outbound.VmessOption
		switch x := o.(type) {
		case outbound.VmessOption:
			opt = &x
		case *outbound.VmessOption:
			opt = x
		default:
			return fmt.Errorf("invalid option type: %T", o)
		}

		u := &url.URL{
			Scheme: "vmess",
		}

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.UUID != "" {
			u.User = url.User(opt.UUID)
		}

		query := url.Values{}

		if opt.ServerName != "" {
			query.Set("sni", opt.ServerName)
		}

		if len(opt.HTTPOpts.Headers) > 0 {
			if value, ok := opt.HTTPOpts.Headers["Host"]; ok && len(value) > 0 {
				query.Set("http-host", strings.Join(pie.Sort(value), ","))
			}
		}

		if len(opt.HTTPOpts.Path) > 0 {
			query.Set("http-path", strings.Join(pie.Sort(opt.HTTPOpts.Path), ","))
		}

		if len(opt.HTTP2Opts.Host) > 0 {
			query.Set("h2-host", strings.Join(pie.Sort(opt.HTTP2Opts.Host), ","))
		}

		if opt.HTTP2Opts.Path != "" {
			query.Set("h2-path", opt.HTTP2Opts.Path)
		}

		if len(opt.WSOpts.Headers) > 0 {
			if value, ok := opt.WSOpts.Headers["Host"]; ok && value != "" {
				query.Set("ws-host", value)
			}
		}

		if opt.WSOpts.Path != "" {
			query.Set("ws-path", opt.WSOpts.Path)
		}

		if opt.GrpcOpts.GrpcServiceName != "" {
			query.Set("grpc-service-name", opt.GrpcOpts.GrpcServiceName)
		}

		u.RawQuery = urlx.SortQuery(query).Encode()

		p.vanillaLink = u.String()
		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

	case constant.Vless:
		var opt *outbound.VlessOption
		switch x := o.(type) {
		case outbound.VlessOption:
			opt = &x
		case *outbound.VlessOption:
			opt = x
		default:
			return fmt.Errorf("invalid option type: %T", o)
		}

		u := &url.URL{
			Scheme: "vless",
		}

		query := u.Query()

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.UUID != "" {
			u.User = url.User(opt.UUID)
		}

		if opt.Flow != "" {
			query.Set("flow", opt.Flow)
		}

		if opt.PacketEncoding != "" {
			query.Set("packet", opt.PacketEncoding)
		}

		if opt.Network != "" {
			query.Set("net", opt.Network)
		}

		if opt.RealityOpts.ShortID != "" {
			query.Set("short-id", opt.RealityOpts.ShortID)
		}

		if opt.RealityOpts.PublicKey != "" {
			query.Set("public-key", opt.RealityOpts.PublicKey)
		}

		if opt.HTTPOpts.Method != "" {
			query.Set("http-method", opt.HTTPOpts.Method)
		}

		if len(opt.HTTPOpts.Path) > 0 {
			query.Set("http-path", strings.Join(pie.Sort(opt.HTTPOpts.Path), ","))
		}

		if len(opt.HTTPOpts.Headers) > 0 {
			if value, ok := opt.HTTPOpts.Headers["Host"]; ok && len(value) > 0 {
				query.Set("http-host", strings.Join(pie.Sort(value), ","))
			}
		}

		if len(opt.HTTP2Opts.Host) > 0 {
			query.Set("h2-host", strings.Join(pie.Sort(opt.HTTP2Opts.Host), ","))
		}

		if opt.HTTP2Opts.Path != "" {
			query.Set("h2-path", opt.HTTP2Opts.Path)
		}

		if opt.GrpcOpts.GrpcServiceName != "" {
			query.Set("grpc-service-name", opt.GrpcOpts.GrpcServiceName)
		}

		if opt.WSOpts.Path != "" {
			query.Set("ws-path", opt.WSOpts.Path)
		}

		if len(opt.WSOpts.Headers) > 0 {
			if value, ok := opt.WSOpts.Headers["Host"]; ok && value != "" {
				query.Set("ws-host", value)
			}
		}

		if opt.WSOpts.V2rayHttpUpgradeFastOpen {
			query.Set("v2ray-http-upgrade-fast-open", "true")
		}

		if opt.WSOpts.EarlyDataHeaderName != "" {
			query.Set("early-data-header-name", opt.WSOpts.EarlyDataHeaderName)
		}

		if opt.WSOpts.V2rayHttpUpgrade {
			query.Set("v2ray-http-upgrade", "true")
		}

		if opt.WSPath != "" {
			query.Set("ws-path", opt.WSPath)
		}

		if len(opt.WSHeaders) > 0 {
			if value, ok := opt.WSHeaders["Host"]; ok && value != "" {
				query.Set("ws-host", value)
			}
		}

		u.RawQuery = urlx.SortQuery(query).Encode()

		p.vanillaLink = u.String()
		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

	case constant.Trojan:
		var opt *outbound.TrojanOption
		switch x := o.(type) {
		case outbound.TrojanOption:
			opt = &x
		case *outbound.TrojanOption:
			opt = x
		default:
			return fmt.Errorf("invalid option type: %T", o)
		}

		u := &url.URL{
			Scheme: "trojan",
		}
		query := u.Query()

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.Password != "" {
			u.User = url.User(opt.Password)
		}

		if opt.Network != "" {
			query.Set("net", opt.Network)
		}

		if opt.GrpcOpts.GrpcServiceName != "" {
			query.Set("grpc-service-name", opt.GrpcOpts.GrpcServiceName)
		}

		if len(opt.WSOpts.Headers) > 0 {
			if value, ok := opt.WSOpts.Headers["Host"]; ok && value != "" {
				query.Set("ws-host", value)
			}
		}

		if opt.WSOpts.Path != "" {
			query.Set("ws-path", opt.WSOpts.Path)
		}

		if opt.RealityOpts.ShortID != "" {
			query.Set("short-id", opt.RealityOpts.ShortID)
		}

		if opt.RealityOpts.PublicKey != "" {
			query.Set("public-key", opt.RealityOpts.PublicKey)
		}

		u.RawQuery = urlx.SortQuery(query).Encode()

		p.vanillaLink = u.String()
		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

	case constant.Hysteria:
		var opt *outbound.HysteriaOption
		switch x := o.(type) {
		case outbound.HysteriaOption:
			opt = &x
		case *outbound.HysteriaOption:
			opt = x
		default:
			return fmt.Errorf("invalid option type: %T", o)
		}

		u := &url.URL{
			Scheme: "hysteria",
		}

		query := u.Query()

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.AuthString != "" {
			u.User = url.User(opt.AuthString)
		}

		if opt.Protocol != "" {
			query.Set("protocol", opt.Protocol)
		}

		if opt.ObfsProtocol != "" {
			query.Set("obfs-protocol", opt.ObfsProtocol)
		}

		if opt.Obfs != "" {
			query.Set("obfs", opt.Obfs)
		}

		if opt.SNI != "" {
			query.Set("sni", opt.SNI)
		}

		if opt.CustomCAString != "" {
			query.Set("ca", opt.CustomCAString)
		}

		if opt.ReceiveWindowConn != 0 {
			query.Set("recv-window-conn", strconv.Itoa(opt.ReceiveWindowConn))
		}

		if opt.ReceiveWindow != 0 {
			query.Set("recv-window", strconv.Itoa(opt.ReceiveWindow))
		}

		if opt.DisableMTUDiscovery {
			query.Set("disable-mtu-discovery", "true")
		}

		if opt.FastOpen {
			query.Set("fast-open", "true")
		}

		if opt.HopInterval != 0 {
			query.Set("hop-interval", strconv.Itoa(opt.HopInterval))
		}

		u.RawQuery = urlx.SortQuery(query).Encode()

		p.vanillaLink = u.String()
		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

	case constant.WireGuard:
		var opt *outbound.WireGuardOption
		switch x := o.(type) {
		case outbound.WireGuardOption:
			opt = &x
		case *outbound.WireGuardOption:
			opt = x
		default:
			return fmt.Errorf("invalid option type: %T", o)
		}

		u := &url.URL{
			Scheme: "wireguard",
		}

		query := u.Query()

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.PublicKey != "" {
			u.User = url.User(opt.PublicKey)
		}

		u.RawQuery = urlx.SortQuery(query).Encode()

		p.vanillaLink = u.String()
		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

	case constant.Tuic:
		var opt *outbound.TuicOption
		switch x := o.(type) {
		case outbound.TuicOption:
			opt = &x
		case *outbound.TuicOption:
			opt = x
		default:
			return fmt.Errorf("invalid option type: %T", o)
		}

		u := &url.URL{
			Scheme: "tuic",
		}

		query := u.Query()

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.UUID != "" && opt.Password != "" {
			u.User = url.UserPassword(opt.UUID, opt.Password)
		} else if opt.UUID != "" {
			u.User = url.User(opt.UUID)
		} else if opt.Password != "" {
			u.User = url.UserPassword("", opt.Password)
		}

		if opt.ReduceRtt {
			query.Set("reduce-rtt", "true")
		}

		if opt.UDPOverStream {
			query.Set("udp-over-stream", "true")
		}

		if opt.UdpRelayMode != "" {
			query.Set("udp-relay-mode", opt.UdpRelayMode)
		}

		if opt.CongestionController != "" {
			query.Set("congestion-controller", opt.CongestionController)
		}

		if opt.DisableSni {
			query.Set("disable-sniffing", "true")
		}

		if opt.FastOpen {
			query.Set("fast-open", "true")
		}

		if opt.CWND != 0 {
			query.Set("cwnd", strconv.Itoa(opt.CWND))
		}

		if opt.CustomCAString != "" {
			query.Set("ca", opt.CustomCAString)
		}

		if opt.SNI != "" {
			query.Set("sni", opt.SNI)
		}

		if opt.ReceiveWindow != 0 {
			query.Set("recv-window", strconv.Itoa(opt.ReceiveWindow))
		}

		if opt.ReceiveWindowConn != 0 {
			query.Set("recv-window-conn", strconv.Itoa(opt.ReceiveWindowConn))
		}

		if opt.UDPOverStream {
			query.Set("udp-over-stream", "true")
		}

		if opt.DisableMTUDiscovery {
			query.Set("disable-mtu-discovery", "true")
		}

		if opt.UDPOverStreamVersion != 0 {
			query.Set("udp-over-stream-version", strconv.Itoa(opt.UDPOverStreamVersion))
		}

		u.RawQuery = urlx.SortQuery(query).Encode()

		p.vanillaLink = u.String()
		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(p.vanillaLink)))

	default:
		return fmt.Errorf("unsupported protocol: %s", p.Type())
	}

	return nil
}

type simpleObfsOption struct {
	Mode string `obfs:"mode,omitempty"`
	Host string `obfs:"host,omitempty"`
}

type v2rayObfsOption struct {
	Mode           string            `obfs:"mode"`
	Host           string            `obfs:"host,omitempty"`
	Path           string            `obfs:"path,omitempty"`
	TLS            bool              `obfs:"tls,omitempty"`
	Headers        map[string]string `obfs:"headers,omitempty"`
	SkipCertVerify bool              `obfs:"skip-cert-verify,omitempty"`
	Mux            bool              `obfs:"mux,omitempty"`
}
