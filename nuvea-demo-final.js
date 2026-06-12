// nuvea-demo-final.js — OAuth summary + Pipeline in one video
const { chromium } = require('playwright');
const path = require('path');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    recordVideo: { dir: __dirname, size: { width: 1280, height: 900 } },
    viewport: { width: 1280, height: 900 },
  });

  // Part 1 — OAuth summary (static page)
  const p1 = await context.newPage();
  console.log('[1] OAuth summary...');
  await p1.goto('file://' + path.join(__dirname, 'oauth-demo.html'), { waitUntil: 'domcontentloaded' });
  await p1.waitForTimeout(5000);
  await p1.close();

  // Part 2 — Pinterest auth screen
  const p2 = await context.newPage();
  console.log('[2] Pinterest OAuth screen...');
  await p2.goto(`https://www.pinterest.com/oauth/?client_id=1560888&redirect_uri=https://developers.pinterest.com/oauth/callback&response_type=code&scope=pins:read,pins:write,boards:read,boards:write&state=nuveafinds`, { waitUntil: 'domcontentloaded', timeout: 8000 }).catch(() => {});
  await p2.waitForTimeout(4000);
  await p2.close();

  // Part 3 — Pipeline
  const page = await context.newPage();
  const formPath = 'file://' + path.join(__dirname, 'pin-upload-form', 'index.html');
  console.log('[3] Loading form...');
  await page.goto(formPath, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(1000);

  // Switch to English
  console.log('[3] EN toggle...');
  await page.click('#langToggle', { timeout: 5000 }).catch(() => {});
  await page.waitForTimeout(600);

  // API URL
  await page.fill('#apiBase', '');
  await page.fill('#apiBase', 'https://api.nuveafinds.com');
  await page.waitForTimeout(300);

  // Fill product
  console.log('[4] Product info...');
  await page.fill('#productName', 'Silk Pillowcase Set — Luxury 22 Momme');
  await page.fill('#affiliateLink', 'https://www.amazon.com/dp/B0CN62X123?tag=nuveafinds-20');
  await page.fill('#productImageUrl', 'https://m.media-amazon.com/images/I/71K7Q4qB8fL._AC_SL1500_.jpg');
  await page.fill('#influencerHandle', '@beautywithanna');
  await page.fill('#productTags', 'silk pillowcase, hair care, beauty sleep');
  await page.fill('#productDescription', 'Luxury 22 Momme silk pillowcase set. Prevents hair breakage and skin wrinkles.');
  await page.waitForTimeout(500);

  // AI generation
  console.log('[5] AI generating...');
  await page.click('button[onclick="goToStep2()"]', { timeout: 5000 }).catch(() => {});
  await page.waitForTimeout(8000);
  await page.click('#cardA', { timeout: 5000 }).catch(() => {});
  await page.waitForTimeout(500);
  await page.click('#btnGoToStep3', { timeout: 5000 }).catch(() => {});
  await page.waitForTimeout(500);

  // Upload
  console.log('[6] Uploading...');
  await page.setInputFiles('#videoFile', path.join(__dirname, 'test-video.mp4'), { timeout: 5000 }).catch(() => {});
  await page.waitForTimeout(800);
  await page.click('#btnUploadVideo', { timeout: 5000 }).catch(() => {});
  await page.waitForTimeout(15000);

  // Publish
  console.log('[7] Publishing...');
  try {
    const h = await page.$eval('#step4', el => el.classList.contains('hidden'));
    if (!h) { await page.click('#btnPublish', { timeout: 5000 }).catch(() => {}); await page.waitForTimeout(15000); }
  } catch(e) {}

  console.log('Done!');
  await page.waitForTimeout(2000);
  await context.close();
  await browser.close();
  console.log('Video saved.');
})();
