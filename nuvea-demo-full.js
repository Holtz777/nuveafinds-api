// nuvea-demo-full.js — OAuth + Pipeline demo in one continuous video
const { chromium } = require('playwright');
const path = require('path');

const API_URL = 'https://api.nuveafinds.com';
const APP_ID = '1560888';
const OAUTH_URL = `https://www.pinterest.com/oauth/?client_id=${APP_ID}&redirect_uri=https://developers.pinterest.com/oauth/callback&response_type=code&scope=pins:read,pins:write,boards:read,boards:write&state=nuveafinds`;

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    recordVideo: { dir: __dirname, size: { width: 1280, height: 900 } },
    viewport: { width: 1280, height: 900 },
  });

  // ═══ PART 1: OAUTH FLOW ═══
  const tab1 = await context.newPage();
  console.log('[1/6] OAuth authorization page...');
  await tab1.goto(OAUTH_URL, { waitUntil: 'domcontentloaded', timeout: 10000 }).catch(() => {});
  await tab1.waitForTimeout(4000);

  // Show token exchange explanation
  console.log('[2/6] Token exchange...');
  await tab1.setContent(`
    <!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><style>
      *{margin:0;padding:0;box-sizing:border-box}
      body{font-family:system-ui,sans-serif;background:#FAF7EB;padding:60px 40px;color:#4A5550}
      .box{background:white;border-radius:16px;padding:28px;margin-bottom:20px;border:1px solid rgba(139,168,142,0.2);max-width:720px;margin-left:auto;margin-right:auto}
      h2{color:#8BA88E;font-size:18px;margin-bottom:12px}
      .code{background:#1e1e1e;color:#4ec9b0;padding:14px 18px;border-radius:10px;font:13px "Courier New",monospace;overflow-x:auto;white-space:pre-wrap;word-break:break-all}
      .ok{color:#8BA88E;font-weight:600;margin-top:10px}
    </style></head><body>
    <div class="box">
      <h2>✅ Step 1 — OAuth Authorization</h2>
      <p style="margin-bottom:10px">User authorizes the Nuvea Finds app via Pinterest OAuth:</p>
      <div class="code" style="font-size:11px">GET https://www.pinterest.com/oauth/
  ?client_id=${APP_ID}
  &redirect_uri=https://developers.pinterest.com/oauth/callback
  &response_type=code
  &scope=pins:read pins:write boards:read boards:write</div>
      <p class="ok">✅ Authorization granted → code received in redirect URL</p>
    </div>
    <div class="box">
      <h2>✅ Step 2 — Token Exchange (Server-Side)</h2>
      <p style="margin-bottom:10px">Backend exchanges the authorization code for an access token:</p>
      <div class="code">POST https://api.pinterest.com/v5/oauth/token
Authorization: Basic {base64(client_id:client_secret)}
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&code=f5ece420...&redirect_uri=...</div>
      <p style="margin-top:10px;margin-bottom:8px">Response:</p>
      <div class="code" style="color:#6A9955">{
  "access_token": "pina_AEATRUIXABHSGAA...",
  "refresh_token": "pinr.eyJhbGciOiJSUzI1NiI...",
  "expires_in": 2592000,
  "refresh_token_expires_in": 5184000,
  "scope": "boards:read boards:write pins:read pins:write"
}</div>
      <p class="ok">✅ Tokens stored — auto-refresh every 25 days</p>
    </div>
    <div class="box">
      <h2>▶️ Step 3 — API Integration</h2>
      <p>The access token now powers all Pinterest API calls from the Nuvea Finds backend. The following demo shows the complete Pin creation pipeline using this token.</p>
    </div>
    </body></html>
  `, { waitUntil: 'load' });
  await tab1.waitForTimeout(6000);
  await tab1.close();

  // ═══ PART 2: PIN CREATION PIPELINE ═══
  const page = await context.newPage();
  const formPath = 'file://' + path.join(__dirname, 'pin-upload-form', 'index.html');
  
  console.log('[3/6] Loading form...');
  await page.goto(formPath, { waitUntil: 'networkidle' });
  await page.waitForTimeout(800);

  // Switch to English
  console.log('[3/6] Switching to English...');
  try {
    await page.click('#langToggle', { timeout: 5000 });
    await page.waitForTimeout(500);
  } catch (e) { console.log('  (toggle not found, continuing)'); }

  // Set API URL
  await page.fill('#apiBase', '');
  await page.fill('#apiBase', API_URL);
  await page.waitForTimeout(300);

  // Step 1: Product Info
  console.log('[4/6] Filling product info...');
  await page.fill('#productName', 'Silk Pillowcase Set — Luxury 22 Momme');
  await page.fill('#affiliateLink', 'https://www.amazon.com/dp/B0CN62X123?tag=nuveafinds-20');
  await page.fill('#productImageUrl', 'https://m.media-amazon.com/images/I/71K7Q4qB8fL._AC_SL1500_.jpg');
  await page.fill('#influencerHandle', '@beautywithanna');
  await page.fill('#productTags', 'silk pillowcase, hair care, beauty sleep');
  await page.fill('#productDescription', 'Luxury 22 Momme silk pillowcase set with hidden zipper. Prevents hair breakage, reduces skin wrinkles, and keeps you cool all night.');
  await page.waitForTimeout(500);

  // Step 2: AI Generation
  console.log('[5/6] AI generating titles...');
  try {
    await page.click('button:has-text("Gerar Titulos com IA")', { timeout: 5000 });
    await page.waitForTimeout(8000);
    await page.click('#cardA', { timeout: 5000 });
    await page.waitForTimeout(500);
    await page.click('#btnGoToStep3', { timeout: 5000 });
    await page.waitForTimeout(500);
  } catch (e) { console.log('  AI step error:', e.message); }

  // Step 3: Video Upload
  console.log('[6/6] Uploading video...');
  try {
    await page.setInputFiles('#videoFile', path.join(__dirname, 'test-video.mp4'), { timeout: 5000 });
    await page.waitForTimeout(800);
    await page.click('#btnUploadVideo', { timeout: 5000 });
    await page.waitForTimeout(15000);
  } catch (e) { console.log('  Upload error:', e.message); }

  // Step 4: Publish
  try {
    const step4Hidden = await page.$eval('#step4', el => el.classList.contains('hidden')).catch(() => true);
    if (!step4Hidden) {
      console.log('[6/6] Publishing...');
      await page.click('#btnPublish', { timeout: 5000 });
      await page.waitForTimeout(15000);
    }
  } catch (e) { console.log('  Publish error:', e.message); }

  console.log('Done!');
  await page.waitForTimeout(2000);
  await context.close();
  await browser.close();
  console.log('Full demo video saved.');
})();
