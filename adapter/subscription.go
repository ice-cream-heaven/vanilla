package adapter

import (
	"github.com/ice-cream-heaven/log"
	"gopkg.in/yaml.v3"
	"strings"
)

func ParseSubscription(b []byte) (nodes []*Adapter) {
	// NOTE: clash
	{
		var c struct {
			Proxies []map[string]any `yaml:"proxies,omitempty"`
		}
		err := yaml.Unmarshal(b, &c)
		if err == nil {
			for _, m := range c.Proxies {
				node, err := ParseClash(m)
				if err != nil {
					log.Debugf("err:%v", err)
					continue
				}
				nodes = append(nodes, node)
			}
			return
		}
	}

	// NOTE: base64
	{
		for _, link := range strings.Split(Base64Decode(string(b)), "\n") {
			node, err := ParseLink(link)
			if err != nil {
				log.Debugf("err:%v", err)
				continue
			}

			nodes = append(nodes, node)
		}

		if len(nodes) > 0 {
			return
		}
	}

	return
}
