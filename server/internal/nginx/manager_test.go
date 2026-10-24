package nginx

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"tunnel-server/internal/config"
)

type mockRunner struct {
	failTest bool
	calls    [][]string
}

func (m *mockRunner) Run(_ context.Context, name string, args ...string) (Result, error) {
	m.calls = append(m.calls, append([]string{name}, args...))
	if len(args) > 0 && args[0] == "-t" && m.failTest {
		return Result{Stderr: "syntax error"}, errors.New("nginx -t failed")
	}
	return Result{}, nil
}

func TestApplyReload(t *testing.T) {
	dir := t.TempDir()
	runner := &mockRunner{}
	m := NewManager(config.NginxConfig{
		Enabled:           true,
		ConfDir:           dir,
		NginxBin:          "nginx",
		IngressUpstream:   "127.0.0.1:8081",
		ClientMaxBodySize: "50m",
		ACMEWebroot:       "/var/www/acme",
	}, runner)
	if _, err := m.Apply("app.example.com", false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "app.example.com.conf")); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) < 2 {
		t.Fatalf("want nginx -t and reload, got %v", runner.calls)
	}
}

func TestApplyRollbackOnTestFail(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.example.com.conf")
	if err := os.WriteFile(path, []byte("OLD"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := NewManager(config.NginxConfig{
		Enabled: true,
		ConfDir: dir,
		NginxBin: "nginx",
		IngressUpstream: "127.0.0.1:8081",
	}, &mockRunner{failTest: true})
	if _, err := m.Apply("app.example.com", false, false); err == nil {
		t.Fatal("expected nginx -t failure")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "OLD" {
		t.Fatalf("want rollback to OLD, got %s", got)
	}
}

func TestPreviewWhenDisabled(t *testing.T) {
	runner := &mockRunner{}
	m := NewManager(config.NginxConfig{Enabled: false, IngressUpstream: "127.0.0.1:8081"}, runner)
	conf, err := m.Apply("app.example.com", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if conf == "" {
		t.Fatal("expected preview text")
	}
	if len(runner.calls) != 0 {
		t.Fatalf("disabled nginx should not run commands: %v", runner.calls)
	}
}
