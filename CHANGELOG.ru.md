[English](CHANGELOG.md) | **Русский**

# Changelog

Здесь записаны все заметные изменения проекта.
Формат — [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/),
версии — [Semantic Versioning](https://semver.org/lang/ru/spec/v2.0.0.html).

## [Unreleased]

## [2.0.0] - 2026-11-10

Полное обновление. Поведение CLI совместимо; внутреннее устройство и инструменты
изменились сильно, поэтому мажорная версия.

### Изменено
- Переход с `glide` на **Go modules**; обновлены все зависимости.
- `github.com/Sirupsen/logrus` → `github.com/sirupsen/logrus` v1.9.3.
- `urfave/cli` v1 → **v2**.
- `gopkg.in/yaml.v2` → `gopkg.in/yaml.v3`.
- Пакеты перенесены в `internal/` (`config`, `engine`, `result`, `sensu`).
- Ручной тестовый сервер на iris заменён тестами на `net/http/httptest`.

### Добавлено
- Модульные и интеграционные тесты (`config`, `engine`, `result`) с покрытием.
- CI на GitHub Actions: fmt-check, vet, golangci-lint, тесты, govulncheck, сборка.
- `Makefile`, `.golangci.yml`, `CHANGELOG.md`, настоящий `README.md`.
- **Длительность** запроса по каждой проверке — в выводе и в теле запроса к Sensu.
- Параметр `timeout` у проверки (по умолчанию 30 с) и таймаут HTTP-клиента.

### Исправлено
- Паника, если в ответе не было заголовка `Content-Type` (обращение к `[0]`).
- `All OK` выводился, только если была задана проверка `response`; теперь — всегда, когда
  проходят все заданные проверки.
- Разыменование nil, если POST в API Sensu завершался ошибкой.
- Паника, если Consul возвращал пустой массив KV.
- Не было таймаута HTTP-клиента (запросы могли висеть бесконечно).
- Не проверялась ошибка разбора YAML при загрузке проверок.
- Неизвестный `--env` молча ничего не делал; теперь возвращает ошибку.

### Удалено
- `glide.yaml`, `glide.lock`, `LINKS.md`, лишний `t.yaml`, тестовый сервер на iris.
- Зашитый в метаданные CLI e-mail автора.

## [1.0.0] - 2017
- Первая версия: HTTP-проверки файл/stdout (dev) и Consul/Sensu (prod).
