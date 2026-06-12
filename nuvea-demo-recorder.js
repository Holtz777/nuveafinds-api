// nuvea-demo-recorder.js — Grava demo video para Pinterest (OAuth + API integration)
const { chromium } = require('playwright');
const path = require('path');

const API_URL = 'https://api.nuveafinds.com';

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    recordVideo: { dir: __dirname, size: { width: 1280, height: 900 } },
    viewport: { width: 1280, height: 900 },
  });
  const page = await context.newPage();

  try {
    // Load form
    console.log('[0/5] Loading form...');
    await page.goto('file://' + path.join(__dirname, 'pin-upload-form', 'index.html'), { waitUntil: 'networkidle' });
    await page.waitForTimeout(500);

    // Switch to English + set sandbox mode
    console.log('[0/5] Switching to English...');
    await page.click('#langToggle');
    await page.waitForTimeout(500);

    // Set API URL
    await page.fill('#apiBase', '');
    await page.fill('#apiBase', API_URL);
    // Enable sandbox toggle
    const sandboxOn = await page.$eval('#sandboxToggle', el => el.classList.contains('bg-nuvea-green'));
    if (!sandboxOn) await page.click('#sandboxToggle');
    await page.waitForTimeout(500);

    // Step 1: Product Info
    console.log('[1/5] Filling product info...');
    await page.fill('#productName', 'Silk Pillowcase Set - Luxury 22 Momme');
    await page.fill('#affiliateLink', 'https://www.amazon.com/dp/B0CN62X123?tag=nuveafinds-20');
    await page.fill('#productImageUrl', 'https://m.media-amazon.com/images/I/71K7Q4qB8fL._AC_SL1500_.jpg');
    await page.fill('#influencerHandle', '@beautywithanna');
    await page.fill('#productTags', 'silk pillowcase, hair care, beauty sleep');
    await page.fill('#productDescription', 'Luxury 22 Momme silk pillowcase. Prevents hair breakage and skin wrinkles.');
    await page.waitForTimeout(500);

    // Step 2: AI Generation
    console.log('[2/5] Generating AI titles...');
    await page.click('button:has-text("Gerar Titulos com IA")');
    await page.waitForTimeout(6000);
    await page.click('#cardA');
    await page.waitForTimeout(300);

    // Step 3: Video Upload
    console.log('[3/5] Uploading video...');
    await page.click('#btnGoToStep3');
    await page.waitForTimeout(500);
    await page.setInputFiles('#videoFile', path.join(__dirname, 'test-video.mp4'));
    await page.waitForTimeout(500);
    await page.click('#btnUploadVideo');
    console.log('[3/5] Waiting for upload to complete...');
    await page.waitForTimeout(12000);

    // Step 4: Publish
    console.log('[4/5] Publishing...');
    const step4Hidden = await page.$eval('#step4', el => el.classList.contains('hidden')).catch(() => true);
    if (!step4Hidden) {
      await page.click('#btnPublish');
      await page.waitForTimeout(12000);
    }

    console.log('[5/5] Done! Keeping visible for 2s...');
    await page.waitForTimeout(2000);
  } catch (err) {
    console.error('Error:', err.message);
  } finally {
    await context.close();
    await browser.close();
    console.log('Video saved to:', path.join(__dirname, '*.webm'));
  }
})();
