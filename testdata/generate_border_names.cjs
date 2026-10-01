// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Capture unmodified D2 MathJax 3.2.2 with a fresh runtime per input.
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
const crypto = require('node:crypto');
const {spawnSync} = require('node:child_process');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_border_names.cjs PINNED_ASSETS');
const hashes = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
const scripts = Object.entries(hashes).map(([file, hash]) => {
  const bytes = fs.readFileSync(path.join(assets, file));
  if (crypto.createHash('sha256').update(bytes).digest('hex') !== hash) throw new Error(`unverified ${file}`);
  return new vm.Script(bytes.toString(), {filename: file});
});
function runtime() {
  const context = vm.createContext({console});
  context.globalThis = context;
  for (const script of scripts) script.runInContext(context);
  return context;
}
const sides = ['top', 'right', 'bottom', 'left'];
if (process.argv[3] === '--worker') {
  const records = JSON.parse(fs.readFileSync(0, 'utf8')).map(record => {
    const context = runtime();
    if (record.type === 'svg') {
      const svg = context.adaptor.innerHTML(context.html.convert(record.tex, {em: 16, ex: 8, display: record.display}));
      if (svg.includes('data-mml-node="merror"')) throw new Error(`original rejected ${record.name}`);
      return {...record, svg};
    }
    const Styles = context.MathJax._.util.Styles.Styles;
    const styles = new Styles(record.type === 'style' ? record.value : '');
    if (record.type === 'split') {
      styles.set('padding', record.value);
      return {...record, parts: sides.map(side => styles.get('padding-' + side))};
    }
    return {...record, border: sides.map(side => ['width', 'style', 'color'].map(part => styles.get('border-' + side + '-' + part)))};
  });
  process.stdout.write(JSON.stringify(records));
  process.exit(0);
}

const whites = [
  ['space', ' '], ['tab', '\t'], ['line-feed', '\n'], ['vertical-tab', '\v'],
  ['form-feed', '\f'], ['carriage-return', '\r'], ['nbsp', '\u00a0'],
  ['ogham', '\u1680'], ['en-quad', '\u2000'], ['em-quad', '\u2001'],
  ['en-space', '\u2002'], ['em-space', '\u2003'], ['third-em', '\u2004'],
  ['quarter-em', '\u2005'], ['sixth-em', '\u2006'], ['figure-space', '\u2007'],
  ['punctuation-space', '\u2008'], ['thin-space', '\u2009'], ['hair-space', '\u200a'],
  ['line-separator', '\u2028'], ['paragraph-separator', '\u2029'],
  ['narrow-nbsp', '\u202f'], ['medium-space', '\u205f'], ['ideographic-space', '\u3000'],
  ['bom', '\ufeff'], ['nel-control', '\u0085'], ['mvs-control', '\u180e'], ['zwsp-control', '\u200b'],
];
const styles = [];
for (const [name, white] of whites) {
  for (const [position, before, after] of [['leading', white, ''], ['trailing', '', white], ['both', white, white]]) {
    for (const [kind, value] of [
      ['shorthand', `${before}border${after}:4px solid red`],
      ['followed', `${before}border${after}:4px solid red; border-bottom-width:8px`],
      ['component', `${before}border-top-width${after}:4px; border-top-style:solid; border-top-color:red`],
    ]) styles.push([`${name}-${position}-${kind}`, value]);
  }
}
styles.push(
  ['uppercase', 'Border:4px solid red; border-bottom-width:8px'],
  ['embedded-nel', 'border\u0085-width:4px; border-bottom-width:8px'],
  ['retain-previous', 'border:4px solid red; \u0085border-bottom:8px solid blue; border-left-width:12px'],
);
const pending = [];
for (const [label, value] of styles) {
  for (const kind of ['mi', 'mtext']) {
    for (const display of [false, true]) {
      pending.push({type: 'svg', name: `${kind}-${label}-${display ? 'display' : 'inline'}`, tex: `\\mmlToken{${kind}}[style='${value}']{x}`, display});
    }
  }
}
for (const [name, value] of styles) pending.push({type: 'style', name, value});
const records = [];
for (let start = 0; start < pending.length; start += 32) {
  const worker = spawnSync(process.execPath, [__filename, assets, '--worker'], {input: JSON.stringify(pending.slice(start, start + 32)), encoding: 'utf8', maxBuffer: 16 * 1024 * 1024});
  if (worker.error || worker.status !== 0) throw worker.error || new Error(worker.stderr || `oracle worker ${worker.status}`);
  records.push(...JSON.parse(worker.stdout));
}
fs.writeFileSync(path.join(__dirname, 'border_names_mathjax_3_2_2.json'), JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee', assets: hashes,
  scope: 'Border property names use the source lowercase pattern and exact JavaScript whitespace, stopping at non-whitespace prefixes. Full SVGs and direct Styles assertions use fresh frozen D2 runtimes.',
  cases: records.filter(r => r.type === 'svg').map(({type, ...record}) => record),
  styleCases: records.filter(r => r.type === 'style').map(({type, ...record}) => record),
  splitCases: records.filter(r => r.type === 'split').map(({type, ...record}) => record),
}, null, 2) + '\n');
