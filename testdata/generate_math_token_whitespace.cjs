// SPDX-License-Identifier: Apache-2.0
// Usage: node --jitless testdata/generate_math_token_whitespace.cjs PINNED_ASSETS
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
for (const [name, expected] of Object.entries(hashes)) {
  const actual = crypto.createHash('sha256').update(fs.readFileSync(path.join(assets, name))).digest('hex');
  if (actual !== expected) throw Error(`unverified ${name}`);
}
for (const name of ['math_token_whitespace_mathjax_3_2_2.json', 'math_token_whitespace_residuals.json', 'math_token_historical_boundaries.json']) {
  const file = path.join(__dirname, name);
  const fixture = JSON.parse(fs.readFileSync(file, 'utf8'));
  const strict = name === 'math_token_whitespace_mathjax_3_2_2.json';
  for (let start = 0; start < fixture.cases.length; start += 24) {
    const batch = fixture.cases.slice(start, start + 24);
    const input = batch.map(c => JSON.stringify({tex:c.tex, options:{Display:c.display}})
      .replace(/[\u0080-\uffff]/g, c => '\\u' + c.charCodeAt(0).toString(16).padStart(4,'0'))).join('\n') + '\n';
    const run = spawnSync(process.execPath, ['--jitless', path.join(__dirname, 'differential/oracle.mjs'), '--asset-dir', assets],
      {input, encoding:'utf8', maxBuffer:16*1024*1024});
    if (run.status !== 0) throw Error(run.stderr || String(run.error));
    const results = run.stdout.replace(/\n$/, '').split('\n').map(line => JSON.parse(line));
    if (results.length !== batch.length) throw Error('incomplete original output');
    results.forEach((result, i) => {
      const c = batch[i];
      if (strict) {
        if (typeof result.svg !== 'string' || result.error) throw Error(JSON.stringify(result));
        c.svg = result.svg;
      } else if (typeof c.original.error === 'string') {
        // Preserve the entire captured stack; compare the exception's source
        // identity without rewriting paths from the original capture process.
        if (typeof result.error !== 'string' || result.error.split('\n')[0] !== c.original.error.split('\n')[0])
          throw Error('original runtime failure changed');
      } else c.original = result;
    });
  }
  fs.writeFileSync(file, JSON.stringify(fixture, null, 2) + '\n');
}
