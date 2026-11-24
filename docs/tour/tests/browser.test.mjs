import test from 'node:test';
import assert from 'node:assert/strict';
import { chromium } from 'playwright';
import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';
import { extname } from 'node:path';
import { sources } from '../sources.mjs';

test('nine-slide tour works under a GitHub project Pages path on desktop and mobile', { timeout: 120000 }, async () => {
  const root = new URL('../', import.meta.url);
  const mime = { '.html': 'text/html', '.mjs': 'text/javascript', '.css': 'text/css', '.svg': 'image/svg+xml', '.png': 'image/png' };
  const server = createServer(async (req, res) => {
    try {
      const path = new URL(req.url, 'http://localhost').pathname;
      if (!path.startsWith('/socialpredict/')) { res.writeHead(404).end(); return; }
      const relative = decodeURIComponent(path.slice('/socialpredict/'.length)) || 'index.html';
      if (relative.includes('..')) { res.writeHead(400).end(); return; }
      const data = await readFile(new URL(relative, root));
      res.writeHead(200, { 'Content-Type': mime[extname(relative)] || 'application/octet-stream' }).end(data);
    } catch { res.writeHead(404).end(); }
  });
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  let browser;
  try {
    browser = await chromium.launch();
    const page = await browser.newPage({ viewport: { width: 1440, height: 1000 } });
    const errors = [];
    page.on('pageerror', error => errors.push(error.message));
    page.on('response', r => { if (r.status() >= 400) errors.push(`${r.status()}: ${r.url()}`); });
    const url = `http://127.0.0.1:${server.address().port}/socialpredict/`;
    const next = async () => {
      const current = await page.locator('.slide').getAttribute('data-slide');
      await page.getByRole('button', { name: 'Next slide', exact: true }).click();
      await page.locator(`.slide:not([data-slide="${current}"])`).waitFor();
    };
    await page.goto(url);
    assert.equal(await page.getByText('Nine slides. Real app captures. Click to try.', { exact: true }).count(), 0);
    assert.equal(await page.getByRole('heading', { name: 'Forecast together. Learn together.', exact: true }).isVisible(), true);
    await page.getByRole('button', { name: 'Try the tour', exact: true }).click();
    await page.getByLabel('Homepage title', { exact: true }).fill('<img src=x onerror=alert(1)>');
    await page.getByRole('button', { name: 'Save homepage', exact: true }).click();
    assert.equal(await page.locator('#home-preview-title img').count(), 0);
    assert.equal(await page.locator('#home-preview-title').textContent(), '<img src=x onerror=alert(1)>');
    await next();
    await page.getByLabel('Question', { exact: true }).fill('Will our lab publish a replication?');
    await page.getByRole('button', { name: 'Submit proposal', exact: true }).click();
    await page.getByRole('button', { name: 'Open review queue', exact: true }).click();
    await page.getByRole('button', { name: 'Reject & refund 10', exact: true }).click();
    assert.equal(await page.locator('#review-form .feedback').textContent(), 'Give the proposer a reason.');
    await page.getByLabel('Decision reason', { exact: true }).fill('Add the publication URL.');
    await page.getByRole('button', { name: 'Reject & refund 10', exact: true }).click();
    assert.match(await page.locator('.audit').textContent(), /500 available credits/);
    await page.getByRole('button', { name: 'Replay review', exact: true }).click();
    await page.getByRole('button', { name: 'Approve & publish', exact: true }).click();
    await next();
    assert.equal(await page.getByRole('button', { name: 'Will our lab publish a replication?', exact: true }).isVisible(), true);
    await page.getByRole('button', { name: 'Research & Replication', exact: true }).click();
    await page.getByLabel('Search questions', { exact: true }).fill('independent');
    assert.equal(await page.locator('#market-results tr').count(), 1);
    await page.getByLabel('Search questions', { exact: true }).fill('meetup');
    assert.match(await page.locator('#market-results').textContent(), /No matching/);
    await page.getByRole('button', { name: 'All topics', exact: true }).click();
    assert.equal(await page.locator('#market-results tr').count(), 2);
    await next();
    await page.getByRole('button', { name: 'Confirm YES purchase', exact: true }).click();
    assert.equal(await page.locator('#trade-credit').textContent(), '479');
    await page.getByLabel('Order amount', { exact: true }).fill('5');
    await page.getByRole('button', { name: 'NO', exact: true }).click();
    await page.getByRole('button', { name: 'Confirm NO purchase', exact: true }).click();
    assert.equal(await page.locator('#trade-credit').textContent(), '474');
    await next();
    assert.equal(await page.getByRole('button', { name: 'Confirm sale', exact: true }).isEnabled(), false);
    await page.getByRole('button', { name: 'Add recorded later trades', exact: true }).click();
    await page.getByRole('button', { name: 'Preview sale', exact: true }).click();
    await page.getByRole('button', { name: 'Confirm five-credit sale', exact: true }).click();
    assert.match(await page.locator('.sale-numbers').textContent(), /484/);
    assert.match(await page.locator('.sale-numbers').textContent(), /33/);
    await next();
    assert.equal(await page.locator('input[type=range]').count(), 0, 'Prices are read-only, not slider inputs');
    const prices = await page.locator('.group-market-price').allTextContents();
    await page.getByRole('button', { name: /Weekly meetup questions/ }).click();
    assert.deepEqual(await page.locator('.group-market-price').allTextContents(), prices);
    assert.match(await page.locator('.group-selected').textContent(), /Weekly meetup questions/);
    await page.getByText('Moderator: add an answer', { exact: true }).click();
    await page.getByLabel('Propose another answer', { exact: true }).fill('Public health forecasting');
    await page.getByRole('button', { name: 'Propose answer · 2 credits', exact: true }).click();
    await page.getByRole('button', { name: 'Approve queued answer', exact: true }).click();
    assert.equal(await page.getByRole('button', { name: /Public health forecasting/ }).isVisible(), true);
    await page.getByRole('button', { name: 'Browse appendix captures', exact: true }).click();
    for (const capture of sources) {
      await page.getByLabel('Choose an appendix screenshot', { exact: true }).selectOption(capture.id);
      await page.locator('#capture-image').evaluate(img => img.decode());
      assert.equal(await page.locator('#capture-image').evaluate(img => img.naturalWidth), 1440);
    }
    await page.keyboard.press('Escape');
    assert.equal(await page.locator('#evidence').evaluate(d => d.open), false);
    await next();
    assert.equal(await page.locator('#counter').textContent(), '09 / 09');
    const slides = ['overview','community','create','review','discover','trade','sell','groups','start'];
    for (const width of [375, 768, 1440]) {
      await page.setViewportSize({ width, height: 900 });
      for (const slide of slides) {
        await page.goto(`${url}#${slide}`);
        await page.locator(`[data-slide="${slide}"]`).waitFor();
        assert.equal(await page.getByRole('heading', { level: 1 }).count(), 1);
        assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), true, `${slide} overflows at ${width}px`);
      }
    }
    await page.getByRole('button', { name: 'Reset tour', exact: true }).click();
    await page.locator('[data-slide="overview"]').waitFor();
    await page.keyboard.press('ArrowRight');
    await page.locator('[data-slide="community"]').waitFor();
    await page.keyboard.press('End');
    await page.locator('[data-slide="start"]').waitFor();
    assert.deepEqual(errors, []);
  } finally {
    await browser?.close();
    await new Promise(resolve => server.close(resolve));
  }
});
