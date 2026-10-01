// SPDX-License-Identifier: Apache-2.0
// Capture unmodified D2 MathJax 3.2.2 output in a fresh runtime per expression.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const zlib = require('node:zlib');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_mathtools_options.cjs PINNED_ASSETS');
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
const inputs = [
  ['text-colon', String.raw`\mathtoolsset{centercolon=true}\text{Time: 2}`],
  ['text-and-math-colons', String.raw`\mathtoolsset{centercolon=true}\text{Time: }a:b`],
  ['toggle-off', String.raw`\mathtoolsset{centercolon=true}a:b\mathtoolsset{centercolon=false}c:d`],
  ['toggle-on', String.raw`a:b\mathtoolsset{centercolon=true}c:d`],
  ['toggle-offset', String.raw`\mathtoolsset{centercolon=true}a:b\mathtoolsset{centercolon-offset=.3em}c:d`],
  ['option-name-colon', String.raw`\mathtoolsset{centercolon=true}\newtagform{a:b}{[}{]}\usetagform{a:b}\begin{equation}a:b\tag{2}\end{equation}`],
  ['macro-before', String.raw`\newcommand{\foo}{:}\mathtoolsset{centercolon=true}\foo`],
  ['macro-after', String.raw`\mathtoolsset{centercolon=true}\newcommand{\foo}{:}\mathtoolsset{centercolon=false}\foo`],
  ['text-macro', String.raw`\mathtoolsset{centercolon=true}\newcommand{\foo}{\text{Time: }}\foo a:b`],
  ['array', String.raw`\mathtoolsset{centercolon=true}\begin{array}{cc}a:b&\text{A: B}\end{array}`],
  ['alignment-toggle', String.raw`\begin{aligned}\mathtoolsset{centercolon=true}a:b&=c\\\mathtoolsset{centercolon=false}d:e&=f\end{aligned}`],
  ['script', String.raw`\mathtoolsset{centercolon=true}x^{a:b}`],
  ['fraction', String.raw`\mathtoolsset{centercolon=true}\frac{a:b}{c:d}`],
  ['color', String.raw`\color{red}\mathtoolsset{centercolon=true}a:b`],
  ['font', String.raw`\mathbf{\mathtoolsset{centercolon=true}a:b}`],
  ['braket', String.raw`\mathtoolsset{centercolon=true}\ket{a:b}`],
  ['forced-colon-control', String.raw`\mathtoolsset{centercolon=false}a\centercolon b\ordinarycolon c\MTThinColon d`],
  ['plain-colon-control', 'a:b'],
  ['centered-text-math', String.raw`\mathtoolsset{centercolon=true}\text{Time: $a:b$}`],
];
for (const [name, value] of [
  ['true', 'true'], ['false', 'false'], ['upper-true', 'TRUE'], ['upper-false', 'FALSE'],
  ['zero-string', '0'], ['one-string', '1'], ['empty-value', ''], ['empty-braces', '{}'],
  ['blank-braces', '{ }'], ['space-braces', ' {true} '], ['nested-true', '{{true}}'],
  ['nested-false', '{{ false }}'], ['other-string', '{other}'],
]) {
  inputs.push([`centercolon-${name}`, String.raw`\mathtoolsset{centercolon=${value}}a:b`]);
  inputs.push([`use-unicode-${name}`, String.raw`\mathtoolsset{use-unicode=${value}}a\coloneqq b`]);
}
for (const [name, options] of [
  ['empty', ''], ['space', ' '], ['comma', ','], ['empty-entries', ',,'],
  ['bare-key', 'centercolon'],
  ['trailing-comma', 'centercolon=true,'], ['repeated-commas', 'centercolon=true,,use-unicode=true'],
  ['duplicates', 'centercolon=false,centercolon=true'],
  ['braced-offset', 'centercolon-offset={{.3em}},centercolon=true'],
  ['unknown-key', 'centercolon=true,unknown=x'], ['forbidden-key', 'allow-mathtoolsset=false'],
  ['empty-key', 'centercolon=true,=true'],
]) inputs.push([`keyval-${name}`, String.raw`\mathtoolsset{${options}}a:b`]);
for (const [name, whitespace] of [['bom', '\uFEFF'], ['nel', '\u0085']]) {
  for (const [suffix, options] of [
    ['key', `${whitespace}centercolon${whitespace}=true`],
    ['true', `centercolon=${whitespace}true${whitespace}`],
    ['false', `centercolon=${whitespace}false${whitespace}`],
    ['offset', `centercolon-offset=${whitespace}.3em${whitespace},centercolon=true`],
    ['unicode-false', `use-unicode=${whitespace}false${whitespace}`],
  ]) inputs.push([`keyval-${name}-${suffix}`, String.raw`\mathtoolsset{${options}}a:b\coloneqq c`]);
}
const cases = [];
for (const [name, tex] of inputs) {
  for (const display of [false, true]) {
    const context = vm.createContext({console});
    context.globalThis = context;
    for (const script of scripts) script.runInContext(context);
    const svg = context.adaptor.innerHTML(context.html.convert(tex, {em: 16, ex: 8, display}));
    cases.push({name: `${name}-${display ? 'display' : 'inline'}`, tex, display, svg});
  }
}
fs.writeFileSync(path.join(__dirname, 'mathtools_options_mathjax_3_2_2.json.gz'), zlib.gzipSync(JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee',
  assets: hashes,
  scope: 'Complete unmodified SVGs from the frozen D2 MathJax bundle; each case uses a fresh runtime.',
  cases,
}, null, 2) + '\n'));

// Empheq uses the same ParseUtil.keyvalOptions helper with left/right allowed.
// Observe that original helper directly without enabling an extra package.
const keyvals = [
  ['plain', 'left={(},right={)}'],
  ['empty', ''],
  ['booleans', 'left=false,right'],
  ['nested-braces', 'left={{(}},right={{)}}'],
];
for (const [name, whitespace] of [['bom', '\uFEFF'], ['nel', '\u0085']]) {
  for (const [suffix, raw] of [
    ['key', `${whitespace}left${whitespace}={(}`],
    ['value', `left=${whitespace}(${whitespace},right=${whitespace})${whitespace}`],
    ['braced-value', `left={${whitespace}(${whitespace}}`],
    ['true', `left=${whitespace}true${whitespace}`],
    ['false', `left=${whitespace}false${whitespace}`],
    ['brace-edges', `left=${whitespace}{(}${whitespace}`],
  ]) keyvals.push([`${name}-${suffix}`, raw]);
}
const keyvalCases = [];
let keyvalMethod;
for (const [name, raw] of keyvals) {
  const context = vm.createContext({console});
  context.globalThis = context;
  for (const script of scripts) script.runInContext(context);
  const method = context.MathJax._.input.tex.ParseUtil.default.keyvalOptions;
  keyvalMethod = method.toString();
  let options = null, error = null;
  try { options = method(raw, {left: 1, right: 1}, true); }
  catch (e) { error = {id: e.id, message: e.message}; }
  keyvalCases.push({name, raw, options, error});
}
fs.writeFileSync(path.join(__dirname, '../internal/tex/testdata/option_keyval_trim_mathjax_3_2_2.json'), JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee', assets: hashes,
  scope: 'Direct original ParseUtil.keyvalOptions with the Empheq left/right allowlist; fresh frozen runtime per call.',
  keyvalMethod, cases: keyvalCases,
}, null, 2) + '\n');
