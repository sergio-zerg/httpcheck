package config

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"github.com/sergio-zerg/httpcheck/result"
	"regexp"
	"strings"
)

const (
	msg_error_http_status  = "Config status_code = %d and returned status_code = %d are not equal"
	msg_error_content_type = "Config format = %v and returned content_type = %v are not equal"
	msg_error_answer       = "Config answer = %v and returned answer = %v are not equal"
	msg_success            = "All OK"
)

func New(data []byte) map[string]Config {
	checks := map[string]Config{}
	yaml.Unmarshal(data, &checks)
	return checks
}

type Config struct {
	Ip        string   `yaml:"ip"`
	Protocols []string `yaml:"protocols"`
	Domains   []string `yaml:"domains"`
	Auth      struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"auth"`
	Headers  map[string]string
	Path     string            `yaml:"path"`
	Method   string            `yaml:"method"`
	Params   map[string]string `yaml:"params"`
	Format   string            `yaml:"format"`
	Response string            `yaml:"response"`
	Status   int               `yaml:"status"`
}

func (c Config) checkResponseFormat(data []byte, format_type, content_type string) (bool, error) {
	var test interface{}
	var err error
	switch format_type {
	case "xml":
		err = xml.Unmarshal([]byte(data), &test)
	case "json":
		err = json.Unmarshal([]byte(data), &test)
	case "yaml":
		err = yaml.Unmarshal([]byte(data), &test)
	}
	matched, _ := regexp.MatchString(format_type, content_type)
	return matched, err
}

func (c *Config) request(protocol, domain string) (*http.Response, error) {
	client := &http.Client{}
	address := protocol + "://" + domain + c.Path
	if protocol == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{}, //InsecureSkipVerify: true},
		}
		client.Transport = tr
	}
	//@todo timeout, request count, params(post & post form)
	buffer := new(bytes.Buffer)
	//
	//if len(c.Params)>0{
	//	params := url.Values{}
	//	for name, value := range c.Params {
	//		params.Add(name, value)
	//	}
	//	buffer.WriteString(params.Encode())
	//}

	if c.Method==""{
		c.Method="GET"
	}

	req, err := http.NewRequest(strings.ToUpper(c.Method), address, buffer)
	if err != nil {
		return nil, err
	}
	if c.Auth.Username != "" {
		req.SetBasicAuth(c.Auth.Username, c.Auth.Password)
	}
	if len(c.Headers) > 0 {
		for name, value := range c.Headers {
			req.Header.Set(name, value)
		}
	}
	req.Host = c.Ip
	//fmt.Printf("%#v", req.URL.Query().Get("a"))

	resp, err := client.Do(req)
	return resp, err
}

func (c *Config) Apply(name string) []result.Result {
	results := []result.Result{}
	if len(c.Protocols) == 0 {
		c.Protocols = append(c.Protocols, "http")
	}
	for _, protocol := range c.Protocols {
		for _, domain := range c.Domains {
			has_errors := false
			key := name + ":" + protocol + "://" + domain + c.Path
			resp, err := c.request(protocol, domain)
			if err != nil {
				has_errors = true
				results = append(results, result.New(c.Ip, key, err.Error(), has_errors))
				continue
			}
			data, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				has_errors = true
				results = append(results, result.New(c.Ip, key, err.Error(), has_errors))
				continue
			}

			//check response format
			if c.Format != "" {
				content_type := resp.Header["Content-Type"][0]
				matched, err := c.checkResponseFormat(data, c.Format, content_type)
				if !matched {
					msg := fmt.Sprintf(msg_error_content_type, c.Format, content_type)
					has_errors = true
					results = append(results, result.New(c.Ip, key, msg, has_errors))
				}
				if err != nil {
					has_errors = true
					results = append(results, result.New(c.Ip, key, err.Error(), has_errors))
				}
			}

			//check status code
			if c.Status != 0 && resp.StatusCode != c.Status {
				msg := fmt.Sprintf(msg_error_http_status, c.Status, resp.StatusCode)
				has_errors = true
				results = append(results, result.New(c.Ip, key, msg, has_errors))
			}

			//check response equality
			if c.Response!=""{
				response := string(data)
				if c.Response != response {
					msg := fmt.Sprintf(msg_error_answer, c.Response, response)
					has_errors = true
					results = append(results, result.New(c.Ip, key, msg, has_errors))
				}
				if !has_errors {
					results = append(results, result.New(c.Ip, key, msg_success, has_errors))
				}
			}

		}
	}
	return results
}