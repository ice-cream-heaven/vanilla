package adapter_test

import (
	"github.com/ice-cream-heaven/vanilla/adapter"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestHttp(t *testing.T) {
	request, err := http.NewRequest(http.MethodGet, "https://www.baidu.com", nil)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "+
		"(KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http",
				Host:   "127.0.0.1:7891",
			}),
			//TLSClientConfig: &tls.Config{
			//	InsecureSkipVerify: true,
			//},
		},
	}
	defer client.CloseIdleConnections()

	resp, err := client.Do(request)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	t.Logf("status code :%v", resp.StatusCode)
	t.Log(string(buf))
}

func TestMitm(t *testing.T) {
	request, err := http.NewRequest(http.MethodGet, "https://www.baidu.com", nil)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http",
				Host:   "127.0.0.1:7891",
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

func TestClash(t *testing.T) {
	var m map[string]any
	err := yaml.Unmarshal([]byte(`{name: b89a2b0077ac|c, server: 140.99.94.19, port: 443, type: vmess, uuid: 418048af-a293-4b99-9b0c-98ca3580dd24, alterId: 64, cipher: auto, tls: true, skip-cert-verify: true, servername: www.89184508.xyz, network: ws, ws-opts: {path: /path/211734121312}, udp: true}`), &m)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	a, err := adapter.ParseClash(m)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	t.Log(a.ToClash())
	t.Log(a.UniqueId())

	t.Log(a.Addr())
}
