# api/ — Nuvea Finds API (Go)

## Purpose

Go backend that automates Pinterest Video Pin creation for Nuvea Finds. Generates titles/descriptions via OpenRouter AI, registers media and creates pins via Pinterest API v5.

## Ownership

- `cmd/server/main.go` — entrypoint: config load, dependency wiring, HTTP server, graceful shutdown, token refresher
- `internal/config/` — .env loader (stdlib-only, `bufio.Scanner`), env vars → Config struct
- `internal/httpx/` — JSON response helpers, CORS middleware, request logger
- `internal/ai/` — OpenRouter client, pin generation prompt, board selection
- `internal/pinterest/` — Pinterest v5 client (media register, S3 upload, pin create, board list) + OAuth token store with auto-refresh
- `internal/handlers/` — HTTP route handlers, request parsing, response formatting
- `deploy/` — Nginx config for VPS reverse proxy
- `Dockerfile` — multi-stage build (golang:1.23-alpine → alpine:3.20), static binary CGO_ENABLED=0
- `docker-compose.yml` — single service, binds to 127.0.0.1:8080

## Local Contracts

### Response format
All endpoints return `{"status":"success","data":{...}}` or `{"status":"error","message":"..."}`.
Use `httpx.Success(w, data)` and `httpx.Error(w, status, msg)`.

### Routes
| Route | Method | Handler |
|-------|--------|---------|
| `/health` | GET | health check |
| `/boards` | GET | list Pinterest boards + generate board map |
| `/token-info` | GET | OAuth token metadata |
| `/token-refresh` | POST | force token refresh |
| `/mode` | GET/POST | get/set sandbox mode |
| `/pin-upload` | POST | AI title/description generation |
| `/pin-register-video` | POST | register media on Pinterest |
| `/proxy/upload-video` | POST | proxy upload to Pinterest S3 |
| `/pin-publish` | POST | create Video Pin |

### Config loading
`config.Load()` must be called once at startup. It loads `.env` via `loadDotEnv()` (stdlib `bufio.Scanner`), then reads env vars. System env vars take priority over `.env` file.

### Dependencies
`handlers.Deps` struct holds all dependencies. Pass it to `handlers.Register(mux, deps)`.

## Work Guidance

### Go conventions
- **No external dependencies.** The `go.mod` must stay clean (stdlib only). If a dep becomes unavoidable, discuss first.
- Use `net/http` + `http.ServeMux` (Go 1.22+ method-based routing).
- Use `encoding/json` for JSON. Use `mime/multipart` for multipart.
- Use `context.Context` for cancellation and timeouts.
- Error handling: return errors up, log at the top level. Don't log and return the same error.
- Config: env vars only. No config files, no flags.

### Code style
- Keep files small and focused. One responsibility per file.
- Use descriptive variable names. Avoid single-letter names except for loop indices and receivers.
- Receivers: use short abbreviations (`c` for Client, `s` for TokenStore).
- Comments: document public functions and types. Explain "why", not "what".
- No naked returns. No panics in library code.

### Pinterest integration
- Always check sandbox vs production mode before calling Pinterest APIs.
- Media upload must use `bytes.Buffer` (not streaming) because Pinterest S3 requires `Content-Length`.
- Token refresh runs automatically every ~25 days via `TokenStore.RunRefresher`.

### AI integration
- OpenRouter endpoint: `https://openrouter.ai/api/v1/chat/completions`
- Model is configurable via `OPENROUTER_MODEL` env var.
- Pin generation uses JSON mode (`response_format: { type: "json_object" }`).
- Valid board slugs are hardcoded in `validBoards` map.

## Verification

```bash
go build ./...        # must exit 0
go vet ./...          # must exit 0
go run ./cmd/server   # starts HTTP server, loads .env
```

## Child DOX Index

| Path | Scope | Description |
|------|-------|-------------|
| `internal/ai/` | OpenRouter AI client | Covered by this doc (single file, no child needed) |
| `internal/pinterest/` | Pinterest API + OAuth | Covered by this doc (2 files, no child needed) |
| `internal/handlers/` | HTTP route handlers | Covered by this doc (single file, no child needed) |
| `internal/config/` | Environment config | Covered by this doc (single file, no child needed) |
| `internal/httpx/` | HTTP utilities | Covered by this doc (single file, no child needed) |
