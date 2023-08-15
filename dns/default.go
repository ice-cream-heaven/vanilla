package dns

import (
	"errors"
	"github.com/Dreamacro/clash/common/cache"
	"github.com/elliotchance/pie/v2"
	"github.com/ice-cream-heaven/log"
	"github.com/ice-cream-heaven/utils/wait"
	"net"
	"sync"
	"time"
)

var DefaultDial = func(network, address string) (net.Conn, error) {
	return net.DialTimeout(network, address, time.Second*3)
}

type defaultResolver struct {
	resolvers []Resolver
	cache     *cache.LruCache[string, []net.IP]
}

var (
	DefaultResolver = newDefaultResolver()
)

func newDefaultResolver() *defaultResolver {
	return &defaultResolver{
		cache: cache.New[string, []net.IP](
			cache.WithAge[string, []net.IP](60),
			cache.WithSize[string, []net.IP](50),
			cache.WithUpdateAgeOnGet[string, []net.IP](),
		),
	}
}

var ErrEmptyResponse = errors.New("empty response")

func (p *defaultResolver) LookupHost(host string) (ip net.IP, err error) {
	ips, err := p.LookupIPv4(host)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	if len(ips) == 0 {
		return nil, ErrEmptyResponse
	}

	return ips[0], nil
}

func (p *defaultResolver) LookupIP(host string) (ips []net.IP, err error) {
	var ok bool
	ips, ok = p.cache.Get("ip:" + host)
	if ok {
		return ips, nil
	}

	ipExisted := map[string]bool{}

	for _, resolver := range p.resolvers {
		_ips, err := resolver.LookupIP(host)
		if err != nil {
			log.Errorf("err:%v", err)
		}

		for _, ip := range _ips {
			if ipExisted[ip.String()] {
				continue
			}

			ips = append(ips, _ips...)
			ipExisted[ip.String()] = true
		}
	}

	if len(ips) == 0 {
		return nil, ErrEmptyResponse
	}

	p.cache.Set("ip:"+host, ips)

	return ips, nil
}

func (p *defaultResolver) LookupIPv4(host string) (ips []net.IP, err error) {
	var ok bool
	ips, ok = p.cache.Get("ip4:" + host)
	if ok {
		return ips, nil
	}

	ipExisted := map[string]bool{}

	for _, resolver := range p.resolvers {
		_ips, err := resolver.LookupIPv4(host)
		if err != nil {
			log.Errorf("err:%v", err)
		}

		for _, ip := range _ips {
			if ipExisted[ip.String()] {
				continue
			}

			ips = append(ips, _ips...)
			ipExisted[ip.String()] = true
		}
	}

	if len(ips) == 0 {
		return nil, ErrEmptyResponse
	}

	p.cache.Set("ip4:"+host, ips)

	return ips, nil
}

func (p *defaultResolver) LookupIPv6(host string) (ips []net.IP, err error) {
	var ok bool
	ips, ok = p.cache.Get("ip6:" + host)
	if ok {
		return ips, nil
	}

	ipExisted := map[string]bool{}

	for _, resolver := range p.resolvers {
		_ips, err := resolver.LookupIPv6(host)
		if err != nil {
			log.Errorf("err:%v", err)
		}

		for _, ip := range _ips {
			if ipExisted[ip.String()] {
				continue
			}

			ips = append(ips, _ips...)
			ipExisted[ip.String()] = true
		}
	}

	if len(ips) == 0 {
		return nil, ErrEmptyResponse
	}

	p.cache.Set("ip6:"+host, ips)

	return ips, nil
}

func (p *defaultResolver) AddResolver(resolver ...Resolver) {
	p.resolvers = append(p.resolvers, resolver...)
}

func (p *defaultResolver) QueryA(host string) map[string][]net.IP {
	var lock sync.Mutex
	m := map[string][]net.IP{}

	wait.Async(
		5,
		func(ms chan Resolver) {
			for _, resolver := range p.resolvers {
				ms <- resolver
			}
		},
		func(resolver Resolver) {
			_ips, err := resolver.LookupIP(host)
			if err != nil {
				log.Errorf("err:%v", err)
			}

			if len(_ips) == 0 {
				return
			}

			lock.Lock()
			m[resolver.Name()] = pie.SortUsing(_ips, func(a, b net.IP) bool {
				return a.String() < b.String()
			})
			lock.Unlock()
		},
	)

	return m
}
