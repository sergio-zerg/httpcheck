package config

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNewValid(t *testing.T) {
	checks, err := New([]byte("site:\n  domains: [example.com]\n  status: 200\n"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, ok := checks["site"]; !ok {
		t.Fatalf("expected check %q in %v", "site", checks)
	}
}

func TestNewInvalid(t *testing.T) {
	if _, err := New([]byte("site: [unterminated\n")); err == nil {
		t.Fatalf("expected parse error for malformed yaml")
	}
}

// target splits an httptest server URL into (protocol, host) for a Config.
func target(t *testing.T, srv *httptest.Server) (string, string) {
	t.Helper()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	return u.Scheme, u.Host
}

func newConfig(proto, host string) Config {
	var c Config
	c.Protocols = []string{proto}
	c.Domains = []string{host}
	c.Path = "/"
	return c
}

func TestApplySuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer srv.Close()

	proto, host := target(t, srv)
	c := newConfig(proto, host)
	c.Status = 200
	c.Format = "text"
	c.Response = "OK"

	res := c.Apply("ok")
	if len(res) != 1 {
		t.Fatalf("want 1 result, got %d: %+v", len(res), res)
	}
	if res[0].IsError || res[0].Message != msgSuccess {
		t.Fatalf("want success, got %+v", res[0])
	}
}

func TestApplyStatusMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	proto, host := target(t, srv)
	c := newConfig(proto, host)
	c.Status = 200

	res := c.Apply("status")
	if len(res) != 1 || !res[0].IsError {
		t.Fatalf("want one error result, got %+v", res)
	}
}

func TestApplyResponseMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("actual body"))
	}))
	defer srv.Close()

	proto, host := target(t, srv)
	c := newConfig(proto, host)
	c.Response = "expected body"

	res := c.Apply("resp")
	if len(res) != 1 || !res[0].IsError {
		t.Fatalf("want one error result, got %+v", res)
	}
}

func TestApplyConnectionError(t *testing.T) {
	c := newConfig("http", "127.0.0.1:0") // nothing listening
	c.Timeout = 1
	res := c.Apply("down")
	if len(res) != 1 || !res[0].IsError {
		t.Fatalf("want one error result for unreachable host, got %+v", res)
	}
}

func TestApplyDefaultsProtocolHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hi"))
	}))
	defer srv.Close()

	_, host := target(t, srv)
	var c Config
	c.Domains = []string{host}
	c.Path = "/"
	// no Protocols set -> defaults to http

	res := c.Apply("default")
	if len(res) != 1 || res[0].IsError {
		t.Fatalf("want one success result, got %+v", res)
	}
}
