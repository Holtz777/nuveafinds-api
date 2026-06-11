---
name: nuvea-api
description: Go API backend para Nuvea Finds — convenções de código, estrutura internal/, contratos HTTP, OpenRouter AI, Pinterest v5. Use para qualquer modificação no código Go da API.
---

# Nuvea API — Skill de Desenvolvimento Go

Backend Go da Nuvea Finds: pin automation pipeline com OpenRouter AI e Pinterest API v5. Stdlib-only, sem frameworks externos.

---

## 0. Estrutura

```
api/
├── cmd/server/main.go             # entrypoint: config -> deps -> server
├── internal/
│   ├── config/config.go           # .env loader (bufio.Scanner, stdlib)
│   ├── httpx/httpx.go             # JSON helpers + CORS + Logger
│   ├── ai/openrouter.go           # OpenRouter client + pin generation prompt
│   ├── pinterest/pinterest.go     # Pinterest v5 client (media, upload, pins, boards)
│   ├── pinterest/tokens.go        # OAuth token store + auto-refresh
│   └── handlers/handlers.go       # HTTP route handlers
├── deploy/nginx-api.conf          # Nginx config para VPS
├── Dockerfile                     # multi-stage build (golang:1.23 → alpine:3.20)
├── docker-compose.yml             # serviço único, bind 127.0.0.1:8080
├── go.mod                         # module github.com/Holtz777/nuveafinds-api
├── .env.example
└── .env                           # gitignored
```

## 1. Convenções de Código

### Regra de ouro: stdlib-only

```go
// ✅ Correto
import "net/http"
import "encoding/json"

// ❌ Errado (sem discussão prévia)
import "github.com/gorilla/mux"
```

### HTTP

- Roteador nativo `http.ServeMux` (Go 1.22+ com method-based routing: `"POST /rota"`)
- Handlers recebem `*handlers.Deps` via闭包
- Response sempre `{"status":"success","data":{...}}` ou `{"status":"error","message":"..."}`
- Usar `httpx.Success(w, data)` e `httpx.Error(w, status, msg)`

### Config

- `config.Load()` no boot, carrega `.env` + env vars
- System env vars têm prioridade sobre `.env`
- `PINTEREST_ACCESS_TOKEN` é opcional (sem ele, endpoints Pinterest falham mas API sobe)

### Error handling

- Retornar errors pra cima, logar no topo
- Não logar E retornar o mesmo erro
- Usar `fmt.Errorf("contexto: %w", err)` para wrapping

### Convenções Go idiomáticas

- Receivers curtos: `c` pra Client, `s` pra TokenStore
- Nada de naked returns
- Nada de panic em library code
- Testar com `go test ./...` (ainda sem testes implementados)

## 2. Rotas da API

| Método | Rota | Handler | Descrição |
|--------|------|---------|-----------|
| GET | `/health` | `d.health` | Healthcheck |
| GET | `/boards` | `d.listBoards` | Lista boards + gera board map |
| GET | `/token-info` | `d.tokenInfo` | Info do token OAuth |
| POST | `/token-refresh` | `d.tokenRefresh` | Força refresh de token |
| GET | `/mode` | `d.getMode` | Modo sandbox/production |
| POST | `/mode` | `d.setMode` | Alterna sandbox/production |
| POST | `/pin-upload` | `d.pinUpload` | Gera títulos/descrições (IA) |
| POST | `/pin-register-video` | `d.pinRegisterVideo` | Registra mídia no Pinterest |
| POST | `/proxy/upload-video` | `d.proxyUploadVideo` | Upload proxy pro S3 Pinterest |
| POST | `/pin-publish` | `d.pinPublish` | Cria Video Pin |

## 3. Pipeline do Pin

```
Produto → IA (OpenRouter) → 2 versões A/B + board slug
  → Register media (Pinterest /v5/media) → mediaId + uploadUrl
  → Upload video (S3 multipart, bufferizado com Content-Length)
  → Poll media status até "succeeded"
  → Create pin (Pinterest POST /v5/pins) → Pin publicado
```

## 4. OpenRouter (IA)

- Endpoint: `https://openrouter.ai/api/v1/chat/completions`
- Modelo configurável via `OPENROUTER_MODEL` (default: `anthropic/claude-sonnet-4`)
- Prompt gera JSON estruturado: 2 versões A/B + board slug
- JSON mode: `response_format: { type: "json_object" }`
- Boards válidas hardcoded em `validBoards` map

## 5. Pinterest API v5

- Sandbox: `https://api-sandbox.pinterest.com/v5/`
- Produção: `https://api.pinterest.com/v5/`
- App ID: 1560888
- Status: **Trial** (Standard Access negado 2026-06-11 — precisa recriar demo video)

### Upload de vídeo

- S3 do Pinterest exige `Content-Length` explícito
- Upload é bufferizado (`bytes.Buffer`), não streaming
- Limite: `maxUploadSize = 2 * 1024 * 1024 * 1024` (2 GB)

### Token OAuth

- Prefixo `pina_` (access), `pinr_` (refresh)
- Auto-refresh a cada 25 dias via `TokenStore.RunRefresher`
- Refresh token e credenciais salvos em `pinterest_token.json`

## 6. Comandos

```bash
cd api

# Desenvolvimento
go run ./cmd/server          # sobe servidor (carrega .env)
go build ./...               # verifica compilação
go vet ./...                 # lint estático

# Docker
docker compose up --build    # build + run
docker compose logs -f       # acompanhar logs

# Testar endpoints
curl http://localhost:8080/health
curl http://localhost:8080/boards
```

## 7. GitHub

- **Repo:** https://github.com/Holtz777/nuveafinds-api
- **Branch:** `main`
- **Module:** `github.com/Holtz777/nuveafinds-api`
- `.env` é gitignored — nunca commitar secrets
