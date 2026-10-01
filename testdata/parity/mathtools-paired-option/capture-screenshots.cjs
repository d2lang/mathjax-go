const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

(async () => {
  const dir = path.resolve(process.argv[2]);
  const title = process.argv[3];
  const names = ['before', 'after', 'original'];
  const svgs = Object.fromEntries(names.map(n => [n, fs.readFileSync(path.join(dir, n + '.svg'))]));
  if (!svgs.after.equals(svgs.original)) throw Error('Fixed D2 SVG must equal independent original reference');
  if (svgs.before.equals(svgs.original)) throw Error('Witness must demonstrate a change');
  const escape = s => s.replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;');
  const source = fs.readFileSync(path.join(dir, 'witness.d2'), 'utf8');
  const labels = ['Before · Go', 'After · Go fix', 'Expected · original MathJax 3.2.2'];
  const panels = names.map((n, i) => {
    const errors = [...svgs[n].toString().matchAll(/data-mjx-error="([^"]*)"/g)].map(m => m[1]);
    const note = errors.length ? '<p style="padding:0 18px;font-size:16px">Reported: <code>' + escape(errors.join('; ')) + '</code></p>' : '';
    return `<section><h2>${labels[i]}</h2><div class="image"><img src="data:image/svg+xml;base64,${svgs[n].toString('base64')}"></div>${note}</section>`;
  }).join('');
  const html = `<!doctype html><meta charset="utf-8"><style>*{box-sizing:border-box}body{margin:0;background:#eef2f7;color:#172033;font-family:Arial,sans-serif}#comparison{width:1500px;padding:30px}h1{font-size:30px;margin:0 0 12px}p{font-size:18px;line-height:1.5;color:#526178;margin:0 0 24px}.panels{display:grid;grid-template-columns:repeat(3,1fr);gap:18px}section{background:white;border:1px solid #d8e0eb;border-radius:10px;overflow:hidden}h2{font-size:19px;padding:18px;margin:0;border-bottom:1px solid #e5eaf1}.image{height:360px;display:flex;align-items:center;justify-content:center;padding:20px}img{display:block;max-width:none}pre{font-size:17px;line-height:1.5;white-space:pre-wrap;overflow-wrap:anywhere;margin:24px 0;padding:20px;background:white;border:1px solid #d8e0eb;border-radius:10px}</style><div id="comparison"><h1>${escape(title)}</h1><p>Same D2 source and renderer, all three complete SVGs at one shared scale.</p><div class="panels">${panels}</div><pre>${escape(source.trim())}</pre><p>The fixed SVG matches the independently rendered original byte for byte.</p></div>`;
  const executablePath = process.env.CHROMIUM_EXECUTABLE;
  const browser = await chromium.launch({headless: true, ...(executablePath ? {executablePath} : {}), args: ['--disable-gpu']});
  try {
    const page = await browser.newPage({viewport: {width: 1500, height: 1400}, deviceScaleFactor: 2});
    await page.setContent(html);
    await page.locator('img').evaluateAll(imgs => Promise.all(imgs.map(i => i.decode())));
    const scale = await page.locator('img').evaluateAll(imgs => {
      const factor = Math.min(425 / Math.max(...imgs.map(i => i.naturalWidth)), 320 / Math.max(...imgs.map(i => i.naturalHeight)));
      for (const i of imgs) { i.style.width = (i.naturalWidth * factor) + 'px'; i.style.height = (i.naturalHeight * factor) + 'px'; }
      return factor;
    });
    const fits = await page.locator('img').evaluateAll(imgs => imgs.every(i => {
      const r = i.getBoundingClientRect(), p = i.parentElement.getBoundingClientRect();
      return r.left >= p.left && r.top >= p.top && r.right <= p.right && r.bottom <= p.bottom;
    }));
    if (!fits) throw Error('Clipped SVG');
    await page.evaluate(() => document.fonts.ready);
    await page.evaluate(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))));
    fs.writeFileSync(path.join(dir, 'comparison.html'), await page.content());
    await page.locator('#comparison').screenshot({path: path.join(dir, 'comparison.png')});
    for (const n of names) {
      const isolated = await browser.newPage({viewport: {width: 1500, height: 1400}, deviceScaleFactor: 2});
      try {
        await isolated.setContent('<style>body{margin:0;background:white}img{display:block}</style><img src="data:image/svg+xml;base64,' + svgs[n].toString('base64') + '">');
        await isolated.locator('img').evaluate(i => i.decode());
        await isolated.evaluate(() => document.fonts.ready);
        await isolated.evaluate(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))));
        await isolated.locator('img').screenshot({path: path.join(dir, n + '.png')});
      } finally { await isolated.close(); }
    }
    if (!fs.readFileSync(path.join(dir, 'after.png')).equals(fs.readFileSync(path.join(dir, 'original.png')))) throw Error('PNG reference differs');
    console.log(JSON.stringify({scale, fits, afterOriginalSVGExact: true, afterOriginalPNGExact: true}));
  } finally { await browser.close(); }
})().catch(e => { console.error(e); process.exitCode = 1; });
