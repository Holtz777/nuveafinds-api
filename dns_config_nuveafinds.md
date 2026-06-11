# Resumo de Configuração de Domínio e Email (nuveafinds.com)

## Dados do Domínio
- **Domínio:** nuveafinds.com
- **Registro expira em:** 2027-04-08
- **Renovação automática:** ativada
- **Privacidade:** alta
- **Nameservers:**
  - launch1.spaceship.net
  - launch2.spaceship.net

## Registros DNS

### A Record (Site)
| Tipo | Host | Valor | TTL |
|------|------|-------|-----|
| A | @ | 72.60.67.123 | 1800 |

### CNAME (www)
| Tipo | Host | Valor | TTL |
|------|------|-------|-----|
| CNAME | www | nuveafinds.com | 1800 |

### MX Record (Email Google Workspace)
| Tipo | Host | Prioridade | Destino | TTL |
|------|------|-----------|---------|-----|
| MX | @ | 1 | SMTP.GOOGLE.COM | 1800 |

### TXT Records (Email + Verificação)
| Tipo | Host | Valor | TTL |
|------|------|-------|-----|
| TXT | @ | google-site-verification=D9k30d5z6pvPK3hcX_g1dS03WOpJnroeH2WQ1G-aRwU | 1800 |
| TXT | google._domainkey | v=DKIM1;k=rsa;p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAlAKxRCFX//alAGNG1PbgjQjFWU0vquavn071hjXiU0ri9vcV1/RY8tE93GbLw4MyaKIbNr5viZtBhn/0aNGJpcAM6VFSytHFZswDwU0jJ5R9jK2rkpdycoePt43MlBFZdvSPGOtt0/IJ0eewwJPDIdm+tnGOT38YGnI1JzPb+jU2ys0D0iDdEP1SAkGAnLusCbSd1jz6fJC+J+NFiIBuPmecCTuYgNK3l/H5dmLgFHiOEgVhBf7SsLtt5eZSN/YboCtnfXTkFBISnikCNnX90619OZov3sf2bYfFN0ttwLJsvS4/Mv/Dr/KuedqEfJ1eW9sL7JwCnmwhCqPeTupdKwIDAQAB | 1800 |

## E-mail Profissional
- **Endereço Principal:** contact@nuveafinds.com
- **Provedor:** Google Workspace

## Próximo Passo
Configurar Nginx na VPS para servir o site em nuveafinds.com
