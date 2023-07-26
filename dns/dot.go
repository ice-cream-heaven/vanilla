package dns

import (
	"context"
	"crypto/tls"
	"net"
)

type DotClient struct {
	client *net.Resolver
}

func (p *DotClient) LookupIP(host string) (ips []net.IP, err error) {
	return p.client.LookupIP(context.Background(), "ip", host)
}

func (p *DotClient) LookupIPv4(host string) (ips []net.IP, err error) {
	return p.client.LookupIP(context.Background(), "ip4", host)
}

func (p *DotClient) LookupIPv6(host string) (ips []net.IP, err error) {
	return p.client.LookupIP(context.Background(), "ip6", host)
}

func NewDotClient(addr string, dial Dial) (*DotClient, error) {
	client, err := NewDoTResolver(addr, dial)
	if err != nil {
		return nil, err
	}

	return &DotClient{
		client: client,
	}, nil
}

func NewDoTResolver(server string, dial Dial) (*net.Resolver, error) {
	host, port, err := net.SplitHostPort(server)
	if err != nil {
		port = "853"
	} else {
		server = host
	}

	var resolver = net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
	}

	if dial == nil {
		dial = DefaultDial
	}

	resolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		conn, err := dial("tcp", net.JoinHostPort(server, port))
		if err != nil {
			return nil, err
		}
		return tls.Client(conn, &tls.Config{
			ServerName: server,
		}), nil
	}

	return &resolver, nil
}
