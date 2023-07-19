package adapter

import (
	"errors"
	"github.com/Dreamacro/clash/adapter"
	"github.com/Dreamacro/clash/common/structure"
	"github.com/ice-cream-heaven/log"
	"strings"
)

var (
	ErrUnsupportedType = errors.New("unsupported type")
	ErrEmptyDate       = errors.New("empty date")

	decoder = structure.NewDecoder(
		structure.Option{
			TagName:          "proxy",
			WeaklyTypedInput: true,
		},
	)
)

func ParseClash(m map[string]any) (*Adapter, error) {
	p, err := adapter.ParseProxy(m)
	if err != nil {
		log.Debugf("err:%v", err)

		if strings.Contains(err.Error(), "unsupport proxy type") {
			return nil, ErrUnsupportedType
		}

		return nil, err
	}
	return NewAdapter(p, m)
}
