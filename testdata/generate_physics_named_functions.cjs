// SPDX-License-Identifier: Apache-2.0
// Regenerate complete original SVGs; the fixture owns the input inventory.
// Usage: node --jitless testdata/generate_physics_named_functions.cjs PINNED_ASSETS
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
  if (crypto.createHash('sha256').update(bytes).digest('hex') !== hash) {
    throw new Error(`unverified ${file}`);
  }
}
for (const name of ['physics_named_functions_mathjax_3_2_2.json', 'physics_named_functions_residuals.json']) {
  const file = path.join(__dirname, name);
  const fixture = JSON.parse(fs.readFileSync(file, 'utf8'));
  const inventory = [...fixture.cases, ...(fixture.inheritedControls || [])];
  for (let start = 0; start < inventory.length; start += 24) {
    const batch = inventory.slice(start, start + 24);
    const input = batch.map(c => JSON.stringify({tex: c.tex, options: {Display: c.display}})).join('\n') + '\n';
    const run = spawnSync(process.execPath,
      ['--jitless', path.join(__dirname, 'differential/oracle.mjs'), '--asset-dir', assets],
      {input, encoding: 'utf8', maxBuffer: 16 * 1024 * 1024});
    if (run.status !== 0) throw new Error(run.stderr || String(run.error));
    const results = run.stdout.trim().split('\n').map(line => JSON.parse(line));
    if (results.length !== batch.length) throw new Error('incomplete original output');
    results.forEach((result, i) => {
      if (typeof result.svg !== 'string' || result.error) throw new Error(JSON.stringify(result));
      // Never replace original SVGs with Go output; residuals keep their
      // separate baseline/candidate observations when originals regenerate.
      if (batch[i].original) batch[i].original = {svg: result.svg};
      else batch[i].svg = result.svg;
    });
  }
  fs.writeFileSync(file, JSON.stringify(fixture, null, 2) + '\n');
}
