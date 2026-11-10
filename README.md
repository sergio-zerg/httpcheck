# httpcheck

[![CI](https://github.com/sergio-zerg/httpcheck/actions/workflows/ci.yml/badge.svg)](https://github.com/sergio-zerg/httpcheck/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/sergio-zerg/httpcheck.svg)](https://pkg.go.dev/github.com/sergio-zerg/httpcheck)
[![Go Report Card](https://goreportcard.com/badge/github.com/sergio-zerg/httpcheck)](https://goreportcard.com/report/github.com/sergio-zerg/httpcheck)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A small CLI that runs HTTP health checks defined in YAML and reports the result —
to **stdout** for local use, or to a **Sensu** results API (with check
definitions pulled from **Consul KV**) for a monitoring setup.

> A pet project: a compact, dependency-light Go tool kept as a tidy example of a
> CLI with a file/stdout path for development and a Consul→Sensu path for
> production. Originally written in 2017, modernized in 2026 (Go modules, current
> libraries, tests, CI).

## What it does

For every named check it issues an HTTP(S) request to each `protocol × domain`
target and asserts, as configured:

- **status** — the HTTP status code matches `status`;
- **format** — the `Content-Type` matches `format` (`text` / `json` / `xml` / `yaml`) and the body parses as that type;
- **response** — the body equals `response` exactly.

Each target yields either a single `All OK` result or one result per failed
assertion, stamped with the request duration. Checks run concurrently.

## Install

```sh
go install github.com/sergio-zerg/httpcheck@latest
```

Or build from source:

```sh
git clone https://github.com/sergio-zerg/httpcheck.git
cd httpcheck
make build      # produces ./bin/httpcheck with version info
```

## Usage

```sh
# dev: read checks from a file, print results to stdout
httpcheck --env dev --config checks.yaml --log-level info

# prod: read checks from Consul KV, post results to the Sensu API
httpcheck --env prod --consul-api http://consul:8500 --sensu-api http://sensu:4567
```

| Flag | Env var | Default | Meaning |
|------|---------|---------|---------|
| `--env, -e` | `HTTPCHECK_ENV` | `dev` | `dev` (file→stdout) or `prod` (Consul→Sensu) |
| `--config, -f` | `HTTPCHECK_CONFIG` | `checks.yaml` | checks file (dev) |
| `--consul-api, -c` | `HTTPCHECK_CONSUL_API` | `http://consul-api:8500` | Consul API base URL (prod) |
| `--sensu-api, -s` | `HTTPCHECK_SENSU_API` | `http://sensu-api:4567` | Sensu results API base URL (prod) |
| `--ip, -i` | `HTTPCHECK_IP` | — | override the `Host` header / source for every check |
| `--log-level, -l` | `HTTPCHECK_LOG_LEVEL` | `warning` | `debug` / `info` / `warning` / `error` / `fatal` / `panic` |

## Check file format

```yaml
homepage:                     # check name
  protocols: [http, https]    # defaults to [http]
  domains: [example.com]      # one or more hosts (host[:port])
  path: /
  method: GET                 # defaults to GET
  status: 200                 # optional: assert status code
  format: text                # optional: text | json | xml | yaml
  response: ""                # optional: assert exact body
  timeout: 10                 # optional: per-check timeout, seconds (default 30)
  headers:
    X-Env: prod
  auth:                       # optional HTTP basic auth
    username: user
    password: pass
```

See [`checks.yaml`](checks.yaml) for a runnable example.

## Development

```sh
make test       # go test with the race detector + coverage
make lint       # golangci-lint
make vuln       # govulncheck
make ci         # fmt-check + vet + lint + test + build
```

## Layout

```
main.go              CLI wiring (urfave/cli)
internal/config      parse checks, run requests, assert results
internal/engine      input (file/Consul) + output (stdout/Sensu), concurrency
internal/result      a single check outcome
internal/sensu       Sensu results-API payload
```

## License

[MIT](LICENSE)
