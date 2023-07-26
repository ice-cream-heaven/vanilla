package dns

import (
	"context"
	"net"
)

type TcpClient struct {
	client *net.Resolver
}

func (p *TcpClient) LookupIP(host string) (ips []net.IP, err error) {
	return p.client.LookupIP(context.Background(), "ip", host)
}

func (p *TcpClient) LookupIPv4(host string) (ips []net.IP, err error) {
	return p.client.LookupIP(context.Background(), "ip4", host)
}

func (p *TcpClient) LookupIPv6(host string) (ips []net.IP, err error) {
	return p.client.LookupIP(context.Background(), "ip6", host)
}

func NewTcpClient(addr string, dial Dial) *TcpClient {
	host, port, _ := net.SplitHostPort(addr)
	if port == "" {
		host = addr
		port = "53"
	}

	if dial == nil {
		dial = DefaultDial
	}

	return &TcpClient{
		client: &net.Resolver{
			PreferGo:     true,
			StrictErrors: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return dial("tcp", net.JoinHostPort(host, port))
			},
		},
	}
}
