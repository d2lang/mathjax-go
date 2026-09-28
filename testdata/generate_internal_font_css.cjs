// SPDX-License-Identifier: Apache-2.0
// Usage: node --jitless testdata/generate_internal_font_css.cjs PINNED_ASSETS
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const {spawnSync} = require('node:child_process');
const assets = process.argv[2];
if (!assets) throw new Error('expected the pinned D2 MathJax asset directory');
const hashes = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
for (const [file, hash] of Object.entries(hashes)) {
  const bytes = fs.readFileSync(path.join(assets, file));
  if (crypto.createHash('sha256').update(bytes).digest('hex') !== hash) throw new Error(`unverified ${file}`);
}
// Bound process lifetime as well as each VM's lifetime for this large corpus.
function originals(cases) {
  const results = [];
  for (let start = 0; start < cases.length; start += 24) {
    const batch = cases.slice(start, start + 24);
    const input = batch.map(c => JSON.stringify({tex: c.tex, options: {Display: c.display}})).join('\n') + '\n';
    const run = spawnSync(process.execPath,
      ['--jitless', path.join(__dirname, 'differential/oracle.mjs'), '--asset-dir', assets],
      {input, encoding: 'utf8', maxBuffer: 16 * 1024 * 1024});
    if (run.status !== 0) throw new Error(run.stderr || String(run.error));
    const output = run.stdout.trim().split('\n').map(line => JSON.parse(line));
    if (output.length !== batch.length) throw new Error('incomplete original output');
    for (const result of output) {
      if (typeof result.svg !== 'string' || result.error) throw new Error(JSON.stringify(result));
      results.push(result.svg);
    }
  }
  return results;
}
const file = path.join(__dirname, 'internal_font_css_mathjax_3_2_2.json');
const fixture = JSON.parse(fs.readFileSync(file, 'utf8'));
originals(fixture.cases).forEach((svg, i) => {
  const c = fixture.cases[i];
  const dimensions = svg.match(/width="([\d.]+)ex" height="([\d.]+)ex"/);
  if (!dimensions || svg.includes('data-mjx-error=')) throw new Error(`invalid reference ${c.name}`);
  c.svg = svg;
  c.width = Math.ceil(Number(dimensions[1]) * 8);
  c.height = Math.ceil(Number(dimensions[2]) * 8);
});
fs.writeFileSync(file, JSON.stringify(fixture, null, 2) + '\n');
// Refresh originals, preserving historical Go snapshots as unresolved evidence.
const residualFile = path.join(__dirname, 'internal_font_css_residuals_mathjax_3_2_2.json');
const residuals = JSON.parse(fs.readFileSync(residualFile, 'utf8'));
originals(residuals.cases).forEach((svg, i) => residuals.cases[i].originalSVG = svg);
fs.writeFileSync(residualFile, JSON.stringify(residuals, null, 2) + '\n');
