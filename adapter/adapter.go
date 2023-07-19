package adapter

import (
	"github.com/Dreamacro/clash/constant"
	"github.com/bytedance/sonic"
	"github.com/go-resty/resty/v2"
	"time"
)

type Adapter struct {
	constant.ProxyAdapter

	uniqueId string

	client *resty.Client

	opt map[string]any
}

func NewAdapter(c constant.ProxyAdapter, o map[string]any) (*Adapter, error) {
	p := &Adapter{
		opt:          o,
		ProxyAdapter: c,
		client: resty.New().
			SetTimeout(time.Minute).
			SetRetryWaitTime(time.Second).
			SetRetryCount(3).
			SetRedirectPolicy(resty.FlexibleRedirectPolicy(10)),
	}

	err := p.updateUniqueId()
	if err != nil {
		return nil, err
	}

	{
		p.client.JSONUnmarshal = sonic.Unmarshal
		p.client.JSONMarshal = sonic.Marshal
		p.client.SetTransport(p.Transport())
	}

	return p, nil
}

func (p *Adapter) ToClash() map[string]any {
	mapping := p.opt
	mapping["name"] = p.Name()

	if p.SupportUDP() {
		mapping["udp"] = true
	} else {
		delete(mapping, "udp")
	}

	if p.SupportXUDP() {
		mapping["xudp"] = true
	} else {
		delete(mapping, "xudp")
	}

	if p.SupportTFO() {
		mapping["tfo"] = true
	} else {
		delete(mapping, "tfo")
	}

	return mapping
}
