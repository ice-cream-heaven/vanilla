package adapter

import (
	"github.com/metacubex/mihomo/adapter/outbound"
	"github.com/metacubex/mihomo/constant"
)

type ProxyDirect struct {
	*outbound.Direct
}

func NewProxyDirect() constant.ProxyAdapter {
	return &ProxyDirect{
		Direct: outbound.NewDirect(),
	}
}

func NewDirect() (*Adapter, error) {
	return NewAdapter(NewProxyDirect(), map[string]any{
		"type": "direct",
		"name": "direct",
	})
}
