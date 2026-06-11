---
name: nuvea-vps
description: Infraestrutura da VPS Hostinger para Nuvea Finds — Nginx, Docker, domínios, SSL, acesso SSH. Use para troubleshooting de deploy, conectividade, ou modificações na VPS.
---

# Nuvea VPS — Skill de Infraestrutura

Infraestrutura da VPS Hostinger KVM 2 para o projeto Nuvea Finds: Nginx (site estático + reverse proxy API), Docker (API Go), SSL Let's Encrypt, domínios.

---

## 0. Identificação da VPS

| Campo | Valor |
|-------|-------|
| **Hostname** | `srv1295417.hstgr.cloud` |
| **IP público** | `72.60.67.123` |
| **IP Tailscale** | `100.90.73.101` |
| **SSH user** | `murilo` (admin), `root` (emergência) |
| **SO** | Ubuntu 24.04 LTS |
| **Plano** | KVM 2 (2 vCPU, 8 GB RAM, 100 GB SSD, 8 TB banda) |
| **Expira** | 2027-01-24 (renovação automática ativa) |
| **Localização** | USA - Boston |

## 1. Acesso SSH

```bash
# Via Tailscale (preferencial)
ssh murilo@100.90.73.101

# Via IP público (se Tailscale offline)
ssh murilo@72.60.67.123

# Via root (emergência — painel Hostinger mostra senha root)
ssh root@72.60.67.123
```

⚠️ **Se conexão recusada/timeout:**
1. Verificar se Tailscale está rodando na máquina local (`tailscale status`)
2. Acessar painel Hostinger → VPS → Modo de Emergência → console web
3. Rodar `ufw status` e `tailscale status` dentro da VPS
4. Hostinger não tem regras de firewall no painel (0 regras) — bloqueio seria UFW interno

## 2. Serviços na VPS

| Serviço | Descrição | Porta |
|---------|-----------|-------|
| **Nginx** | Site estático + reverse proxy API | 80, 443 |
| **API Go** (Docker) | Backend Pin automation | 127.0.0.1:8080 |
| **PostgreSQL 16** | Banco nativo (não usado ainda) | 5432 |
| **OpenClaw** | ⚠️ ISOLADO — não pode ser exposto publicamente | interno |
| **Docker** | Engine + Compose | — |
| **Certbot** | SSL Let's Encrypt automático | — |

## 3. Nginx — Configuração

### Site estático (`nuveafinds.com`, `www.nuveafinds.com`)

Servido de `/home/murilo/sites/nuveafinds`. Nginx serve arquivos estáticos diretamente.

### API reversa (`api.nuveafinds.com`)

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
    # SSL via Certbot
}
```

Config versionada em `api/deploy/nginx-api.conf`.

## 4. Docker — API Go

```bash
# Localizar diretório da API
cd ~/api    # ou ~/nuvea-api

# Subir
docker compose up -d --build

# Logs
docker compose logs -f

# Status
docker compose ps
```

## 5. Deploy — Procedimento

```bash
# 1. Copiar código pra VPS
scp -r api/ murilo@100.90.73.101:/home/murilo/

# 2. SSH e preencher .env
ssh murilo@100.90.73.101
cd ~/api
cp .env.example .env
nano .env   # preencher OPENROUTER_API_KEY, PINTEREST_ACCESS_TOKEN, PINTEREST_BOARD_MAP

# 3. Build e sobe
docker compose up -d --build
docker compose logs -f

# 4. Testar
curl https://api.nuveafinds.com/health
```

## 6. Domínios e DNS

| Tipo | Host | Valor |
|------|------|-------|
| A | @ | 72.60.67.123 |
| A | api | 72.60.67.123 |
| CNAME | www | nuveafinds.com |
| MX | @ | SMTP.GOOGLE.COM (prio 1) |

DNS gerenciado na Spaceship. Nameservers: `launch1.spaceship.net`, `launch2.spaceship.net`.

## 7. Troubleshooting

| Sintoma | Causa provável | Ação |
|---------|---------------|------|
| SSH timeout | Tailscale parado ou UFW bloqueando | Console emergência Hostinger → `ufw status` → liberar porta 22 |
| Site 500 | Permissão de travessia | `chmod o+x /home/murilo` |
| API unreachable | Container parou | `docker compose ps` → `docker compose up -d` |
| SSL expirado | Certbot não renovou | `sudo certbot renew --dry-run` |
