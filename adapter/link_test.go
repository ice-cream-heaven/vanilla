package adapter_test

import (
	"encoding/json"
	"github.com/ice-cream-heaven/vanilla/adapter"
	"testing"
)

func TestParseLink(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	tests := []struct {
		name string

		link string

		net bool

		wantErr bool
	}{
		{
			name: "http://223.5.5.5:7890",
		},
		{
			name: "https://223.5.5.5:7890",
		},
		{
			name: "socks5://223.5.5.5:7890",
		},
		{
			name: "vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIumfqeWbvTAx44CQ57K+6YCJ44CRIiwNCiAgImFkZCI6ICJjZG4ud3hmei50ayIsDQogICJwb3J0IjogIjgwIiwNCiAgImlkIjogIjU0YTkyODI0LWQ1NmMtNGVlNS1lZWVmLTc4NDg4ODU4OWYwMCIsDQogICJhaWQiOiAiNjQiLA0KICAic2N5IjogImF1dG8iLA0KICAibmV0IjogIndzIiwNCiAgInR5cGUiOiAibm9uZSIsDQogICJob3N0IjogImtyMDEuY2N0dnZpcC5jZiIsDQogICJwYXRoIjogIi9jY3R2MTMubTN1OCIsDQogICJ0bHMiOiAiIiwNCiAgInNuaSI6ICIiLA0KICAiYWxwbiI6ICIiLA0KICAiZnAiOiAiIg0KfQ==",
		},
		{
			name: "vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIumfqeWbvTAx44CQQ0YtQ0RO44CRIiwNCiAgImFkZCI6ICJrcjAxLmNjdHZ2aXAuY2YiLA0KICAicG9ydCI6ICI4MCIsDQogICJpZCI6ICI1NGE5MjgyNC1kNTZjLTRlZTUtZWVlZi03ODQ4ODg1ODlmMDAiLA0KICAiYWlkIjogIjY0IiwNCiAgInNjeSI6ICJhdXRvIiwNCiAgIm5ldCI6ICJ3cyIsDQogICJ0eXBlIjogIm5vbmUiLA0KICAiaG9zdCI6ICIiLA0KICAicGF0aCI6ICIvY2N0djEzLm0zdTgiLA0KICAidGxzIjogIiIsDQogICJzbmkiOiAiIiwNCiAgImFscG4iOiAiIiwNCiAgImZwIjogIiINCn0=",
		},
		{
			name: "vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIvCfh7rwn4e4IOe+juWbvTA244CQ57K+6YCJ44CRIiwNCiAgImFkZCI6ICJjZG4ud3hmei50ayIsDQogICJwb3J0IjogIjgwIiwNCiAgImlkIjogIjExMTExZTMwLTUxZTctNGJiZS1hOGY2LTljMTgzNGI2NmM4ZCIsDQogICJhaWQiOiAiMCIsDQogICJzY3kiOiAiYXV0byIsDQogICJuZXQiOiAid3MiLA0KICAidHlwZSI6ICJub25lIiwNCiAgImhvc3QiOiAidXMwNi5jY3R2dmlwLmNmIiwNCiAgInBhdGgiOiAiL2NjdHYxMy5tM3U4IiwNCiAgInRscyI6ICIiLA0KICAic25pIjogIiIsDQogICJhbHBuIjogIiIsDQogICJmcCI6ICIiDQp9",
		},
		{
			name: "vmess",
			link: "vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIue+juWbvTAxIiwNCiAgImFkZCI6ICJ1czEuemliaWUubGluayIsDQogICJwb3J0IjogIjIwODYiLA0KICAiaWQiOiAiNjFmYTg4MGUtOWYzMC00NzU5LTlmMTktNDVlNmQ0YTU0OTQ5IiwNCiAgImFpZCI6ICIwIiwNCiAgInNjeSI6ICJhdXRvIiwNCiAgIm5ldCI6ICJ3cyIsDQogICJ0eXBlIjogIm5vbmUiLA0KICAiaG9zdCI6ICJ1czEudHV5b3lvLmxpbmsiLA0KICAicGF0aCI6ICIvdHV5b3lvbG92ZWV2ZXJ5b25lIiwNCiAgInRscyI6ICJub25lIiwNCiAgInNuaSI6ICIiLA0KICAiYWxwbiI6ICIiLA0KICAiZnAiOiAiIg0KfQ==",
			net:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.link != "" {
				tt.name = tt.link
			}

			got, err := adapter.ParseLink(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if tt.net {
				buf, _ := json.Marshal(got.ToClash())
				t.Log(string(buf))

				_, err = adapter.ParseClash(got.ToClash())
				if err != nil {
					t.Errorf("%s error = %v", tt.name, err)
					return
				}

				return

				resp, err := got.R().Get("https://www.google.com/generate_204")
				if err != nil {
					t.Errorf("err:%v", err)
					return
				}

				if resp.StatusCode() != 204 {
					t.Errorf("%s status code:%d", tt.name, resp.StatusCode())
					return
				}
			}
		})
	}
}
