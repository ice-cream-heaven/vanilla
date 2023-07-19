package adapter_test

import (
	"github.com/AdguardTeam/gomitmproxy"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"
)

func TestMitm(t *testing.T) {
	go func() {
		err := gomitmproxy.NewProxy(gomitmproxy.Config{
			ListenAddr: &net.TCPAddr{
				IP:   net.IPv4(0, 0, 0, 0),
				Port: 8080,
			},
			TLSConfig:      nil,
			MITMConfig:     nil,
			MITMExceptions: nil,
			OnConnect: func(session *gomitmproxy.Session, proto string, addr string) net.Conn {
				t.Logf("OnConnect session:%v proto:%v addr:%v", session.ID(), proto, addr)
				return nil
			},
			OnRequest: func(session *gomitmproxy.Session) (*http.Request, *http.Response) {
				t.Log(session.Request().URL)
				t.Log(session.Request().Header)
				return nil, nil
			},
			OnResponse: nil,
			OnError: func(session *gomitmproxy.Session, err error) {
				t.Errorf("%s err:%v", session.ID(), err)
			},
		}).Start()
		if err != nil {
			t.Errorf("err:%v", err)
			os.Exit(1)
		}
	}()

	request, err := http.NewRequest(http.MethodGet, "https://www.baidu.com", nil)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http",
				Host:   "127.0.0.1:8080",
			}),
		},
	}
	defer client.CloseIdleConnections()

	resp, err := client.Do(request)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	defer resp.Body.Close()

	t.Logf("status code :%v", resp.StatusCode)
}
