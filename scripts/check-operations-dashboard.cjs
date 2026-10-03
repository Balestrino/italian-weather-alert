// Synthetic dashboard browser acceptance; optional screenshots stay outside Git.
const { chromium } = require(process.env.IWA_PLAYWRIGHT_MODULE || 'playwright');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

(async () => {
  const browser = await chromium.launch({headless: true});
  try {
    for (const javaScriptEnabled of [true, false]) {
      const context = await browser.newContext({javaScriptEnabled});
      const page = await context.newPage();
      const failures = [];
      page.on('pageerror', e => failures.push(e.message));
      page.on('console', m => { if (m.type() === 'error') failures.push(m.text()); });
      page.on('request', r => assert.equal(new URL(r.url()).origin, new URL(process.argv[2]).origin));
      for (const width of [1440, 768, 390]) {
        await page.setViewportSize({width, height: 1000});
        const response = await page.goto(process.argv[2] + '/admin/operations/?period=7d');
        assert.equal(response.status(), 200);
        assert.match(response.headers()['content-security-policy'], /default-src 'none'/);
        assert.equal(await page.locator('select[name="period"]').inputValue(), '7d');
        assert.equal(await page.locator('.job-metrics .metric').count(), 6);
        assert.ok(await page.locator('svg[role="img"]').isVisible());
        assert.ok(await page.locator('rect.chart-success').first().evaluate(e => e.getBoundingClientRect().height > 0));
        assert.equal(await page.locator('rect.chart-success').first().evaluate(e => getComputedStyle(e).fill), 'rgb(126, 226, 168)');
        assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), `page overflow at ${width}`);
        await page.getByText('Valori esatti dello storico · tabella accessibile', {exact: true}).click();
        assert.ok(await page.getByRole('region', {name: 'Valori dello storico', exact: true}).isVisible());
        await page.selectOption('select[name="period"]', '30d');
        await Promise.all([page.waitForURL('**/admin/operations/?period=30d'), page.getByRole('button', {name: 'Mostra storico'}).click()]);
        assert.equal(await page.locator('select[name="period"]').inputValue(), '30d');
        if (process.env.IWA_DASHBOARD_SCREENSHOTS) {
          fs.mkdirSync(process.env.IWA_DASHBOARD_SCREENSHOTS, {recursive: true});
          await page.screenshot({path: path.join(process.env.IWA_DASHBOARD_SCREENSHOTS, `dashboard-${width}-${javaScriptEnabled}.png`), fullPage: true});
        }
      }
      // Browser zoom 200% halves the CSS viewport of a 1440x1000 window.
      await page.setViewportSize({width: 720, height: 500});
      assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), 'page overflow at effective 200% zoom');
      const link = page.locator('a[href*="kind=ocr_resource"][href*="state=queued"]').first();
      await Promise.all([page.waitForURL('**/admin/operations/jobs?**'), link.click()]);
      assert.equal(new URL(page.url()).searchParams.get('kind'), 'ocr_resource');
      assert.equal(new URL(page.url()).searchParams.get('queue'), 'inference');
      assert.equal(new URL(page.url()).searchParams.get('state'), 'queued');
      assert.equal(failures.length, 0, failures.join('\n'));
      await context.close();
    }
    console.log('Operations dashboard: desktop/tablet/mobile, 200% zoom, chart, native periods, drilldowns, no JS, CSP and local assets passed.');
  } finally {
    await browser.close();
  }
})().catch(e => {console.error(e); process.exit(1);});
