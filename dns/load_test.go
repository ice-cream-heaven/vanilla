package dns_test

import (
	"github.com/ice-cream-heaven/vanilla/dns"
	"net"
	"testing"
)

func TestUdp(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{
			addr: "119.29.29.29",
		},
		{
			addr: "119.29.29.29:53",
		},
		{
			addr: "udp://119.29.29.29",
		},
		{
			addr: "udp://119.29.29.29:53",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dns.NewResolver(tt.addr)
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			ips, err := got.LookupIP("baidu.com")
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			if len(ips) == 0 {
				t.Errorf("ips is empty")
				return
			}

			t.Logf("ips:%v", ips)
		})
	}
}

func TestTcp(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{
			addr: "tcp://119.29.29.29",
		},
		{
			addr: "tcp://119.29.29.29:53",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dns.NewResolverWithProxy(tt.addr, func(network, address string) (net.Conn, error) {
				return net.Dial(network, address)
			})
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			ips, err := got.LookupIP("baidu.com")
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			if len(ips) == 0 {
				t.Errorf("ips is empty")
				return
			}

			t.Logf("ips:%v", ips)
		})
	}
}

func TestDot(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{
			addr: "tls://1.12.12.12",
		},
		{
			addr: "tls://dot.pub",
		},
		{
			addr: "tls://1.12.12.12:853",
		},
		{
			addr: "tls://dot.pub:853",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dns.NewResolver(tt.addr)
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			ips, err := got.LookupIP("baidu.com")
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			if len(ips) == 0 {
				t.Errorf("ips is empty")
				return
			}

			t.Logf("ips:%v", ips)
		})
	}
}

func TestDoh(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{
			addr: "https://1.12.12.12/dns-query",
		},
		{
			addr: "https://doh.pub/dns-query",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dns.NewResolver(tt.addr)
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			ips, err := got.LookupIP("baidu.com")
			if err != nil {
				t.Errorf("err:%v", err)
				return
			}

			if len(ips) == 0 {
				t.Errorf("ips is empty")
				return
			}

			t.Logf("ips:%v", ips)
		})
	}
}
