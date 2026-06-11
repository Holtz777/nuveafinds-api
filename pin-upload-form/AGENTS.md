# pin-upload-form/ — Pin Upload Pipeline UI

## Purpose

Browser-based 4-step form that drives the complete Pin creation pipeline: product input → AI generation → video upload → Pinterest publish. Communicates with the Go API via `fetch()`.

## Ownership

- `index.html` — single-file application (HTML + Tailwind CDN + vanilla JS)

## Local Contracts

### Pipeline steps
1. **Produto** — product name, affiliate link, image URL, influencer @, description, tags
2. **Gerar Pin** — calls `POST /pin-upload`, shows Version A (soft/beauty tone) and Version B (scroll-stopper) side by side, with AI-chosen board slug
3. **Video** — drag & drop .mp4, registers via `POST /pin-register-video`, uploads via `POST /proxy/upload-video`
4. **Publicar** — review data, publish via `POST /pin-publish`, show created pin link

### API base URL
Configurable via input field. Default: `http://localhost:8080`. In production: `https://api.nuveafinds.com`.

### Response format
Expects `{"status":"success","data":{...}}` or `{"status":"error","message":"..."}`.

## Work Guidance

### Frontend conventions
- **Vanilla JavaScript only** — no frameworks, no build step
- **Tailwind CSS via CDN** — no custom CSS files
- Use `fetch()` for all API calls
- Pipeline log with timestamps for user feedback
- Drag & drop with file size preview
- Error/success messages inline, not alerts

### State management
- Keep state in DOM elements and global variables (simple enough, no need for state library)
- Selected version (A/B) stored as `selectedVersion` variable
- `mediaId`, `uploadUrl`, `uploadParameters` passed between steps 3 and 4

### Pinterest OAuth helper
- Generates OAuth URL dynamically using `clientId` from user input
- Only used for initial token acquisition; auto-refresh handled server-side

### Testing locally
Open directly in browser: `file://` protocol works because API returns CORS `*` in dev mode.

## Verification

- Open `index.html` in browser
- Fill step 1 with test data, click "Gerar Pin" → AI returns two versions
- Drop a test .mp4 → registers + uploads
- Check pipeline log for errors

## Child DOX Index

No children. Single-file application.
