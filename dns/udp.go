package dns

import (
	"context"
	"net"
)

type UdpClient struct {
	client *net.Resolver
}

func (p *UdpClient) LookupIP(host string) (ips []net.IP, err error) {
	return p.client.LookupIP(context.Background(), "ip", host)
}

func (p *UdpClient) LookupIPv4(host string) (ips []net.IP, err error) {
	return p.client.LookupIP(context.Background(), "ip4", host)
}

func (p *UdpClient) LookupIPv6(host string) (ips []net.IP, err error) {
	return p.client.LookupIP(context.Background(), "ip6", host)
}

func NewUdpClient(addr string, dial Dial) *UdpClient {
	host, port, _ := net.SplitHostPort(addr)
	if port == "" {
		host = addr
		port = "53"
	}

	if dial == nil {
		dial = DefaultDial
	}

	return &UdpClient{
		client: &net.Resolver{
			PreferGo:     true,
			StrictErrors: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return dial("udp", net.JoinHostPort(host, port))
			},
		},
	}
}
