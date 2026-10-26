package config

import "testing"

func TestBaseDomain(t *testing.T) {
	cases := []struct {
		cfg  AdminConfig
		want string
	}{
		{AdminConfig{PublicURL: "https://fish.czypro.cn"}, "fish.czypro.cn"},
		{AdminConfig{PublicURL: "https://Fish.CZYPRO.cn:8443/admin"}, "fish.czypro.cn"},
		{AdminConfig{PublicURL: "https://fish.czypro.cn", DomainSuffix: "*.tun.example.com."}, "tun.example.com"},
		{AdminConfig{PublicURL: "http://43.142.26.241:8080"}, ""},
		{AdminConfig{PublicURL: "http://localhost:8080"}, ""},
		{AdminConfig{}, ""},
	}
	for _, tc := range cases {
		if got := tc.cfg.BaseDomain(); got != tc.want {
			t.Errorf("%+v: want %q, got %q", tc.cfg, tc.want, got)
		}
	}
}
