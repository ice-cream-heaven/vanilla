package adapter

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"github.com/Dreamacro/clash/adapter/outbound"
	"github.com/Dreamacro/clash/constant"
	"github.com/elliotchance/pie/v2"
	"github.com/ice-cream-heaven/log"
	"github.com/ice-cream-heaven/utils/urlx"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func (p *Adapter) UniqueId() string {
	return p.uniqueId
}

func (p *Adapter) ShortId() string {
	if len(p.uniqueId) < 8 {
		return p.uniqueId
	}
	return p.uniqueId[:8]
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

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.Password != "" {
			u.User = url.User(opt.Password)
		}

		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(u.String())))

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

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.Password != "" {
			u.User = url.User(opt.Password)
		}

		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(u.String())))

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

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.Psk != "" {
			u.User = url.User(opt.Psk)
		}

		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(u.String())))

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

		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(u.String())))

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

		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(u.String())))

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

		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(u.String())))

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

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.UUID != "" {
			u.User = url.User(opt.UUID)
		}

		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(u.String())))

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

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.Password != "" {
			u.User = url.User(opt.Password)
		}

		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(u.String())))

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

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.AuthString != "" {
			u.User = url.User(opt.AuthString)
		}

		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(u.String())))

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

		if opt.Port > 0 {
			u.Host = net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))
		} else {
			u.Host = opt.Server
		}

		if opt.PublicKey != "" {
			u.User = url.User(opt.PublicKey)
		}

		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(u.String())))

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

		p.uniqueId = fmt.Sprintf("%x", sha512.Sum512([]byte(u.String())))

	default:
		return fmt.Errorf("unsupported protocol: %s", p.Type())
	}

	return nil
}
