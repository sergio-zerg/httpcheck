package engine

import (
	"bytes"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"io/ioutil"
	"net/http"
	"github.com/sergio-zerg/httpcheck/config"
	"github.com/sergio-zerg/httpcheck/result"
	"sync"
	"github.com/sergio-zerg/httpcheck/sensu"
	"fmt"
)

func NewEngine(c *cli.Context) *engine {
	engine := engine{
		Context: c,
	}
	return &engine
}

type engine struct {
	*cli.Context
	checks map[string]config.Config
	input  func() ([]byte, error)
	output func(result.Result) error
}

func (e *engine) setLog() {
	logLevel, err := log.ParseLevel(e.Context.String("log-level"))
	log.SetLevel(logLevel)
	if err != nil {
		log.SetLevel(log.WarnLevel)
	}
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
}

func (e *engine) init() {
	env := e.Context.String("env")
	switch env {
	case "dev":
		e.input = e.getConfigFromFile
		e.output = e.sendToStdout
	case "prod":
		e.input = e.getConfigFromConsul
		e.output = e.sendToSensu
	}
	rules, err := e.input()
	if err != nil {
		log.Fatal(err)
	}
	e.checks = config.New(rules)
}

func (e *engine) getConfigFromFile() ([]byte, error) {
	filename := e.Context.String("config")
	out, err := ioutil.ReadFile(filename)
	return out, err
}

func (e *engine) getConfigFromConsul() ([]byte, error) {
	url := e.Context.String("consul-api")
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	type consul_type []struct {
		Value []byte `json:"Value"`
	}
	var data consul_type
	err = json.Unmarshal(body, &data)
	return data[0].Value, err
}

func (e *engine) sendToStdout(data result.Result) error {
	var err error
	if data.IsError {
		log.Warn(data)
	} else {
		log.Info(data)
	}

	return err
}

func (e *engine) sendToSensu(data result.Result) error {
	payload, err:=e.getPayload(data)
	sensuApi := fmt.Sprintf("%s/results", e.Context.String("sensu-api"))
	fmt.Println(string(payload))
	req, err := http.NewRequest("POST", sensuApi, bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	return err
}

func (e engine) getPayload(data result.Result)([]byte, error) {
	status:=0
	if data.IsError{
		status=2
	}
	checkResult:=sensu.SensuCheckResult{
		Source: data.Ip,
		Name: data.Name,
    		Output: data.Message,
    		Status: status,
    		Duration: 0, //@todo calculate duration for each check
    		Occurrences: 3,
	}
	payload, err := json.Marshal(checkResult)
	return payload, err
}

func (e *engine) Run() {
	e.setLog()
	e.init()
	var wg sync.WaitGroup
	wg.Add(len(e.checks))
	for name, check := range e.checks {
		ip := e.Context.String("ip")
		if len(ip) > 0 {
			check.Ip = ip
		}
		go func(name string, check config.Config) {
			defer wg.Done()
			errors := check.Apply(name)
			for _, msg := range errors {
				err := e.output(msg)
				if err != nil {
					log.Error(err)
				}
			}
		}(name, check)
	}
	wg.Wait()
}