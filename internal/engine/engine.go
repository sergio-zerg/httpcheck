// Package engine wires inputs (file/Consul) and outputs (stdout/Sensu) together
// and runs every configured check concurrently.
package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/sergio-zerg/httpcheck/internal/config"
	"github.com/sergio-zerg/httpcheck/internal/result"
	"github.com/sergio-zerg/httpcheck/internal/sensu"
)

const httpTimeout = 15 * time.Second

// Engine owns the input/output strategy selected by the --env flag.
type Engine struct {
	ctx    *cli.Context
	checks map[string]config.Config
	input  func() ([]byte, error)
	output func(result.Result) error
	client *http.Client
}

// New builds an Engine from the CLI context.
func New(c *cli.Context) *Engine {
	return &Engine{
		ctx:    c,
		client: &http.Client{Timeout: httpTimeout},
	}
}

func (e *Engine) setLog() {
	level, err := log.ParseLevel(e.ctx.String("log-level"))
	if err != nil {
		level = log.WarnLevel
	}
	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
}

func (e *Engine) init() error {
	switch e.ctx.String("env") {
	case "dev":
		e.input = e.getConfigFromFile
		e.output = e.sendToStdout
	case "prod":
		e.input = e.getConfigFromConsul
		e.output = e.sendToSensu
	default:
		return fmt.Errorf("unknown env %q (want dev or prod)", e.ctx.String("env"))
	}

	rules, err := e.input()
	if err != nil {
		return err
	}
	checks, err := config.New(rules)
	if err != nil {
		return err
	}
	e.checks = checks
	return nil
}

func (e *Engine) getConfigFromFile() ([]byte, error) {
	return os.ReadFile(e.ctx.String("config"))
}

func (e *Engine) getConfigFromConsul() ([]byte, error) {
	resp, err := e.client.Get(e.ctx.String("consul-api"))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Consul KV returns a JSON array; Value is base64 in the wire format and is
	// decoded automatically into the []byte field by encoding/json.
	var data []struct {
		Value []byte `json:"Value"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("consul returned no key/value entries")
	}
	return data[0].Value, nil
}

func (e *Engine) sendToStdout(data result.Result) error {
	if data.IsError {
		log.Warn(data)
	} else {
		log.Info(data)
	}
	return nil
}

func (e *Engine) sendToSensu(data result.Result) error {
	payload, err := e.getPayload(data)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/results", e.ctx.String("sensu-api"))
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("sensu api %s returned %s", url, resp.Status)
	}
	return nil
}

func (e *Engine) getPayload(data result.Result) ([]byte, error) {
	status := 0
	if data.IsError {
		status = 2
	}
	return json.Marshal(sensu.CheckResult{
		Source:      data.IP,
		Name:        data.Name,
		Output:      data.Message,
		Status:      status,
		Duration:    data.Duration.Seconds(),
		Occurrences: 3,
	})
}

// Run sets up logging, loads checks, and runs them all concurrently.
func (e *Engine) Run() error {
	e.setLog()
	if err := e.init(); err != nil {
		return err
	}

	ip := e.ctx.String("ip")
	var wg sync.WaitGroup
	wg.Add(len(e.checks))
	for name, check := range e.checks {
		if ip != "" {
			check.IP = ip
		}
		go func(name string, check config.Config) {
			defer wg.Done()
			for _, msg := range check.Apply(name) {
				if err := e.output(msg); err != nil {
					log.Error(err)
				}
			}
		}(name, check)
	}
	wg.Wait()
	return nil
}
