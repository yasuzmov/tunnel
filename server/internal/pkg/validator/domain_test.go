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
