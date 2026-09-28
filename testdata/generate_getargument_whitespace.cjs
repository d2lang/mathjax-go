// SPDX-License-Identifier: Apache-2.0
// Usage: node --jitless testdata/generate_getargument_whitespace.cjs PINNED_ASSETS
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
    try {
      const node = context.html.convert(c.tex, {em: 16, ex: 8, display: c.display});
      return {svg: context.adaptor.innerHTML(node)};
    } catch (error) {
      return {error: String(error?.stack ?? error)};
    }
  });
}
if (process.argv[3] === '--batch') {
  fs.writeFileSync(1, JSON.stringify(render(JSON.parse(fs.readFileSync(0, 'utf8')))));
  process.exit(0);
}
for (const name of [
  'getargument_whitespace_mathjax_3_2_2.json', 'getargument_whitespace_residuals.json',
]) {
  const file = path.join(__dirname, name);
  const fixture = JSON.parse(fs.readFileSync(file, 'utf8'));
  // Bound VM retention in Node while preserving an independent context for
  // every expression. Workers receive only inputs, never Go-generated SVGs.
  const cases = [...fixture.cases, ...(fixture.inheritedControls || [])];
  for (let i = 0; i < cases.length; i += 24) {
    const batch = cases.slice(i, i + 24);
    const worker = spawnSync(process.execPath, [...process.execArgv, __filename, assets, '--batch'], {
      input: JSON.stringify(batch.map(({tex, display}) => ({tex, display}))),
      encoding: 'utf8',
      maxBuffer: 64 * 1024 * 1024,
    });
    if (worker.error || worker.status !== 0) {
      throw new Error(`oracle worker failed: ${worker.error || worker.stderr}`);
    }
    const outputs = JSON.parse(worker.stdout);
    if (outputs.length !== batch.length) throw new Error('incomplete oracle batch');
    batch.forEach((c, index) => {
      const result = outputs[index];
      if (c.original?.error) {
        // Runtime-only observations are not SVG assertions. Keep the complete
        // captured stack unchanged (its caller path can differ by checkout),
        // but verify the fresh original still raises the same error message.
        if (!result.error || result.error.split('\n')[0] !== c.original.error.split('\n')[0]) {
          throw new Error(`original runtime boundary changed: ${c.name}`);
        }
      } else if (!result.svg) {
        throw new Error(`original SVG became a runtime exception: ${c.name}`);
      } else if (c.original) {
        c.original = {svg: result.svg};
      } else {
        c.svg = result.svg;
      }
    });
  }
  fs.writeFileSync(file, JSON.stringify(fixture, null, 2) + '\n');
}
