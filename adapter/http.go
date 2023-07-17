package adapter

import (
	"context"
	"github.com/Dreamacro/clash/constant"
	"github.com/go-resty/resty/v2"
	"net"
	"net/http"
	"time"
)

func (p *Adapter) HttpDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	meta := &constant.Metadata{
		NetWork: constant.TCP,
	}
	meta.Host, meta.DstPort, _ = net.SplitHostPort(addr)
	return p.DialContext(ctx, meta)
}

func (p *Adapter) Transport() http.RoundTripper {
	return &http.Transport{
		DialContext: p.HttpDialContext,
	}
}

func (p *Adapter) GetClient() *http.Client {
	return &http.Client{
		Transport: p.Transport(),
	}
}

func (p *Adapter) GetClientWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: p.HttpDialContext,
		},
		Timeout: timeout,
	}
}

func (p *Adapter) R() *resty.Request {
	return p.client.R()
}
