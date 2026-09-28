// SPDX-License-Identifier: Apache-2.0
// Regenerate complete original SVGs in a fresh MathJax runtime per expression.
// Usage: node --jitless testdata/generate_fallback_correction.cjs PINNED_ASSETS
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
const inputs = [];
const residualInputs = [];
function add(name, tex, residual = false) {
  for (const display of [false, true]) {
    (residual ? residualInputs : inputs).push({
      name: name + (display ? '-display' : '-inline'), tex, display,
    });
  }
}
for (const [font, a, u] of [['mathscr', 'A', 'b'], ['mathcal', 'T', 'x'],
  ['mathit', 'f', '☃'], ['mathrm', 'A', '中']]) {
  for (const [kind, value] of [['known-unknown', a + u], ['known-unknown-known', a + u + a],
    ['unknown-known', u + a], ['two-unknown', a + u + u], ['unknown', u], ['known', a]]) {
    const base = `\\${font}{${value}}`;
    // These initial 60 inputs expose a separate, unchanged parser TeXAtom
    // grouping mismatch. Preserve their full original and Go snapshots in
    // the residual file instead of presenting them as passing references.
    const residual = font === 'mathit' && value.includes('☃');
    for (const [context, tex] of [['plain', base], ['hat', `\\hat{${base}}`],
      ['script', `${base}_i^2`], ['fraction', `\\frac{${base}}{1}`],
      ['boxed', `\\boxed{${base}}`], ['size', `\\Huge{${base}}`]]) {
      add(`${font}-${kind}-${context}`, tex, residual);
    }
  }
}
// mmlToken is public TeX syntax and keeps compound text in one TextNode.
// These exercise the renderer directly without the above parser grouping.
const textCases = [
  ['known-unknown', 'f☃'], ['unknown-known', '☃f'], ['known-unknown-known', 'f☃f'],
  ['two-unknown', 'f☃☃'], ['unknown', '☃'], ['known', 'f'],
  ['cjk-last', 'f中'], ['cjk-first', '中f'], ['cjk-middle', 'f中f'],
  ['astral-last', 'f𠀀'], ['astral-first', '𠀀f'], ['astral-middle', 'f𠀀f'],
  ['digit-first', '3☃'], ['digit-last', '☃3'],
];
for (const kind of ['mi', 'mn', 'mo', 'mtext']) {
  for (const variant of ['italic', 'script']) {
    for (const [name, text] of textCases) {
      const base = `\\mmlToken{${kind}}[mathvariant="${variant}"]{${text}}`;
      add(`token-${kind}-${variant}-${name}-plain`, base);
      add(`token-${kind}-${variant}-${name}-script`, `${base}_i^2`);
      if (kind === 'mi' && ['known-unknown', 'unknown-known', 'known', 'cjk-last'].includes(name)) {
        add(`token-${kind}-${variant}-${name}-hat`, `\\hat{${base}}`);
      }
    }
  }
}
// Explicit-font measurement is a separate branch and must remain unchanged.
for (const [name, text] of [['last', 'f☃'], ['first', '☃f'], ['middle', 'f☃f'],
  ['known', 'f'], ['cjk', 'f中'], ['astral', 'f𠀀']]) {
  add(`explicit-font-${name}`, `\\hat{\\mmlToken{mi}[fontfamily="serif",fontstyle="italic"]{${text}}}_i^2`);
}
function render(c) {
  const context = vm.createContext({console});
  context.globalThis = context;
  for (const script of scripts) script.runInContext(context);
  const node = context.html.convert(c.tex, {em: 16, ex: 8, display: c.display});
  return context.adaptor.innerHTML(node);
}
const cases = inputs.map(c => {
  const svg = render(c);
  const dimensions = svg.match(/width="([\d.]+)ex" height="([\d.]+)ex"/);
  if (!dimensions || svg.includes('data-mjx-error=')) throw new Error(`invalid reference ${c.name}`);
  return {...c, svg, width: Math.ceil(Number(dimensions[1]) * 8), height: Math.ceil(Number(dimensions[2]) * 8)};
});
const fixture = {mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee', cases};
fs.writeFileSync(path.join(__dirname, 'fallback_correction_mathjax_3_2_2.json'), JSON.stringify(fixture, null, 2) + '\n');
// Refresh only original outputs. Historical Go snapshots must never be
// regenerated from a later candidate and silently reclassified as correct.
const residualFile = path.join(__dirname, 'fallback_correction_residuals_mathjax_3_2_2.json');
const residuals = JSON.parse(fs.readFileSync(residualFile, 'utf8'));
const byName = new Map(residuals.cases.map(c => [c.name, c]));
if (byName.size !== residualInputs.length) throw new Error('unexpected residual snapshot inventory');
residuals.cases = residualInputs.map(c => {
  const previous = byName.get(c.name);
  if (!previous || previous.tex !== c.tex || previous.display !== c.display) throw new Error(`missing residual ${c.name}`);
  return {...previous, originalSVG: render(c)};
});
fs.writeFileSync(residualFile, JSON.stringify(residuals, null, 2) + '\n');
