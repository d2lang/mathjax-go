// SPDX-License-Identifier: Apache-2.0
// node --jitless testdata/generate_array_horizontal_rules.cjs PINNED_ASSETS
// Rebuild only original observations. No candidate invocation or classification.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const {spawnSync} = require('node:child_process');
const assets = process.argv[2];
if (!assets) throw Error('expected pinned D2 asset directory');
const hashes = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
for (const [name, hash] of Object.entries(hashes)) {
  const actual = crypto.createHash('sha256').update(fs.readFileSync(path.join(assets, name))).digest('hex');
  if (actual !== hash) throw Error(`unverified original asset ${name}`);
}
for (const name of ['array_horizontal_rules_mathjax_3_2_2.json', 'array_horizontal_rules_residuals.json']) {
  const file = path.join(__dirname, name);
  const fixture = JSON.parse(fs.readFileSync(file, 'utf8'));
  if (fixture.mathjaxGitCommit !== 'ad8f5c21cb810236551da8c6512ba733e67357ee') throw Error('unbound original revision');
  for (let start = 0; start < fixture.cases.length; start += 24) {
    const batch = fixture.cases.slice(start, start + 24);
    const input = batch.map(c => JSON.stringify({tex: c.tex, options: {Display: c.display}})
      .replace(/[\u0080-\uffff]/g, c => '\\u' + c.charCodeAt(0).toString(16).padStart(4, '0'))).join('\n') + '\n';
    const run = spawnSync(process.execPath, ['--jitless', path.join(__dirname, 'differential/oracle.mjs'), '--asset-dir', assets],
      {input, encoding: 'utf8', maxBuffer: 32 * 1024 * 1024});
    if (run.status !== 0) throw Error(run.stderr || String(run.error));
    const results = run.stdout.replace(/\n$/, '').split('\n').map(s => JSON.parse(s));
    if (results.length !== batch.length) throw Error('incomplete original response');
    results.forEach((result, i) => {
      if (name.includes('residuals')) batch[i].original = result;
      else {
        if (typeof result.svg !== 'string' || result.error) throw Error(JSON.stringify(result));
        batch[i].svg = result.svg;
      }
    });
  }
  fs.writeFileSync(file, JSON.stringify(fixture, null, 2) + '\n');
}
