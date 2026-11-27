**English** | [Русский](ROADMAP.ru.md)

# Roadmap

Status: ✅ done · 🚧 in progress · 🗓 planned · ❌ not planned

The tool is complete for its current scope; the planned items are not blocking.

## Done

- ✅ HTTP(S) checks from YAML: status code, content type with body parsing (text / json / xml / yaml), exact body
- ✅ Two paths: file → stdout for development, Consul KV → Sensu results API for production
- ✅ Concurrent checks, request duration in every result, per-check timeout
- ✅ 2.0.0 modernization: Go modules, current libraries, `internal/` layout, tests, CI with lint and govulncheck

## Planned

- 🗓 **Request bodies / params** — POST form, JSON and XML bodies via the `params` field (currently parsed but unused)
- 🗓 **Retries** — optional retry count with backoff per check before reporting an error
- 🗓 **Validate URLs** — validate the Consul and Sensu API URLs on startup in prod
- 🗓 **Health indicators** — per-check status indicators, e.g. in the style of
  [docker-health-bridge](https://github.com/shurshun/docker-health-bridge)
- 🗓 **cmd/ layout** — if the surface grows, move `main.go` into `cmd/httpcheck/`
