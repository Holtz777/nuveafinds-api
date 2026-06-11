# nuveafinds-api

API em Go que substitui o workflow n8n do projeto Nuvea Finds. Um binário único expõe:

| Rota                     | O que faz                                                                                   |
| ------------------------ | ------------------------------------------------------------------------------------------- |
| `GET  /health`           | Healthcheck.                                                                                |
| `POST /pin-upload`       | Recebe dados do produto, chama OpenRouter e devolve 2 versões de título/descrição + board.  |
| `POST /pin-register-video` | Registra uma mídia de vídeo no Pinterest (`POST /v5/media`) e devolve `upload_url` + params. |
| `POST /proxy/upload-video` | Recebe o arquivo do browser e faz o POST multipart no S3 do Pinterest (evita CORS).         |
| `POST /pin-publish`      | Cria o Video Pin (`POST /v5/pins`) após a mídia ficar pronta.                               |

## Stack

- Go 1.23, só biblioteca padrão (`net/http`, `encoding/json`, `mime/multipart`).
- Router nativo do `http.ServeMux` (Go 1.22+ entende `"POST /rota"`).
- Docker multi-stage → imagem Alpine final (~15 MB).

## Rodar local (dev)

```powershell
cp .env.example .env   # preenche OPENROUTER_API_KEY e PINTEREST_ACCESS_TOKEN
go run ./cmd/server
```

Testa:

```powershell
curl http://localhost:8080/health
```

## Rodar via Docker

```bash
docker compose up --build
```

## Deploy na VPS

1. `scp -r api/ murilo@100.90.73.101:/home/murilo/`
2. Na VPS: preenche `/home/murilo/api/.env`
3. `cd /home/murilo/api && docker compose up -d --build`
4. Nginx: reverse proxy de `api.nuveafinds.com` → `http://127.0.0.1:8080`

## Estrutura

```
api/
├── cmd/server/main.go              # entrypoint
├── internal/
│   ├── config/                     # env -> struct
│   ├── httpx/                      # helpers JSON + middlewares
│   ├── ai/                         # cliente OpenRouter
│   ├── pinterest/                  # cliente Pinterest v5 + upload S3
│   └── handlers/                   # rotas HTTP
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── .env.example
└── README.md
```

## Variáveis obrigatórias

- `OPENROUTER_API_KEY` — https://openrouter.ai/keys
- `PINTEREST_ACCESS_TOKEN` — OAuth v5, scopes `pins:read pins:write boards:read`
- `PINTEREST_BOARD_MAP` — `slug1=boardID1,slug2=boardID2` (slugs definidos em `internal/ai/openrouter.go`).
