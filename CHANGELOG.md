# Changelog

All notable changes to this project are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.0.0] - 2026-11-10

Full modernization pass. The CLI behaviour is compatible; internals and the
toolchain changed substantially, hence the major bump.

### Changed
- Migrated from `glide` to **Go modules**; refreshed every dependency.
- `github.com/Sirupsen/logrus` → `github.com/sirupsen/logrus` v1.9.3.
- `urfave/cli` v1 → **v2**.
- `gopkg.in/yaml.v2` → `gopkg.in/yaml.v3`.
- Moved packages under `internal/` (`config`, `engine`, `result`, `sensu`).
- Replaced the iris-based manual test server with `net/http/httptest` tests.

### Added
- Unit + integration tests (`config`, `engine`, `result`) with coverage.
- GitHub Actions CI: fmt-check, vet, golangci-lint, tests, govulncheck, build.
- `Makefile`, `.golangci.yml`, `CHANGELOG.md`, real `README.md`.
- Per-check request **duration**, surfaced in output and the Sensu payload.
- Per-check `timeout` option (default 30s) and an HTTP client timeout.

### Fixed
- Panic when a response had no `Content-Type` header (indexing `[0]`).
- `All OK` was only emitted when a `response` assertion was configured; now it
  is emitted whenever every configured assertion passes.
- Nil-pointer dereference when the POST to the Sensu API failed.
- Panic when Consul returned an empty KV array.
- Missing HTTP client timeout (requests could hang indefinitely).
- Unchecked YAML parse error when loading checks.
- Unknown `--env` silently did nothing; it now returns an error.

### Removed
- `glide.yaml`, `glide.lock`, `LINKS.md`, stray `t.yaml`, iris test server.
- Hard-coded author e-mail in the CLI metadata.

## [1.0.0] - 2017
- Initial version: file/stdout (dev) and Consul/Sensu (prod) HTTP checks.
