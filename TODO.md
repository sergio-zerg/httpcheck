# TODO

Planned functional improvements (not blocking; the tool is complete for its
current scope).

- **Request bodies / params** — support POST form, JSON and XML bodies via the
  `params` field (currently parsed but unused).
- **Retries** — optional retry count with backoff per check before reporting an
  error.
- **Validate URLs** — validate the Consul and Sensu API URLs on startup in prod.
- **Health indicators** — expose per-check status indicators, e.g. in the style of
  <https://github.com/shurshun/docker-health-bridge>.
- **cmd/ layout** — if the surface grows, move `main.go` into `cmd/httpcheck/`.
