package adapter

import (
	"errors"
	"github.com/ice-cream-heaven/log"
	"github.com/ice-cream-heaven/utils/json"
	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/common/structure"
	"gopkg.in/yaml.v3"
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

func ParseClashWithJson(s []byte) (*Adapter, error) {
	var m map[string]any
	err := json.Unmarshal(s, &m)
	if err != nil {
		log.Debugf("err:%v", err)
		return nil, err
	}

	return ParseClash(m)
}

func ParseClashWithYaml(s []byte) (*Adapter, error) {
	var m map[string]any
	err := yaml.Unmarshal(s, &m)
	if err != nil {
		log.Debugf("err:%v", err)
		return nil, err
	}

	return ParseClash(m)
}
