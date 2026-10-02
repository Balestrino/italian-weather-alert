// Only the disposable fixture served by TestTerritoryDashboardBrowser may be used.
const assert = require('node:assert/strict');
const { chromium } = require(process.env.IWA_PLAYWRIGHT_MODULE || 'playwright');
const base = new URL(process.argv[2]);
assert.equal(base.protocol, 'http:');
assert.ok(['127.0.0.1', 'localhost', '[::1]'].includes(base.hostname));

(async () => {
  const browser = await chromium.launch({ headless: true });
  try {
    const context = await browser.newContext({ javaScriptEnabled: false, viewport: { width: 390, height: 844 } });
    await context.route('**/*', route => new URL(route.request().url()).origin === base.origin ? route.continue() : route.abort());
    const page = await context.newPage();
    const visit = async path => { const response = await page.goto(new URL(path, base).href); assert.equal(response.status(), 200); };
    const fits = async () => {
      const layout = await page.evaluate(() => ({ width: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth, overflow: [...document.querySelectorAll('main,form,.table-wrap,.subnav')].map(e=>({tag:e.tagName,cls:e.className,width:e.getBoundingClientRect().width,right:e.getBoundingClientRect().right})) }));
      assert.ok(layout.scroll <= layout.width + 2, `page overflows viewport at ${page.url()}: ${JSON.stringify(layout)}`);
    };
    const keyboardToggle = async (action, name) => {
      const form = page.locator(`form[action="${action}"]`);
      const summary = form.locator('..').locator('summary');
      await summary.focus();
      await page.keyboard.press('Enter');
      assert.equal(await form.isVisible(), true);
      await page.keyboard.press('Tab');
      assert.equal(await page.locator(':focus').getAttribute('name'), 'actor');
      await page.keyboard.type('synthetic-browser');
      await page.keyboard.press('Tab');
      assert.equal(await page.locator(':focus').textContent(), name);
      await Promise.all([page.waitForNavigation(), page.keyboard.press('Enter')]);
    };
    await visit('/admin/regions?state=all');
    await fits();
    await keyboardToggle('/admin/regions/09/enabled', 'Abilita regione Toscana');
    assert.equal(new URL(page.url()).searchParams.get('state'), 'all');
    const row = page.locator('tr').filter({ has: page.locator('a[href="/admin/regions/09"]') });
    assert.match(await row.textContent(), /Abilitata/);
    await Promise.all([page.waitForNavigation(), row.getByRole('link', { name: 'Configura regione' }).press('Enter')]);
    await Promise.all([page.waitForNavigation(), page.getByRole('link', { name: 'Comuni', exact: true }).press('Enter')]);
    await fits();
    const next = page.getByRole('link', { name: 'Comuni successivi →' });
    assert.equal(await next.count(), 1);
    await Promise.all([page.waitForNavigation(), next.press('Enter')]);
    const retained = page.url();
    assert.ok(new URL(retained).searchParams.get('after'));
    const form = page.locator('form[action$="/enabled"]').first();
    const action = await form.getAttribute('action');
    const name = await form.locator('button').textContent();
    await keyboardToggle(action, name);
    assert.equal(page.url(), retained, 'municipal action lost pagination');
    const current = page.locator(`form[action="${action}"]`);
    assert.equal(await current.locator('[name=enabled]').getAttribute('value'), 'false');
    await fits();
    // 200% rendering preserves readable content, independently scrollable tables and native controls.
    await page.evaluate(() => { document.documentElement.style.zoom = '2'; });
    await fits();
    await current.locator('..').locator('summary').focus();
    await page.keyboard.press('Enter');
    assert.equal(await current.locator('input[name=actor]').isVisible(), true);
    await page.evaluate(() => { document.documentElement.style.zoom = ''; });
    await visit('/admin/regions/09?tab=history&kind=configuration');
    assert.match(await page.locator('main').textContent(), /synthetic-browser/);
    await visit('/admin/regions?state=all');
    await keyboardToggle('/admin/regions/09/enabled', 'Disabilita regione Toscana');
    await visit(new URL(retained).pathname + new URL(retained).search);
    assert.match(await page.locator('main').textContent(), /Sospeso: regione disabilitata/);
    assert.equal(await page.locator(`form[action="${action}"] [name=enabled]`).getAttribute('value'), 'false', 'parent reset child choice');
    await context.close();
    console.log('Territorial controls passed: native forms, keyboard, no JavaScript, pagination, 390px, 200% zoom and audit history.');
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
