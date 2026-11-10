// Package config parses check definitions and runs them against their targets.
package config

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/sergio-zerg/httpcheck/internal/result"
)

const (
	msgErrorHTTPStatus  = "config status_code = %d and returned status_code = %d are not equal"
	msgErrorContentType = "config format = %q and returned content_type = %q are not equal"
	msgErrorAnswer      = "config answer = %q and returned answer = %q are not equal"
	msgSuccess          = "All OK"

	// defaultTimeout caps a single HTTP check when none is set in the config.
	defaultTimeout = 30 * time.Second
)

// Config is a single named check definition.
type Config struct {
	IP        string   `yaml:"ip"`
	Protocols []string `yaml:"protocols"`
	Domains   []string `yaml:"domains"`
	Auth      struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"auth"`
	Headers  map[string]string `yaml:"headers"`
	Path     string            `yaml:"path"`
	Method   string            `yaml:"method"`
	Params   map[string]string `yaml:"params"`
	Format   string            `yaml:"format"`
	Response string            `yaml:"response"`
	Status   int               `yaml:"status"`
	Timeout  int               `yaml:"timeout"` // per-check timeout in seconds; 0 = default
}

// New parses a YAML document into a map of named checks.
func New(data []byte) (map[string]Config, error) {
	checks := map[string]Config{}
	if err := yaml.Unmarshal(data, &checks); err != nil {
		return nil, fmt.Errorf("parse checks: %w", err)
	}
	return checks, nil
}

func (c Config) timeout() time.Duration {
	if c.Timeout > 0 {
		return time.Duration(c.Timeout) * time.Second
	}
	return defaultTimeout
}

func (c Config) checkResponseFormat(data []byte, formatType, contentType string) (bool, error) {
	var err error
	switch formatType {
	case "xml":
		var test interface{}
		err = xml.Unmarshal(data, &test)
	case "json":
		var test interface{}
		err = json.Unmarshal(data, &test)
	case "yaml":
		var test interface{}
		err = yaml.Unmarshal(data, &test)
	}
	matched, matchErr := regexp.MatchString(formatType, contentType)
	if err == nil {
		err = matchErr
	}
	return matched, err
}

func (c *Config) request(protocol, domain string) (*http.Response, error) {
	client := &http.Client{Timeout: c.timeout()}
	if protocol == "https" {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		}
	}

	method := strings.ToUpper(c.Method)
	if method == "" {
		method = http.MethodGet
	}

	address := protocol + "://" + domain + c.Path
	req, err := http.NewRequest(method, address, new(bytes.Buffer))
	if err != nil {
		return nil, err
	}
	if c.Auth.Username != "" {
		req.SetBasicAuth(c.Auth.Username, c.Auth.Password)
	}
	for name, value := range c.Headers {
		req.Header.Set(name, value)
	}
	if c.IP != "" {
		req.Host = c.IP
	}

	return client.Do(req)
}

// fetch performs one request and returns the body, closing it before returning.
func (c *Config) fetch(protocol, domain string) (*http.Response, []byte, error) {
	resp, err := c.request(protocol, domain)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return resp, data, nil
}

// Apply runs the check against every protocol/domain combination and returns
// one Result per target: a success Result when every configured assertion
// passed, otherwise one Result per failed assertion.
func (c *Config) Apply(name string) []result.Result {
	results := []result.Result{}
	if len(c.Protocols) == 0 {
		c.Protocols = []string{"http"}
	}
	for _, protocol := range c.Protocols {
		for _, domain := range c.Domains {
			key := name + ":" + protocol + "://" + domain + c.Path
			start := time.Now()
			resp, data, err := c.fetch(protocol, domain)
			dur := time.Since(start)
			if err != nil {
				results = append(results, result.New(c.IP, key, err.Error(), dur, true))
				continue
			}

			hasErrors := false

			if c.Format != "" {
				contentType := resp.Header.Get("Content-Type")
				matched, ferr := c.checkResponseFormat(data, c.Format, contentType)
				if !matched {
					msg := fmt.Sprintf(msgErrorContentType, c.Format, contentType)
					hasErrors = true
					results = append(results, result.New(c.IP, key, msg, dur, true))
				}
				if ferr != nil {
					hasErrors = true
					results = append(results, result.New(c.IP, key, ferr.Error(), dur, true))
				}
			}

			if c.Status != 0 && resp.StatusCode != c.Status {
				msg := fmt.Sprintf(msgErrorHTTPStatus, c.Status, resp.StatusCode)
				hasErrors = true
				results = append(results, result.New(c.IP, key, msg, dur, true))
			}

			if c.Response != "" && c.Response != string(data) {
				msg := fmt.Sprintf(msgErrorAnswer, c.Response, string(data))
				hasErrors = true
				results = append(results, result.New(c.IP, key, msg, dur, true))
			}

			if !hasErrors {
				results = append(results, result.New(c.IP, key, msgSuccess, dur, false))
			}
		}
	}
	return results
}
