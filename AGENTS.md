# AGENTS.md — Nuvea Finds — DOX Rail

> **DOX framework installed.** Este arquivo é o DOX rail: instruções globais, preferências, regras de workflow, e o Child DOX Index. Child AGENTS.md em `api/`, `site-nuveafinds/`, `pin-upload-form/` contêm contratos locais.

## DOX Core Contract

- AGENTS.md files are binding work contracts for their subtrees.
- Work products, source materials, instructions, records, and durable docs must stay understandable from the nearest applicable AGENTS.md plus every parent AGENTS.md above it.

### Read Before Editing

1. Read this root AGENTS.md.
2. Identify every file or folder you expect to touch.
3. Walk from the repository root to each target path.
4. Read every AGENTS.md found along each route.
5. Use the nearest AGENTS.md as the local contract and parent docs for repo-wide rules.
6. If docs conflict, the closer doc controls local work details, but no child doc may weaken DOX.

Do not rely on memory. Re-read the applicable DOX chain in the current session before editing.

### Update After Editing

Every meaningful change requires a DOX pass before the task is done. Update the closest owning AGENTS.md when a change affects purpose, scope, ownership, contracts, workflows, or operating rules. Update parent docs when parent-level structure or child index changes.

---

Este arquivo contem TODA a informacao do projeto Nuvea Finds. Se voce e um agente de IA continuando este projeto, leia este arquivo inteiro antes de qualquer acao.

> **Importante:** o projeto NAO usa mais n8n nem nenhum script Python. A automacao e toda feita em **Go** (ver diretorio `api/`). O objetivo tambem e educativo: o usuario quer **aprender Go** enquanto o projeto cresce. Explique decisoes, escreva codigo idiomatico e prefira a biblioteca padrao sempre que possivel.

---

## 1. Visao Geral do Projeto

- **Objetivo:** Vender produtos como afiliado da Amazon para o mercado gringo (EUA, CA, UK) via trafego organico no Pinterest.
- **Marca:** Nuvea Finds
- **Estrategia principal:** Repostar videos virais de influencers do TikTok como Video Pins no Pinterest, com links de afiliado da Amazon. O processo e automatizado pela API Go em `api/`.
- **Nicho publico-alvo:** Mulheres (majoria do Pinterest) e publico geral de EUA, Canada e Europa.
- **Categorias de produto:** Beauty, Wellness, Home Gadgets, Self-Care, Viral Gadgets.

---

## 2. Stack Tecnica (atual)

### Backend: Go (`api/`)
- Linguagem: **Go 1.23+** (o host roda 1.26, build do Docker usa 1.23).
- Sem frameworks HTTP. Usa `net/http` + `http.ServeMux` (roteador nativo com suporte a `"POST /rota"` desde Go 1.22).
- Sem dependencias externas ainda (`go.mod` limpo). Tudo com biblioteca padrao.
- JSON: `encoding/json`. Multipart: `mime/multipart`. Contexto/timeouts: `context`.
- Deploy: **Docker multi-stage** (build em `golang:1.23-alpine`, runtime em `alpine:3.20`, binario estatico CGO_ENABLED=0).
- Orquestracao local e na VPS: **docker compose**.

### Integracoes
- **OpenRouter** (https://openrouter.ai) como provedor de IA unificado. Endpoint compativel com Chat Completions da OpenAI. Modelo atual configurado no .env: `google/gemini-3.1-flash-lite-preview` (configuravel via env, default no codigo `anthropic/claude-sonnet-4`).
- **Pinterest API v5** (https://developers.pinterest.com/docs/api/v5/) para upload de media e criacao de Video Pins.
- **Amazon Associates** (link de afiliado, Associate ID `nuveafinds-20`). Amazon PA API ainda **nao** integrada.

### Frontend
- Site estatico: `site-nuveafinds/` (HTML + Tailwind via CDN).
- Formulario de upload de pin: `pin-upload-form/index.html` — formulario completo em 4 passos (Produto -> Gerar Pin -> Video -> Publicar), HTML estatico com Tailwind, abre direto no browser apontando para a API local.

---

## 3. API Go (`api/`)

### 3.1 Estrutura

```
api/
├── cmd/server/main.go             # entrypoint: config -> deps -> http.Server + graceful shutdown
├── internal/
│   ├── config/config.go           # carrega .env (bufio.Scanner, stdlib) + env vars
│   ├── httpx/httpx.go             # helpers JSON + middlewares (CORS, Logger)
│   ├── ai/openrouter.go           # cliente OpenRouter + prompt de geracao de pins
│   ├── pinterest/pinterest.go     # cliente Pinterest v5 (media register, S3 upload, create pin, list boards)
│   └── handlers/handlers.go       # rotas HTTP
├── deploy/nginx-api.conf          # config Nginx da VPS para api.nuveafinds.com
├── Dockerfile                     # multi-stage build
├── docker-compose.yml             # servico unico "api"
├── go.mod                         # module github.com/Holtz777/nuveafinds-api
├── .env.example
├── .env                           # gitignored, contem as chaves reais
├── .gitignore
├── .dockerignore
└── README.md
```

### 3.2 Rotas

| Rota                       | Metodo | O que faz                                                                                                 |
| -------------------------- | ------ | --------------------------------------------------------------------------------------------------------- |
| `/health`                  | GET    | Healthcheck.                                                                                              |
| `/boards`                  | GET    | Lista boards do Pinterest + gera `PINTEREST_BOARD_MAP` pronto pra colar no .env.                          |
| `/pin-upload`              | POST   | Recebe dados do produto, chama OpenRouter, devolve 2 versoes (A/B) de titulo+descricao + slug da board.   |
| `/pin-register-video`      | POST   | Chama `POST /v5/media` no Pinterest; devolve `mediaId`, `uploadUrl`, `uploadParameters`.             |
| `/proxy/upload-video`      | POST   | Recebe o arquivo do browser + upload_url + params, faz o multipart no S3 do Pinterest server-side. |
| `/pin-publish`             | POST   | Espera a media ficar `succeeded`, cria o Video Pin (`POST /v5/pins`) com board, titulo, descricao, link. |

Formato de resposta: `{"status":"success","data":{...}}` ou `{"status":"error","message":"..."}`.

### 3.3 Variaveis de ambiente

Arquivo: `api/.env` (gitignored, modelo em `.env.example`).

> **IMPORTANTE:** O config carrega `.env` automaticamente via `loadDotEnv()` em `config/config.go`. Usa `bufio.Scanner` (stdlib pura). So seta vars que ainda nao existem no ambiente (vars de sistema tem prioridade).

| Variavel                | Obrigatoria | Descricao                                                                    |
| ----------------------- | ----------- | ---------------------------------------------------------------------------- |
| `PORT`                  | nao (8080)  | Porta HTTP.                                                                  |
| `OPENROUTER_API_KEY`    | **sim**     | Chave do OpenRouter (https://openrouter.ai/keys).                            |
| `OPENROUTER_MODEL`      | nao         | Default `anthropic/claude-sonnet-4`.                                         |
| `OPENROUTER_REFERER`    | nao         | Header `HTTP-Referer` para tracking no OpenRouter.                           |
| `OPENROUTER_TITLE`      | nao         | Header `X-Title` para tracking no OpenRouter.                                |
| `PINTEREST_ACCESS_TOKEN`| **sim p/ Pinterest** | Token OAuth v5 com scopes `pins:read pins:write boards:read boards:write`. |
| `PINTEREST_BOARD_MAP`   | **sim p/ publish** | Mapa `slug=boardID,slug2=boardID2`. Ja configurado (ver 7.1).           |
| `CORS_ORIGIN`           | nao         | `*` em dev, dominio do site em prod.                                         |

### 3.4 Boards (slugs internos)

Definidos em `api/internal/ai/openrouter.go` (`validBoards`):

- `viral-makeup-skincare-finds` - Makeup/Skincare viral
- `wellness-health-essentials` - Vitamins, supplements, fitness
- `amazon-home-finds-hacks` - Kitchen, organization, decor
- `aesthetic-self-care-routine` - Self-care, lifestyle
- `genius-gadgets-viral-finds` - Gadgets diversos / fallback

Descricoes SEO completas em `boards_seo.md`. A IA escolhe o slug; o handler `/pin-publish` traduz slug -> board ID via `PINTEREST_BOARD_MAP`.

### 3.5 Comandos uteis (local)

```powershell
cd api
go run ./cmd/server               # rodar em dev (carrega .env automaticamente)
go build ./...                    # verificar compilacao
go vet ./...                      # lint estatico
docker compose up --build         # rodar via Docker

# Testar boards
curl http://localhost:8080/boards

# Testar formulario
start ..\pin-upload-form\index.html
```

### 3.6 Deploy na VPS

```bash
# 1. Copia o codigo
scp -r api/ murilo@100.90.73.101:/home/murilo/

# 2. SSH e preenche .env
ssh murilo@100.90.73.101
cd ~/api
cp .env.example .env
nano .env   # preencher OPENROUTER_API_KEY, PINTEREST_ACCESS_TOKEN, PINTEREST_BOARD_MAP

# 3. Sobe
docker compose up -d --build
docker compose logs -f

# 4. Nginx ja esta configurado (ver 4.3). Testa:
curl https://api.nuveafinds.com/health
```

---

## 4. Infraestrutura

### 4.1 VPS (Hostinger KVM)

- **IP Publico:** 72.60.67.123
- **IP Tailscale:** 100.90.73.101
- **Usuario:** murilo
- **Home:** /home/murilo
- **Acesso SSH:** `ssh murilo@100.90.73.101` (via Tailscale; pode pedir auth no browser em nova sessao)
- **SO:** Ubuntu (kernel 6.8.0-106-generic x86_64)
- **Servicos rodando na VPS:**
  - Nginx (site estatico + reverse proxy da API Go)
  - OpenClaw (ISOLADO - NAO pode ser exposto publicamente)
  - PostgreSQL 16 (nativo, nao usado pela API ainda)
  - Docker + Docker Compose
  - Go toolchain instalada em `/home/murilo/go` (pode usar pra builds nativos)
  - Certbot (SSL automatico Let's Encrypt)

### 4.2 Dominio e DNS

- **Dominio:** nuveafinds.com (Spaceship, expira 2027-04-08, renovacao automatica)
- **Nameservers:** launch1.spaceship.net, launch2.spaceship.net
- **Email profissional:** contact@nuveafinds.com (Google Workspace)

| Tipo  | Host               | Valor                                                              |
| ----- | ------------------ | ------------------------------------------------------------------ |
| A     | @                  | 72.60.67.123                                                       |
| A     | api                | 72.60.67.123                                                       |
| CNAME | www                | nuveafinds.com                                                     |
| MX    | @                  | SMTP.GOOGLE.COM (prio 1)                                           |
| TXT   | @                  | google-site-verification=D9k30d5z6pvPK3hcX_g1dS03WOpJnroeH2WQ1G-aRwU |
| TXT   | google._domainkey  | DKIM completo (ver `dns_config_nuveafinds.md`)                     |

### 4.3 Nginx

**nuveafinds.com / www.nuveafinds.com** - Site estatico em `/home/murilo/sites/nuveafinds` (sem mudancas).

**api.nuveafinds.com** - Reverse proxy para a API Go (`127.0.0.1:8080`):

```nginx
server {
    server_name api.nuveafinds.com;
    client_max_body_size 2100m;
    proxy_request_buffering off;
    proxy_read_timeout 600s;
    proxy_send_timeout 600s;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    listen 443 ssl;
    # SSL via Certbot (api.nuveafinds.com, expira 2026-07-07)
}
```

Arquivo versionado em `api/deploy/nginx-api.conf`. Alteracoes: copiar pra VPS em `/etc/nginx/sites-available/api.nuveafinds.com`, rodar `sudo nginx -t && sudo systemctl reload nginx`.

### 4.4 GitHub

- **Site:** https://github.com/Holtz777/nuveafinds-site.git (branch `main`)
- **API:** ainda nao tem repo (module declarado como `github.com/Holtz777/nuveafinds-api`). Criar quando for hora do primeiro push.
- **Diretorio local do site:** `C:\Users\muril\renda-extra\site-nuveafinds`
- **Diretorio local da API:** `C:\Users\muril\renda-extra\api`

---

## 5. Site - Nuvea Finds

### 5.1 Stack

- HTML estatico
- Tailwind CSS via CDN
- CSS proprio em `assets/css/styles.css`

### 5.2 Arquivos principais

- `site-nuveafinds/index.html`
- `site-nuveafinds/assets/css/styles.css`
- `site-nuveafinds/assets/images/brand-logo.png` (logo transparente, usada no hero)
- `site-nuveafinds/assets/images/site-icon.png` (favicon simples)

### 5.3 Deploy

```bash
cd ~/sites/nuveafinds
git fetch origin
git reset --hard origin/main
chmod -R o+r /home/murilo/sites/nuveafinds
chmod o+x /home/murilo/sites/nuveafinds
```

### 5.4 Identidade visual

- **Paleta:**
  - Fundo cream: `#FAF7EB`
  - Verde sage: `#8BA88E`
  - Texto escuro: `#4A5550`
- **Tom:** clean, suave, feminino sofisticado, premium leve
- **Logo principal:** `brand-logo.png`
- **Favicon:** `site-icon.png`

### 5.5 Pendencias

- Confirmar email correto: `contact@nuveafinds.com` ou `contact@nueveafinds.com` (houve typo no footer)
- Inserir links reais de afiliado no `index.html` (alguns anchors ainda usam `#`)

---

## 6. Formulario de Upload (`pin-upload-form/`)

### 6.1 Arquivo
- `pin-upload-form/index.html` — formulario unico, HTML estatico + Tailwind CDN

### 6.2 Pipeline em 4 passos
1. **Produto:** titulo, link afiliado, URL da imagem, @ do influencer, descricao, tags
2. **Gerar Pin:** chama `/pin-upload`, mostra Versao A (tom suave/beauty) e Versao B (scroll-stopper) lado a lado, com board sugerida pela IA
3. **Video:** drag & drop do .mp4, registra no Pinterest via `/pin-register-video`, upload pro S3 via `/proxy/upload-video`
4. **Publicar:** confere dados, publica via `/pin-publish`, mostra link do pin criado

### 6.3 Funcionalidades
- API base URL configuravel (default: `http://localhost:8080`)
- Pipeline log em tempo real (com timestamp)
- Drag & drop de video com preview de tamanho
- Selecao de versao A/B com highlight visual
- Secao de Pinterest OAuth setup com link gerado automaticamente
- Mensagens de erro/sucesso inline

### 6.4 Uso local
Abre direto no browser: `start ..\pin-upload-form\index.html` (ou arrastar pro Chrome). Funciona via fetch com CORS `*`.

---

## 7. Pinterest

- **Perfil:** https://www.pinterest.com/nuveafinds/
- **Nome:** Nuvea Finds | Amazon & Viral Finds
- **Username:** NuveaFinds
- **Formato dos pins:** Video Pins verticais (1000x1500)
- **Tom visual:** Titulo grande e legivel, paleta da marca consistente, CTA discreto

### 7.1 Developer App

- **App ID:** 1560888
- **App Name:** Nuvea Finds Pin Manager
- **Company:** Nuvea Finds
- **Redirect URI:** `https://developers.pinterest.com/oauth/callback` (callback oficial do Pinterest Dev Platform)
- **Status:** **Trial** — upgrade para Standard Access solicitado em 2026-04-18 (aguardando aprovacao). Enquanto Trial, nao e possivel criar pins em producao, apenas no Sandbox.
- **Scopes atuais:** `pins:read pins:write boards:read boards:write`
- **Access token:** prefixo `pina_`, renovar a cada 30 dias via refresh token (prefixo `pinr_`)

### 7.2 Boards (mapeadas)

| Slug interno | Board ID | Nome no Pinterest |
|-------------|----------|-------------------|
| `viral-makeup-skincare-finds` | `1033787358154873174` | Viral Makeup & Skincare Finds 💄 |
| `wellness-health-essentials` | `1033787358154873175` | Wellness & Health Essentials 🌱 |
| `amazon-home-finds-hacks` | `1033787358154873176` | Amazon Home Finds & Hacks 🏡 |
| `aesthetic-self-care-routine` | `1033787358154873177` | Aesthetic Self-Care Routine ✨ |
| `genius-gadgets-viral-finds` | `1033787358154873180` | Genius Gadgets & Viral Finds 💡 |

Ignoradas (boards automaticas): "Salvamentos rapidos", "Products you tagged"

### 7.3 OAuth Flow (como gerar novo token)

1. Abrir no browser (substitua `APP_ID`):
   ```
   https://www.pinterest.com/oauth/?client_id=APP_ID&redirect_uri=https://developers.pinterest.com/oauth/callback&response_type=code&scope=pins:read,pins:write,boards:read,boards:write&state=nuveafinds
   ```
2. Copiar o `code` da URL de redirect
3. Trocar por token (o curl precisa de Basic Auth com base64 de `client_id:client_secret`):
   ```bash
   curl -X POST https://api.pinterest.com/v5/oauth/token \
     --header 'Authorization: Basic {base64(client_id:client_secret)}' \
     --header 'Content-Type: application/x-www-form-urlencoded' \
     --data-urlencode 'grant_type=authorization_code' \
     --data-urlencode 'code=CODE_AQUI' \
     --data-urlencode 'redirect_uri=https://developers.pinterest.com/oauth/callback'
   ```
4. O `access_token` (prefixo `pina_`) vai no `.env`. O `refresh_token` (prefixo `pinr_`) serve pra renovar sem re-autorizar.

### 7.4 Sandbox vs Producao

- **Sandbox:** `https://api-sandbox.pinterest.com/v5/` — funciona em Trial
- **Producao:** `https://api.pinterest.com/v5/` — precisa de Standard Access aprovado

---

## 8. Amazon Associates

- **Associate ID:** nuveafinds-20
- **Tipo de site cadastrado:** Content or Niche Website
- **Descricao cadastrada:** "Nuvea Finds is a curated product discovery website focused on beauty, wellness, and everyday lifestyle finds. Visitors use the site to browse handpicked Amazon product recommendations, discover trending and useful items, and access affiliate links to products that match their interests."

---

## 9. Pipeline (como funciona hoje)

```
[Voce seleciona na mao]
  - nome do produto (titulo da Amazon)
  - link afiliado Amazon
  - URL da imagem de capa
  - @ do influencer (TikTok)
  - descricao do produto (Amazon)
  - video .mp4 baixado do TikTok (yt-dlp/SnapTik)
        |
        v
[Formulario web - pin-upload-form/index.html]
        |
        | POST /pin-upload (JSON)
        v
[API Go] --OpenRouter--> 2 versoes (A/B) de titulo+descricao + slug da board
        |
        | POST /pin-register-video (JSON)
        v
[API Go] --Pinterest /v5/media--> {mediaId, uploadUrl, uploadParameters}
        |
        | POST /proxy/upload-video (multipart: file + upload_url + upload_parameters)
        v
[API Go] --S3 multipart (bufferizado, com Content-Length)--> Pinterest bucket (204)
        |
        | POST /pin-publish (JSON com mediaId, board slug, titulo, descricao, link)
        v
[API Go] polling em /v5/media/{id} ate status=succeeded, depois POST /v5/pins
        |
        v
[Pinterest] Pin publicado
```

### 9.1 Manual

- Pesquisar produtos na Amazon
- Gerar link de afiliado
- Buscar/baixar videos do TikTok

### 9.2 Automatico (via API Go + Formulario)

- Geracao de titulos, descricoes, escolha de board (IA)
- Registro de media, upload pro S3 do Pinterest (proxy server-side)
- Criacao do Video Pin

### 9.3 Status atual do fluxo

- [x] Geracao de titulos com IA (OpenRouter) — **funcionando local**
- [x] Upload de video pro S3 do Pinterest — **funcionando local**
- [ ] Publicacao do pin — **bloqueada**: app em Trial, aguardando aprovacao Standard Access
- [ ] Deploy na VPS — **pendente**

---

## 10. Seguranca

- **OpenClaw roda na mesma VPS** e NAO pode ser exposto publicamente. Nginx so serve o site estatico e a API Go.
- **API Go fica em `127.0.0.1:8080`** (bind apenas no loopback do container via `"127.0.0.1:8080:8080"` no compose). Acesso publico so por HTTPS via Nginx.
- **Tokens em `.env`** (gitignored). Nunca commitar.
- **Permissoes:** `/home/murilo` tem `o+x` para o Nginx poder entrar e servir o site estatico.

---

## 11. Preferencias do Usuario (importantes!)

- **Quer aprender Go** enquanto construi o projeto. Ao escrever codigo, comente decisoes relevantes e prefira abordagens idiomaticas da biblioteca padrao.
- **Nao usar n8n nem Python.** A stack de backend e Go. Apagar qualquer resquicio dessas tecnologias se aparecer.
- Prefere praticidade. Se fica irritado com solucoes excessivamente complexas para coisas simples.
- Sabe codar e trabalhar com AI Agents.
- Priorizar seguranca da VPS por causa do OpenClaw.
- Manter separacao clara entre branding assets e arquivos funcionais.
- Abordagem de entrada de dados: formulario web (HTML estatico) POST-ando direto nos endpoints Go.

---

## 12. Historico de Decisoes

1. **Docker na VPS:** instalado em 2026-04-08 via `get.docker.com`.
2. **Primeira iteracao foi em n8n + proxy FastAPI Python.** Abandonado em 2026-04-18: usuario quis codar tudo em Go para aprender e ter mais controle. n8n desinstalado (containers + imagens + volume + diretorio). Service systemd `pinterest-upload-proxy` removido.
3. **Favicon:** `brand-logo.png` (logo com texto) != `site-icon.png` (simbolo simples, legivel pequeno).
4. **Site 500 Error:** permissao de travessia em `/home/murilo`. Resolvido com `chmod o+x`.
5. **Email inconsistente:** `contact@nuveafinds.com` x `contact@nueveafinds.com`. Pendente confirmar.
6. **Router HTTP:** decidimos usar `http.ServeMux` nativo (Go 1.22+) em vez de chi/gin/echo. Menos deps, mais aprendizado.
7. **Provedor IA:** OpenRouter (endpoint compativel com OpenAI, permite trocar modelo trocando env var).
8. **.env loader:** Go nao carrega `.env` automaticamente (`os.Getenv` so le vars de ambiente do processo). Implementamos `loadDotEnv()` com `bufio.Scanner` (stdlib pura) em `config/config.go`. So seta vars que ainda nao existem (env do sistema tem prioridade).
9. **UploadToS3 bufferizado:** S3 do Pinterest exige `Content-Length` explicito. Trocamos de `io.Pipe` (streaming, sem tamanho) para `bytes.Buffer` que permite `req.ContentLength = int64(buf.Len())`. Desvantagem: carrega o video inteiro em memoria (ok pra videos de ate ~2GB).
10. **Pinterest OAuth:** usa Basic Auth header (base64 de `client_id:client_secret`), Content-Type `application/x-www-form-urlencoded`, e `--data-urlencode`. Codes sao de uso unico e expiram em ~10 minutos.
11. **PINTEREST_ACCESS_TOKEN opcional:** na falta do token, a API sobe com WARNING mas permite testar endpoints que nao dependem do Pinterest (ex: `/pin-upload` com IA).
12. **Rota /boards:** adicionada para listar boards do Pinterest e gerar o `PINTEREST_BOARD_MAP` pronto pra colar no `.env`. Usa `BoardMapInverse` (board ID -> slug) para mostrar slugs ja mapeados.

---

## 13. Proximos Passos

### Bloqueado — aguardando Pinterest Standard Access
- Aprovacao do upgrade Trial -> Standard no Pinterest Developer Platform
- Apos aprovacao: testar `/pin-publish` localmente, depois deploy na VPS

### Curto prazo
1. ~~Criar app no Pinterest Developer Platform~~ **FEITO** (App ID: 1560888)
2. ~~Obter access token OAuth v5~~ **FEITO** (scopes: pins:read pins:write boards:read boards:write)
3. ~~Listar boards e montar PINTEREST_BOARD_MAP~~ **FEITO** (5 boards mapeadas)
4. ~~Testar /pin-upload local~~ **FEITO** (gera titulos com IA)
5. ~~Testar /pin-register-video + upload-video local~~ **FEITO** (upload funciona)
6. ~~Criar formulario de upload~~ **FEITO** (`pin-upload-form/index.html`)
7. **Adicionar refresh de token** — access token expira em 30d, refresh em 60d. Sem refresh automatico, precisa re-fazer OAuth manual.
8. Confirmar email correto do footer do site.
9. **Deployar a API na VPS** (`docker compose up -d --build` em `~/api`) — apos aprovacao Pinterest.

### Medio prazo
10. Persistir historico de pins publicados (PostgreSQL ja instalado na VPS, criar schema simples).
11. Integrar Amazon PA API para buscar dados do produto (imagem, descricao) automaticamente.
12. Implementar downloader de TikTok em Go (ou shell-out pra yt-dlp).
13. Migrar site para gerador estatico (Astro/Eleventy) para facilitar updates automaticos de produtos.

### Longo prazo
14. Analytics (plausible/umami self-hosted na VPS).
15. Expandir para TikTok direto e Instagram.
16. CI/CD: GitHub Actions rodando `go vet`, `go test`, build da imagem e push automatico pra VPS.

---

## 14. Arquivos de Contexto Relacionados

- `boards_seo.md` - Nomes e descricoes SEO das boards do Pinterest
- `estrategia_pinterest.md` - Estrategia geral do Pinterest
- `dns_config_nuveafinds.md` - Registros DNS completos
- `prompt_gerador_pins.txt` - Prompt antigo (manual) para gerar titulos/descricoes; util como referencia mas a API Go ja tem o equivalente em `api/internal/ai/openrouter.go`
- `api/README.md` - Documentacao especifica da API
- `pin-upload-form/index.html` - Formulario de upload de pins (pipeline completo)


---

## 15. Child DOX Index

This root AGENTS.md is the DOX rail. Child AGENTS.md files own domain-specific instructions for their subtrees.

| Path | Scope | Description |
|------|-------|-------------|
| `api/AGENTS.md` | Go API backend | Pin automation pipeline, Pinterest v5, OpenRouter AI, OAuth token management |
| `site-nuveafinds/AGENTS.md` | Static marketing site | Landing page, Tailwind CDN, brand assets, affiliate links |
| `pin-upload-form/AGENTS.md` | Upload pipeline UI | 4-step browser form, vanilla JS, fetch() to API |

- `boards_seo.md`, `estrategia_pinterest.md`, `dns_config_nuveafinds.md` — Reference docs, covered by root.
- `prompt_gerador_pins.txt` — Legacy manual prompt (superseded by `api/internal/ai/openrouter.go`), kept for reference.