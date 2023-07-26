package dns

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

type DohClient struct {
	client *net.Resolver
}

func (p *DohClient) LookupIP(host string) (ips []net.IP, err error) {
	return p.client.LookupIP(context.Background(), "ip", host)
}

func (p *DohClient) LookupIPv4(host string) (ips []net.IP, err error) {
	return p.client.LookupIP(context.Background(), "ip4", host)
}

func (p *DohClient) LookupIPv6(host string) (ips []net.IP, err error) {
	return p.client.LookupIP(context.Background(), "ip6", host)
}

func NewDohClient(addr string, dial Dial) (*DohClient, error) {
	client, err := NewDoHResolver(addr, dial)
	if err != nil {
		return nil, err
	}

	return &DohClient{
		client: client,
	}, nil
}

func NewDoHResolver(uri string, dial Dial) (*net.Resolver, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	port := u.Port()
	if port == "" {
		port = "443"
	}

	if dial == nil {
		dial = DefaultDial
	}

	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        http.DefaultMaxIdleConnsPerHost,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			ForceAttemptHTTP2:   true,
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return dial("tcp", address)
			},
		},
	}

	var resolver = net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			conn := &dnsConn{}
			conn.roundTrip = dohRoundTrip(uri, &client)
			return conn, nil
		},
	}

	return &resolver, nil
}

func dohRoundTrip(uri string, client *http.Client) roundTripper {
	return func(ctx context.Context, msg string) (string, error) {
		req, err := http.NewRequestWithContext(ctx,
			http.MethodPost, uri, bytes.NewBufferString(msg))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/dns-message")

		res, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return "", errors.New(http.StatusText(res.StatusCode))
		}

		var b bytes.Buffer
		_, err = io.Copy(&b, res.Body)
		if err != nil {
			return "", err
		}

		return b.String(), nil
	}
}
