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
}

func NewAdapter(c constant.ProxyAdapter, o any) (*Adapter, error) {
	p := &Adapter{
		ProxyAdapter: c,
		client: resty.New().
			SetTimeout(time.Minute).
			SetRetryWaitTime(time.Second).
			SetRetryCount(3).
			SetRedirectPolicy(resty.FlexibleRedirectPolicy(10)),
	}

	err := p.updateUniqueId(c.Type(), o)
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
	inner, err := p.ProxyAdapter.MarshalJSON()
	if err != nil {
		return map[string]any{}
	}

	mapping := map[string]any{}
	_ = sonic.Unmarshal(inner, &mapping)

	mapping["name"] = p.Name()
	mapping["udp"] = p.SupportUDP()
	mapping["xudp"] = p.SupportXUDP()
	mapping["tfo"] = p.SupportTFO()

	return nil
}
