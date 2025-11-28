# Repository Guidelines

Go 1.23 Telegram bot that fetches sources, summarizes with OpenAI, and posts to a channel. Use this guide to stay consistent.

## Project Structure & Modules
- `cmd/main.go` — wires bot, notifier, poster, and HTTP health check.
- `internal/` — `bot` command handlers, `botkit` wrapper, `fetcher` (paused), `notifier`, `services/poster`, `summary` (OpenAI client), `storage` (Postgres), `source` models/helpers, `helpers`.
- Config in `config.hcl` (shared) and `config.local.hcl` (local). Images default to `internal/storage/images`.
- Containers: `docker-compose.local.yml` for Postgres + bot; `Dockerfile` for service image.

## Build, Test, and Run
- `make install-env` — install `golangci-lint` and `moq` into `./bin`.
- `make lint` / `make lint-fast` — static analysis with `.golangci.yml`.
- `go test ./...` — run tests (add alongside code).
- `go run ./cmd/main.go` — start locally with config/env set.
- `docker-compose -f docker-compose.local.yml up --build` — bot + Postgres locally.

## Configuration & Secrets
- Settings load from `config.local.hcl`, then `config.hcl`, then `$HOME/.config/news-grabber-bot/config.hcl`.
- Env vars mirror `internal/config.go`: `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHANNEL_ID`, `DATABASE_DSN`, `OPENAI_KEY`, `IMAGES_DIRECTORY`, `FETCH_INTERVAL`, `NOTIFICATION_INTERVAL`, `FILTER_KEYWORDS`, `OPENAI_MODEL`. Durations use Go syntax (`10m`, `1h`).
- Keep secrets local; avoid committing tokens or DSNs.

## Coding Style & Naming
- Go defaults: tabs, `gofmt` before commit; keep imports ordered (`goimports`).
- Package naming follows existing pattern; export only what other packages need.
- Log with informative prefixes (`[ERROR] …`), avoid silent failures.
- Generate mocks with `moq` into `./internal/.../mock_*`; commit if tests need them.

## Testing Guidelines
- Tests live next to code as `*_test.go`; functions named `TestFunction_Scenario`.
- Prefer table-driven cases; stub Telegram/OpenAI/Postgres via interfaces and `moq`.
- Add regression tests with fixes; keep `go test ./...` green before PRs.

## Commit & Pull Request Expectations
- Commit messages: short, present-tense summaries (e.g., `Add notifier retries`; history shows `Added storage`).
- PRs explain the change, how to verify (`make lint`, `go test ./...`), and required config/env. Add screenshots/log snippets for bot-visible changes.
- Link issues when available and call out follow-ups.
