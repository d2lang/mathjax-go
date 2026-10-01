// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Capture unmodified D2 MathJax 3.2.2 with a fresh runtime per input.
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
const crypto = require('node:crypto');
const {spawnSync} = require('node:child_process');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_broken_border.cjs PINNED_ASSETS');
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
  fs.writeFileSync(1, JSON.stringify(records));
  process.exit(0);
}

const styles = [];
for (const side of sides) {
  for (const width of ['.1em', '.5em', '1em', '2em']) {
    for (const style of ['dashed', 'dotted']) styles.push([`${side}-${width}-${style}`, `border-${side}:${width} ${style} blue`]);
  }
}
for (const width of ['0px', '2px', '.2em', '.8em']) {
  for (const style of ['dashed', 'dotted']) styles.push([`global-${width}-${style}`, `border:${width} ${style} blue`]);
}
styles.push(
  ['repeated-width', 'border:3px dashed blue; border-width:2px  4px'],
  ['asymmetric-widths', 'border-width:.7em 0px 1em 0px; border-style:dashed; border-color:blue'],
  ['asymmetric-sides', 'border-top:1em dashed blue; border-bottom:2px dashed red'],
  ['padded', 'border-bottom:.5em dashed blue; padding:.2em'],
  ['zero-width-control', 'border-bottom:0px dashed blue'],
  ['solid-control', 'border-bottom:.5em solid blue'],
);
const pending = [];
for (const [label, value] of styles) {
  for (const kind of ['mi', 'mtext']) {
    const token = `\\mmlToken{${kind}}[style='${value}']{x}`;
    for (const [position, tex] of [['direct', token], ['numerator', `\\frac{${token}}{y}`], ['subscript', `z_{${token}}`]]) {
      for (const display of [false, true]) pending.push({type: 'svg', name: `${kind}-${label}-${position}-${display ? 'display' : 'inline'}`, tex, display});
    }
  }
}
const records = [];
for (let start = 0; start < pending.length; start += 32) {
  const worker = spawnSync(process.execPath, [__filename, assets, '--worker'], {input: JSON.stringify(pending.slice(start, start + 32)), encoding: 'utf8', maxBuffer: 16 * 1024 * 1024});
  if (worker.error || worker.status !== 0) throw worker.error || new Error(worker.stderr || `oracle worker ${worker.status}`);
  records.push(...JSON.parse(worker.stdout));
}
fs.writeFileSync(path.join(__dirname, 'broken_border_mathjax_3_2_2.json'), JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee', assets: hashes,
  scope: 'Broken border dash counts, including zero for segments shorter than the authored thickness; full original SVGs from fresh frozen D2 runtimes cover sides, widths, padding, and nested scale.',
  cases: records.filter(r => r.type === 'svg').map(({type, ...record}) => record),
}, null, 2) + '\n');
