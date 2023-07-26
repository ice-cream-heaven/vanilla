package dns

import (
	"errors"
	"github.com/ice-cream-heaven/log"
	"net"
	"net/url"
)

type Resolver interface {
	LookupIP(host string) (ips []net.IP, err error)
	LookupIPv4(host string) (ips []net.IP, err error)
	LookupIPv6(host string) (ips []net.IP, err error)
}

type Dial func(network, address string) (net.Conn, error)

func MustNewResolver(addr string) Resolver {
	res, err := NewResolverWithProxy(addr, nil)
	if err != nil {
		log.Panicf("err:%v", err)
	}

	return res
}

func NewResolver(addr string) (Resolver, error) {
	return NewResolverWithProxy(addr, nil)
}

func NewResolverWithProxy(addr string, dial Dial) (Resolver, error) {
	u, err := url.Parse(addr)
	if err != nil {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		return NewUdpClient(net.JoinHostPort(host, port), dial), nil
	}

	switch u.Scheme {
	case "udp":
		return NewUdpClient(u.Host, dial), nil
	case "tcp":
		return NewTcpClient(u.Host, dial), nil
	case "":
		return NewUdpClient(u.String(), dial), nil
	case "tls", "tpc-tls":
		return NewDotClient(u.Host, dial)
	case "https":
		return NewDohClient(u.String(), dial)
	default:
		return nil, errors.New("invalid dns resolver")
	}
}
