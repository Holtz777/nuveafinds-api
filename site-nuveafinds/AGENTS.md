# site-nuveafinds/ — Static Marketing Site

## Purpose

Static HTML site for Nuvea Finds brand. Serves as landing page and product discovery hub. Hosted on VPS via Nginx at `nuveafinds.com` and `www.nuveafinds.com`.

## Ownership

- `index.html` — main landing page with product showcase
- `privacy-policy.html` — privacy policy page
- `assets/css/styles.css` — custom CSS (brand-specific styles beyond Tailwind)
- `assets/images/brand-logo.png` — primary logo (transparent, used in hero)
- `assets/images/site-icon.png` — favicon

## Local Contracts

- Site is served from `/home/murilo/sites/nuveafinds` on the VPS
- Nginx serves static files directly (no reverse proxy)
- Permissions: `chmod -R o+r` on site dir, `chmod o+x` on `/home/murilo`

## Work Guidance

### Styling
- **Tailwind CSS via CDN** for all utility classes
- Custom CSS in `assets/css/styles.css` only for brand-specific needs not covered by Tailwind
- Brand palette: cream `#FAF7EB`, sage green `#8BA88E`, dark text `#4A5550`
- Tone: clean, soft, feminine, premium-light

### Deploy
```bash
cd ~/sites/nuveafinds
git fetch origin
git reset --hard origin/main
chmod -R o+r /home/murilo/sites/nuveafinds
chmod o+x /home/murilo/sites/nuveafinds
```

### Conventions
- No JavaScript frameworks. Vanilla HTML + Tailwind CDN.
- Amazon affiliate links use Associate ID `nuveafinds-20`.
- Email: `contact@nuveafinds.com` (NOT `nueveafinds.com` — there was a typo to fix).
- All product links must be real affiliate links, not `#` placeholders.

## Verification

- Open `index.html` in browser — page renders, Tailwind loads, links work
- Run link checker on affiliate URLs
- Verify SSL cert is valid: `curl -I https://nuveafinds.com`

## Child DOX Index

No children. Single-level static site structure.
