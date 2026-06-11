---
name: nuvea-pinterest
description: Pinterest Developer Platform para Nuvea Finds — OAuth flow, App config, Sandbox vs Produção, demo video requirements, board mapping. Use para resolver problemas de acesso à API do Pinterest.
---

# Nuvea Pinterest — Skill da Plataforma Pinterest

Integração com Pinterest API v5 para Nuvea Finds: App Developer, OAuth, boards, sanbox vs produção.

---

## 0. App Configuration

| Campo | Valor |
|-------|-------|
| **App ID** | 1560888 |
| **App Name** | Nuvea Finds Pin Manager |
| **Company** | Nuvea Finds |
| **Redirect URI** | `https://developers.pinterest.com/oauth/callback` |
| **Status** | **Trial** 🔴 |
| **Scopes** | `pins:read pins:write boards:read boards:write` |

## 1. Acesso — Bloqueio atual

**Standard Access negado em ~2026-06-11.** Motivos citados pelo Pinterest:

1. ❌ Demo did not show Pinterest integration
2. ❌ Demo did not show full OAuth flow

**Ação requerida:** Gravar novo vídeo de demonstração mostrando:
- Fluxo OAuth completo (autorização → token → uso)
- Integração real com a API (upload de mídia, criação de pin)
- App em funcionamento integrado ao Pinterest

**Guidelines do Pinterest (inferido):**
- Mostrar a tela de autorização OAuth
- Mostrar o token sendo usado em chamadas de API
- Mostrar resultado final (pin criado no perfil)
- Deve ser um screencast real, não slides

## 2. Boards Mapeadas

| Slug interno | Board ID | Nome no Pinterest |
|---|---|---|
| `viral-makeup-skincare-finds` | `1033787358154873174` | Viral Makeup & Skincare Finds 💄 |
| `wellness-health-essentials` | `1033787358154873175` | Wellness & Health Essentials 🌱 |
| `amazon-home-finds-hacks` | `1033787358154873176` | Amazon Home Finds & Hacks 🏡 |
| `aesthetic-self-care-routine` | `1033787358154873177` | Aesthetic Self-Care Routine ✨ |
| `genius-gadgets-viral-finds` | `1033787358154873180` | Genius Gadgets & Viral Finds 💡 |

## 3. OAuth Flow

### Autorização (obter code)

```
https://www.pinterest.com/oauth/?client_id=1560888&redirect_uri=https://developers.pinterest.com/oauth/callback&response_type=code&scope=pins:read,pins:write,boards:read,boards:write&state=nuveafinds
```

Abrir no browser → autorizar → copiar `code` da URL de redirect.

### Trocar code por token

```bash
curl -X POST https://api.pinterest.com/v5/oauth/token \
  --header 'Authorization: Basic {base64(client_id:client_secret)}' \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=authorization_code' \
  --data-urlencode 'code=CODE_AQUI' \
  --data-urlencode 'redirect_uri=https://developers.pinterest.com/oauth/callback'
```

Resposta contém:
- `access_token` (prefixo `pina_`) — 30 dias
- `refresh_token` (prefixo `pinr_`) — 60 dias

### Configurar no .env

```env
PINTEREST_ACCESS_TOKEN=pina_...
PINTEREST_CLIENT_ID=1560888
PINTEREST_CLIENT_SECRET=...
PINTEREST_REFRESH_TOKEN=pinr_...
PINTEREST_SANDBOX=true   # true enquanto Trial
```

## 4. Sandbox vs Produção

| Modo | URL base | Funciona em Trial? |
|------|----------|-------------------|
| Sandbox | `https://api-sandbox.pinterest.com/v5/` | ✅ Sim |
| Produção | `https://api.pinterest.com/v5/` | ❌ Precisa Standard Access |

⚠️ **Enquanto Trial, tudo deve ser testado com `PINTEREST_SANDBOX=true`.**

### Listar boards (sandbox)

```bash
curl -H "Authorization: Bearer pina_..." \
  https://api-sandbox.pinterest.com/v5/boards/
```

A API Go tem a rota `GET /boards` que faz isso automaticamente e gera o `PINTEREST_BOARD_MAP`.

## 5. Perfil Pinterest

- **URL:** https://www.pinterest.com/nuveafinds/
- **Nome:** Nuvea Finds | Amazon & Viral Finds
- **Username:** NuveaFinds
- **Formato dos pins:** Video Pins verticais (1000x1500)

## 6. Demo Video — Requisitos para Reaplicação

O vídeo PRECISA mostrar:

1. **Tela de login/autorização Pinterest** — OAuth dialog com scopes visíveis
2. **Redirect com code** — URL com `?code=...` após autorização
3. **Troca code→token** — Chamada curl ou código mostrando token recebido
4. **Uso do token** — Chamada real de API (ex: listar boards, registrar mídia)
5. **Resultado final** — Pin publicado no perfil do Pinterest (sandbox ok)
6. **Fluxo completo sem cortes** — não pular etapas

Sugestão: gravar usando o formulário `pin-upload-form/index.html` + API Go local, mostrando:
- Step 1: preencher dados do produto
- Step 2: mostrar títulos gerados pela IA
- Step 3: upload do vídeo (mostrar `/pin-register-video` + `/proxy/upload-video`)
- Step 4: publicar (mostrar pin criado no perfil sandbox)
