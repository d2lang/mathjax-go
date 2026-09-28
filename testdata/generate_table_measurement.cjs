// SPDX-License-Identifier: Apache-2.0
// Usage: node --jitless testdata/generate_table_measurement.cjs PINNED_ASSETS
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const vm = require('node:vm');
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
const tableFilter = new vm.Script(`
  html.inputJax[0].postFilters.add(({data}) => {
    data.root.walkTree(node => {
      if (node.kind === 'mtable') {
        for (const [name, value] of Object.entries(tableAttributes)) {
          node.attributes.set(name, value);
        }
      }
    });
  });
`);
const file = path.join(__dirname, 'table_measurement_mathjax_3_2_2.json');
const fixture = JSON.parse(fs.readFileSync(file, 'utf8'));
for (const c of fixture.cases) {
  const context = vm.createContext({console});
  context.globalThis = context;
  for (const script of scripts) script.runInContext(context);
  // The public cases use D2's unmodified conversion. The layout controls
  // set only recorded MathML table attributes after TeX compilation and
  // before SVG wrappers are created. No output-jax code is replaced.
  if (c.tableAttributes) {
    context.tableAttributes = c.tableAttributes;
    tableFilter.runInContext(context);
  }
  const node = context.html.convert(c.tex, {em: 16, ex: 8, display: c.display});
  c.svg = context.adaptor.innerHTML(node);
}
fs.writeFileSync(file, JSON.stringify(fixture, null, 2) + '\n');
