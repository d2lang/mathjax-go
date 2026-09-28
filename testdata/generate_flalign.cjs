// SPDX-License-Identifier: Apache-2.0
// Usage: node --jitless testdata/generate_flalign.cjs PINNED_ASSETS
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const vm = require('node:vm');
const {spawnSync} = require('node:child_process');
const assets = process.argv[2];
if (!assets) throw new Error('expected the pinned D2 MathJax asset directory');
const hashes = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
const scripts = Object.entries(hashes).map(([file, hash]) => {
  const bytes = fs.readFileSync(path.join(assets, file));
  if (crypto.createHash('sha256').update(bytes).digest('hex') !== hash) {
    throw new Error(`unverified ${file}`);
  }
  return new vm.Script(bytes.toString(), {filename: file});
});
function render(cases) {
  return cases.map(c => {
    const context = vm.createContext({console});
    context.globalThis = context;
    for (const script of scripts) script.runInContext(context);
    if (c.tableAttributes && c.tableAttributes.length) {
      context.tableAttributes = c.tableAttributes;
      new vm.Script(`html.inputJax[0].postFilters.add(({data}) => {
        let index = 0;
        data.root.walkTree(node => {
          if (node.kind === 'mtable') {
            for (const [key, value] of Object.entries(tableAttributes[index] || {})) node.attributes.set(key, value);
            index++;
          }
        });
      });`).runInContext(context);
    }
    const node = context.html.convert(c.tex, {em: 16, ex: 8, display: c.display});
    return context.adaptor.innerHTML(node);
  });
}
if (process.argv[3] === '--batch') {
  fs.writeFileSync(1, JSON.stringify(render(JSON.parse(fs.readFileSync(0, 'utf8')))));
  process.exit(0);
}
for (const name of ['flalign_mathjax_3_2_2.json', 'flalign_residuals.json', 'flalign_nested_mathjax_3_2_2.json', 'flalign_nested_output_observations.json']) {
  const file = path.join(__dirname, name);
  const fixture = JSON.parse(fs.readFileSync(file, 'utf8'));
  // Bound VM retention in Node while preserving an independent context for
  // every expression. Workers receive only inputs, never Go-generated SVGs.
  for (let i = 0; i < fixture.cases.length; i += 24) {
    const batch = fixture.cases.slice(i, i + 24);
    const worker = spawnSync(process.execPath, [...process.execArgv, __filename, assets, '--batch'], {
      input: JSON.stringify(batch.map(({tex, display, tableAttributes}) => ({tex, display, tableAttributes}))),
      encoding: 'utf8',
      maxBuffer: 64 * 1024 * 1024,
    });
    if (worker.error || worker.status !== 0) {
      throw new Error(`oracle worker failed: ${worker.error || worker.stderr}`);
    }
    const svgs = JSON.parse(worker.stdout);
    if (svgs.length !== batch.length) throw new Error('incomplete oracle batch');
    batch.forEach((c, index) => {
      const svg = svgs[index];
      if (c.original) {
        // Residuals retain their complete original reference, independently
        // of the recorded baseline/candidate output from the broader audit.
        c.original = {svg};
        c.validOriginal = !svg.includes('data-mjx-error');
      } else {
        c.svg = svg;
      }
    });
  }
  fs.writeFileSync(file, JSON.stringify(fixture, null, 2) + '\n');
}
