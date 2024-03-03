package adapter

import (
	"errors"
	"github.com/elliotchance/pie/v2"
	"github.com/ice-cream-heaven/log"
	"github.com/ice-cream-heaven/utils/anyx"
	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/adapter/outbound"
	"net/url"
	"strconv"
	"strings"
)

func ParseLink(s string) (*Adapter, error) {
	s = strings.TrimSuffix(s, "\n")
	s = strings.TrimSuffix(s, "\r")

	if strings.TrimSpace(s) == "" {
		return nil, ErrEmptyDate
	}

	u, err := url.Parse(s)
	if err != nil {
		log.Errorf("err:%v", err)
		return ParseClashWithYaml([]byte(s))
	}

	switch u.Scheme {
	case "http", "https":
		return ParseLinkHttp(s)
	case "socket4", "socket5", "socket", "socks4", "socks5":
		return ParseLinkSocket5(s)
	case "trojan", "trojan-go":
		return ParseLinkTrojan(s)
	case "vless":
		return ParseLinkVless(s)
	case "vmess":
		return ParseLinkVmess(s)
	case "ss", "shadowsocks":
		return ParseLinkSS(s)
	case "ssr":
		return ParseLinkSSR(s)
	case "hysteria", "hy":
		return ParseHysteria(s)
	case "hysteria2", "hy2":
		return ParseHysteria2(s)
	default:
		log.Debugf("unsupport v2ray scheme:%s(%s)", u.Scheme, s)
		return nil, ErrUnsupportedType
	}
}

func ParseHysteria(s string) (*Adapter, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, ErrParseLink
	}

	//hysteria://152.69.208.195:18209?protocol=udp&auth=d50157&peer=www.bing.com&insecure=true&upmbps=10&downmbps=50&alpn=h3#hysteria-ygkkk

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	//up, err := strconv.Atoi(u.Query().Get("upmbps"))
	//if err != nil {
	//	log.Errorf("err:%v", err)
	//	return nil, err
	//}
	//
	//down, err := strconv.Atoi(u.Query().Get("downmbps"))
	//if err != nil {
	//	log.Errorf("err:%v", err)
	//	return nil, err
	//}

	var alpn []string
	switch u.Query().Get("alpn") {
	case "h3":
		alpn = append(alpn, "h3")
	}

	opt := outbound.HysteriaOption{
		BasicOption:         outbound.BasicOption{},
		Name:                u.Fragment,
		Server:              u.Hostname(),
		Port:                port,
		Ports:               "",
		Protocol:            u.Query().Get("protocol"),
		ObfsProtocol:        "",
		Up:                  "M",
		UpSpeed:             10,
		Down:                "M",
		DownSpeed:           50,
		Auth:                "",
		AuthString:          u.Query().Get("auth"),
		Obfs:                "",
		SNI:                 u.Query().Get("peer"),
		SkipCertVerify:      false,
		Fingerprint:         "",
		ALPN:                alpn,
		CustomCA:            "",
		CustomCAString:      "",
		ReceiveWindowConn:   0,
		ReceiveWindow:       0,
		DisableMTUDiscovery: false,
		FastOpen:            false,
		HopInterval:         0,
	}

	log.Debugf("hysteria opt:%+v", opt)
	at, err := outbound.NewHysteria(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewAdapter(adapter.NewProxy(at), opt)
}

func ParseHysteria2(s string) (*Adapter, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, ErrParseLink
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	var alpn []string
	switch u.Query().Get("alpn") {
	case "h3":
		alpn = append(alpn, "h3")
	}

	opt := outbound.Hysteria2Option{
		BasicOption: outbound.BasicOption{},
		Name:        u.Fragment,
		Server:      u.Hostname(),
		Port:        port,
		Password:    u.User.String(),
		Obfs: func() string {
			switch u.Query().Get("obfs") {
			case "none":
				return ""
			default:
				return u.Query().Get("obfs")
			}
		}(),
		ObfsPassword:   u.Query().Get("obfs-psssowrd"),
		SNI:            u.Query().Get("peer"),
		SkipCertVerify: false,
		Fingerprint:    "",
		ALPN:           alpn,
		CustomCA:       "",
		CustomCAString: "",
		CWND:           0,
	}

	log.Debugf("hysteria2 opt:%+v", opt)
	at, err := outbound.NewHysteria2(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewAdapter(adapter.NewProxy(at), opt)
}

func ParseLinkSSR(s string) (*Adapter, error) {
	urlStr := Base64Decode(strings.TrimPrefix(s, "ssr://"))
	params := strings.Split(urlStr, `:`)
	if len(params) != 6 {
		//
		return nil, errors.New("invalid ssr url")
	}

	port, _ := strconv.Atoi(params[1])

	protocol := params[2]
	obfs := params[4]
	cipher := params[3]

	suffix := strings.Split(params[5], "/?")
	if len(suffix) != 2 {
		return nil, errors.New("invalid ssr url")
	}

	password := Base64Decode(suffix[0])

	m, err := url.ParseQuery(suffix[1])
	if err != nil {
		return nil, ErrParseLink
	}

	var obfsParam, protocolParam, name string
	for k, v := range m {
		de := Base64Decode(v[0])
		switch k {
		case "obfsparam":
			obfsParam = de
		case "protoparam":
			protocolParam = de
		case "remarks":
			name = de
		}
	}

	if protocol == "origin" && obfs == "plain" {
		switch cipher {
		case "aes-128-gcm", "aes-192-gcm", "aes-256-gcm",
			"aes-128-cfb", "aes-192-cfb", "aes-256-cfb",
			"aes-128-ctr", "aes-192-ctr", "aes-256-ctr",
			"rc4-md5", "chacha20", "chacha20-ietf", "xchacha20",
			"chacha20-ietf-poly1305", "xchacha20-ietf-poly1305":
			// opt := outbound.ShadowSocksOption{
			// 	BasicOption: outbound.BasicOption{},
			// 	Name:        name,
			// 	Server:      params[0],
			// 	Port:        port,
			// 	Password:    password,
			// 	Cipher:      cipher,
			// 	UDP:         false,
			// 	Plugin:      "",
			// 	PluginOpts:  nil,
			// }
			return nil, errors.New("invalid ssr url")
		}
	}

	opt := outbound.ShadowSocksROption{
		BasicOption:   outbound.BasicOption{},
		Name:          name,
		Server:        params[0],
		Port:          port,
		Password:      password,
		Cipher:        cipher,
		Obfs:          obfs,
		ObfsParam:     obfsParam,
		Protocol:      protocol,
		ProtocolParam: protocolParam,
		UDP:           true,
	}

	log.Debugf("ssr opt:%+v", opt)

	at, err := outbound.NewShadowSocksR(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewAdapter(adapter.NewProxy(at), opt)
}

func ParseLinkSS(s string) (*Adapter, error) {
	var urlStr string
	var fragment string
	bu, err := url.Parse(s)
	if err != nil {
		log.Errorf("err:%v", err)
		urlStr = "ss://" + Base64Decode(strings.TrimPrefix(s, "ss://"))
	} else {
		fragment = bu.Fragment
		bu.Fragment = ""
		urlStr = "ss://" + Base64Decode(strings.TrimPrefix(bu.String(), "ss://"))
	}

	log.Debugf("urlStr:%s", urlStr)

	u, err := url.Parse(urlStr)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, ErrParseLink
	}

	if fragment == "" {
		fragment = u.Fragment
	}

	port, _ := strconv.Atoi(u.Port())

	var cipher, password string
	// 对username解析
	userStr := Base64Decode(u.User.String())

	log.Debugf("userStr:%s", userStr)

	userSplit := strings.Split(userStr, ":")
	if len(userSplit) > 0 {
		cipher = userSplit[0]
	}

	if len(userSplit) > 1 {
		password = userSplit[1]
	}

	opt := outbound.ShadowSocksOption{
		BasicOption: outbound.BasicOption{},
		Name:        fragment,
		Server:      u.Hostname(),
		Port:        port,
		Password:    password,
		Cipher:      cipher,
		UDP:         true,
		Plugin:      "",
		PluginOpts:  nil,
	}

	log.Debugf("ss opt:%+v", opt)

	at, err := outbound.NewShadowSocks(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewAdapter(adapter.NewProxy(at), opt)
}

func ParseLinkHttp(s string) (*Adapter, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, ErrParseLink
	}

	port, _ := strconv.Atoi(u.Port())

	opt := outbound.HttpOption{
		BasicOption:    outbound.BasicOption{},
		Name:           u.Fragment,
		Server:         u.Hostname(),
		Port:           port,
		SkipCertVerify: true,
	}

	if u.User != nil {
		opt.UserName = u.User.Username()
		opt.Password, _ = u.User.Password()
	}

	if u.Scheme == "https" {
		opt.TLS = true
	}

	log.Debugf("http opt:%+v", opt)

	at, err := outbound.NewHttp(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewAdapter(adapter.NewProxy(at), opt)
}

func ParseLinkSocket5(s string) (*Adapter, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, ErrParseLink
	}

	port, _ := strconv.Atoi(u.Port())

	opt := outbound.Socks5Option{
		BasicOption:    outbound.BasicOption{},
		Name:           u.Fragment,
		Server:         u.Hostname(),
		Port:           port,
		SkipCertVerify: true,
	}

	if u.User != nil {
		opt.UserName = u.User.Username()
		opt.Password, _ = u.User.Password()
	}

	log.Debugf("socket opt:%+v", opt)

	at, err := outbound.NewSocks5(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewAdapter(adapter.NewProxy(at), opt)
}

func ParseLinkTrojan(s string) (*Adapter, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, ErrParseLink
	}

	sni := u.Query().Get("sni")
	if sni == "" {
		sni = u.Hostname()
	}

	// 处理ws
	var wsOpt outbound.WSOptions
	if u.Query().Get("ws") != "1" || strings.ToLower(u.Query().Get("ws")) == "true" {
		wsOpt = outbound.WSOptions{
			Path:                u.Query().Get("wspath"),
			Headers:             nil,
			MaxEarlyData:        0,
			EarlyDataHeaderName: "",
		}
	}

	transformType := u.Query().Get("type")
	transformType, _ = url.QueryUnescape(transformType)

	var alpn []string
	if transformType == "h2" {
		alpn = append(alpn, transformType)
	}

	for _, val := range strings.Split(u.Query().Get("alpn"), ",") {
		if val == "" {
			continue
		}
		alpn = append(alpn, val)
	}

	alpn = pie.Unique(alpn)

	port, _ := strconv.Atoi(u.Port())

	opt := outbound.TrojanOption{
		BasicOption: outbound.BasicOption{
			Interface:   "",
			RoutingMark: 0,
		},
		Name:           u.Fragment,
		Server:         u.Hostname(),
		Password:       u.User.String(),
		Port:           port,
		ALPN:           alpn,
		SNI:            sni,
		SkipCertVerify: true,
		UDP:            true,
		Network:        transformType,
		GrpcOpts: outbound.GrpcOptions{
			GrpcServiceName: "",
		},
		WSOpts: wsOpt,
	}

	log.Debugf("trojan opt:%+v", opt)

	at, err := outbound.NewTrojan(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewAdapter(adapter.NewProxy(at), opt)
}

func ParseLinkVless(s string) (*Adapter, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, ErrParseLink
	}

	sni := u.Query().Get("sni")
	if sni == "" {
		sni = u.Hostname()
	}

	transformType := u.Query().Get("type")
	transformType, _ = url.QueryUnescape(transformType)

	var alpn []string
	if transformType == "h2" {
		alpn = append(alpn, transformType)
	}

	port, _ := strconv.Atoi(u.Port())

	opt := outbound.VlessOption{
		Name:   u.Fragment,
		Server: u.Hostname(),
		Port:   port,
		UUID:   u.User.String(),
		UDP:    true,
		TLS: func() bool {
			return u.Query().Get("security") == "tls"
		}(),
		Network: func() string {
			if u.Query().Get("type") != "" {
				return u.Query().Get("type")
			}

			return "tcp"
		}(),
		WSPath:         u.Query().Get("path"),
		WSHeaders:      nil,
		SkipCertVerify: true,
		ServerName: func() string {
			if u.Query().Get("host") != "" {
				return u.Query().Get("host")
			}
			return u.Query().Get("sni")
		}(),
		Flow: u.Query().Get("flow"),
	}

	log.Debugf("vless opt:%+v", opt)

	at, err := outbound.NewVless(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewAdapter(adapter.NewProxy(at), opt)
}

func ParseLinkVmess(s string) (*Adapter, error) {
	var opt outbound.VmessOption
	base64Str := Base64Decode(strings.TrimPrefix(s, "vmess://"))
	m, err := anyx.NewMapWithJson([]byte(base64Str))
	if err != nil {
		log.Debugf("err:%v", err)

		u, err := url.Parse(s)
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, ErrParseLink
		} else {
			u.Host = Base64Decode(u.Host)
		}

		urlStr, err := url.QueryUnescape(u.String())
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		u, err = url.Parse(urlStr)
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, ErrParseLink
		}

		var wsOpts outbound.WSOptions
		var network string
		switch u.Query().Get("obfs") {
		case "websocket", "ws":
			network = "ws"
			wsOpts = outbound.WSOptions{
				Path: u.Query().Get("path"),
				Headers: map[string]string{
					"Host": u.Query().Get("peer"),
				},
				MaxEarlyData:        0,
				EarlyDataHeaderName: "",
			}
		}

		opt = outbound.VmessOption{
			BasicOption: outbound.BasicOption{},
			Name:        u.Query().Get("remarks"),
			Server:      u.Hostname(),
			Port:        anyx.ToInt(u.Port()),
			UUID: func() string {
				pwd, ok := u.User.Password()
				if ok {
					return pwd
				}
				return u.User.Username()
			}(),
			AlterID: 0,
			Cipher: func() string {
				_, ok := u.User.Password()
				if ok {
					return u.User.Username()
				}
				return ""
			}(),
			UDP:            true,
			Network:        network,
			SkipCertVerify: true,
			WSOpts:         wsOpts,
		}

		tls := u.Query().Get("tls")
		switch tls {
		case "none":
		default:
			opt.TLS = anyx.ToBool(tls)
		}

	} else {
		opt = outbound.VmessOption{
			BasicOption: outbound.BasicOption{},
			Name:        m.GetString("ps"),
			Server: func() string {
				if m.GetString("add") != "" {
					return m.GetString("add")
				}

				return m.GetString("host")
			}(),
			Port:    m.GetInt("port"),
			UUID:    m.GetString("id"),
			AlterID: m.GetInt("aid"),
			Cipher: func() string {
				if m.GetString("scy") != "" {
					return m.GetString("scy")
				}

				return "auto"
			}(),
			UDP:     true,
			Network: m.GetString("net"),
			TLS: func() bool {
				val, err := m.Get("tls")
				if err != nil {
					return false
				}

				switch anyx.CheckValueType(val) {
				case anyx.ValueString:
					switch m.GetString("tls") {
					case "none", "":
						return false
					default:
						return true
					}
				case anyx.ValueBool:
					return m.GetBool("tls")
				default:
					return false
				}
			}(),
			SkipCertVerify: true,
			ServerName:     m.GetString("sni"),
			HTTPOpts:       outbound.HTTPOptions{},
			HTTP2Opts:      outbound.HTTP2Options{},
			GrpcOpts:       outbound.GrpcOptions{},
			WSOpts:         outbound.WSOptions{},
		}

		switch m.GetString("net") {
		case "ws":
			opt.WSOpts = outbound.WSOptions{
				Path: m.GetString("path"),
				Headers: map[string]string{
					"Host": m.GetString("host"),
				},
				MaxEarlyData:        0,
				EarlyDataHeaderName: "",
			}
		case "grpc":
			opt.GrpcOpts = outbound.GrpcOptions{
				GrpcServiceName: m.GetString("path"),
			}
			opt.ServerName = m.GetString("host")
		case "h2":
			opt.HTTP2Opts = outbound.HTTP2Options{
				Host: []string{
					m.GetString("host"),
				},
				Path: m.GetString("path"),
			}
		}

	}

	if opt.Cipher == "zero" {
		opt.Cipher = "none"
	}

	log.Debugf("vmess opt:%+v", opt)

	at, err := outbound.NewVmess(opt)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewAdapter(adapter.NewProxy(at), opt)
}
