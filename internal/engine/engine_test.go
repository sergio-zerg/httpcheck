package engine

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sergio-zerg/httpcheck/internal/result"
	"github.com/sergio-zerg/httpcheck/internal/sensu"
)

func ctxWith(t *testing.T, name, value string) *cli.Context {
	t.Helper()
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.String(name, value, "")
	return cli.NewContext(nil, set, nil)
}

func TestGetPayload(t *testing.T) {
	e := New(ctxWith(t, "env", "prod"))
	r := result.New("host-1", "check:http://x/", "boom", 1500*time.Millisecond, true)

	raw, err := e.getPayload(r)
	if err != nil {
		t.Fatalf("getPayload: %v", err)
	}
	var got sensu.CheckResult
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Status != 2 {
		t.Errorf("want status 2 for error, got %d", got.Status)
	}
	if got.Source != "host-1" || got.Name != "check:http://x/" || got.Output != "boom" {
		t.Errorf("unexpected payload: %+v", got)
	}
	if got.Duration != 1.5 {
		t.Errorf("want duration 1.5s, got %v", got.Duration)
	}
}

func TestSendToSensuPostsPayload(t *testing.T) {
	var gotBody []byte
	var gotCT, gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotCT = r.Header.Get("Content-Type")
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	e := New(ctxWith(t, "sensu-api", srv.URL))
	if err := e.sendToSensu(result.New("ip", "n", "All OK", 0, false)); err != nil {
		t.Fatalf("sendToSensu: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/results" {
		t.Errorf("want POST /results, got %s %s", gotMethod, gotPath)
	}
	if gotCT != "application/json" {
		t.Errorf("want json content-type, got %q", gotCT)
	}
	if len(gotBody) == 0 {
		t.Errorf("expected a request body")
	}
}

func TestSendToSensuErrorsOnBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	e := New(ctxWith(t, "sensu-api", srv.URL))
	if err := e.sendToSensu(result.New("ip", "n", "msg", 0, true)); err == nil {
		t.Fatalf("expected error on 500 from sensu")
	}
}

func TestInitUnknownEnv(t *testing.T) {
	e := New(ctxWith(t, "env", "staging"))
	if err := e.init(); err == nil {
		t.Fatalf("expected error for unknown env")
	}
}

func TestGetConfigFromConsul(t *testing.T) {
	yamlDoc := "site:\n  domains: [example.com]\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		entry := []map[string]string{{"Value": base64.StdEncoding.EncodeToString([]byte(yamlDoc))}}
		_ = json.NewEncoder(w).Encode(entry)
	}))
	defer srv.Close()

	e := New(ctxWith(t, "consul-api", srv.URL))
	got, err := e.getConfigFromConsul()
	if err != nil {
		t.Fatalf("getConfigFromConsul: %v", err)
	}
	if string(got) != yamlDoc {
		t.Errorf("want %q, got %q", yamlDoc, string(got))
	}
}

func TestRunDevEndToEnd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("OK"))
	}))
	defer srv.Close()
	host := srv.URL[len("http://"):]

	doc := "ok:\n  domains: [" + host + "]\n  protocols: [http]\n  path: /\n  status: 200\n"
	path := t.TempDir() + "/checks.yaml"
	if err := os.WriteFile(path, []byte(doc), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.String("env", "dev", "")
	set.String("config", path, "")
	set.String("log-level", "error", "")
	set.String("ip", "", "")
	e := New(cli.NewContext(nil, set, nil))

	if err := e.Run(); err != nil {
		t.Fatalf("Run (dev): %v", err)
	}
}

func TestInitFileMissing(t *testing.T) {
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.String("env", "dev", "")
	set.String("config", "/no/such/checks.yaml", "")
	e := New(cli.NewContext(nil, set, nil))
	if err := e.init(); err == nil {
		t.Fatalf("expected error for missing config file")
	}
}

func TestGetConfigFromConsulEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()

	e := New(ctxWith(t, "consul-api", srv.URL))
	if _, err := e.getConfigFromConsul(); err == nil {
		t.Fatalf("expected error when consul returns no entries")
	}
}
