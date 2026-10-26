package validator

import "testing"

func TestValidateDomainOK(t *testing.T) {
	got, err := ValidateDomain("App.Example.COM")
	if err != nil {
		t.Fatal(err)
	}
	if got != "app.example.com" {
		t.Fatalf("got %s", got)
	}
}

func TestValidateDomainReject(t *testing.T) {
	bads := []string{
		"",
		"localhost",
		"*.example.com",
		"example.com/path",
		"example.com;id",
		"a b.com",
		"127.0.0.1",
		"-abc.com",
	}
	for _, d := range bads {
		if _, err := ValidateDomain(d); err == nil {
			t.Fatalf("should reject %q", d)
		}
	}
}

func TestValidateLabel(t *testing.T) {
	ok := map[string]string{"abc": "abc", " ABC ": "abc", "a": "a", "a-1": "a-1", "9z": "9z"}
	for in, want := range ok {
		got, err := ValidateLabel(in)
		if err != nil || got != want {
			t.Errorf("%q: want %q, got %q err=%v", in, want, got, err)
		}
	}
	long := ""
	for i := 0; i < 64; i++ {
		long += "a"
	}
	for _, in := range []string{"", " ", "-a", "a-", "a.b", "a_b", "中文", "a b", "*", "www", "WWW", long} {
		if _, err := ValidateLabel(in); err == nil {
			t.Errorf("%q: 应被拒绝", in)
		}
	}
	if _, err := ValidateLabel(long[:63]); err != nil {
		t.Errorf("63 位应合法: %v", err)
	}
}

func TestSubdomainLabel(t *testing.T) {
	base := "fish.czypro.cn"
	cases := map[string]string{
		"abc.fish.czypro.cn":     "abc",
		"a.b.fish.czypro.cn":     "",
		"fish.czypro.cn":         "",
		"xfish.czypro.cn":        "",
		"abc.fish.czypro.cn.com": "",
		"abc.other.cn":           "",
	}
	for host, want := range cases {
		if got := SubdomainLabel(host, base); got != want {
			t.Errorf("%s: want %q, got %q", host, want, got)
		}
	}
	if got := SubdomainLabel("abc.fish.czypro.cn", ""); got != "" {
		t.Errorf("空主域名应返回空, got %q", got)
	}
}
