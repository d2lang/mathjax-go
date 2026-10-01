// SPDX-License-Identifier: Apache-2.0
// Capture unmodified frozen MathJax 3.2.2 output with a fresh runtime per input.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const zlib = require('node:zlib');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node --jitless generate_keyval_object_properties.cjs PINNED_ASSETS');
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
function fresh() {
  const context = vm.createContext({console});
  context.globalThis = context;
  for (const script of scripts) script.runInContext(context);
  return context;
}
const inputs = [];
for (const [name, value] of [
  ['bare', null], ['true', 'true'], ['false', 'false'], ['empty', ''],
  ['empty-braces', '{}'], ['nested', '{{x}}'], ['null-string', 'null'], ['other', 'x'],
]) {
  for (const [prefix, key] of [['plain', '__proto__'], ['braced', '{__proto__}'], ['bom', '\uFEFF__proto__\uFEFF']]) {
    const option = key + (value === null ? '' : '=' + value);
    inputs.push([`${prefix}-proto-${name}`, String.raw`\mathtoolsset{${option}}a:b`]);
    inputs.push([`${prefix}-proto-${name}-colon`, String.raw`\mathtoolsset{centercolon=true,${option}}a:b`]);
  }
}
for (const [name, raw] of [
  ['proto-first-toggle', '__proto__=false,centercolon=true'],
  ['proto-last-toggle', 'centercolon=false,__proto__=true,centercolon=true'],
  ['duplicate-proto', '__proto__=true,__proto__=false'],
  ['duplicate-other', 'centercolon=true,__proto__=true,centercolon=false'],
  ['value-proto-control', 'centercolon=__proto__'],
  ['repeated-comma', ',,__proto__=true,,centercolon=true,'],
  ['proto-before-unknown', '__proto__=true,unknown=x'],
  ['unknown-before-proto', 'unknown=x,__proto__=false'],
  ['constructor-control', 'constructor=x'], ['toString-control', 'toString=false'],
  ['hasOwnProperty-control', 'hasOwnProperty=true'], ['prototype-control', 'prototype=x'],
  ['__defineGetter__-control', '__defineGetter__=x'], ['__PROTO__-control', '__PROTO__=x'],
  ['__proto__x-control', '__proto__x=x'], ['proto-unmatched-close', '__proto__=x}'],
]) inputs.push([name, String.raw`\mathtoolsset{${raw}}a:b`]);
for (const key of ['0', '2', '10', '4294967294', '00', '01', '-0', '+2', '2.0', '2e0', '4294967295', '4294967296']) {
  inputs.push([`numeric-${key}`, String.raw`\mathtoolsset{unknown=x,${key}=y}a:b`]);
  inputs.push([`numeric-first-${key}`, String.raw`\mathtoolsset{${key}=y,unknown=x}a:b`]);
}
for (const [name, raw] of [
  ['numeric-sort', 'unknown=x,10=y,2=z,0=w'],
  ['numeric-boundary', 'unknown=x,4294967295=y,4294967294=z'],
  ['numeric-duplicate', 'unknown=x,10=y,2=z,10=w'],
  ['numeric-braced', 'unknown=x,{{2}}=z'],
  ['numeric-space', 'unknown=x, 2 = z'],
  ['proto-and-numeric', 'unknown=x,__proto__=true,2=y'],
  ['empty-key-before-numeric', '=x,2=y'],
  ['empty-key-after-numeric', '2=y,=x'],
]) inputs.push([name, String.raw`\mathtoolsset{${raw}}a:b`]);
inputs.push(['d2-witness', String.raw`\mathtoolsset{__proto__=true,centercolon=true}\text{Rate: }x:y = \frac{a}{b}`]);
const cases = [];
for (const [name, tex] of inputs) {
  for (const display of [false, true]) {
    const context = fresh();
    const svg = context.adaptor.innerHTML(context.html.convert(tex, {em: 16, ex: 8, display}));
    cases.push({name: `${name}-${display ? 'display' : 'inline'}`, tex, display, svg});
  }
}
const base = {mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee', assets: hashes};
fs.writeFileSync(path.join(__dirname, 'keyval_object_properties_mathjax_3_2_2.json.gz'), zlib.gzipSync(JSON.stringify({
  ...base, scope: 'Complete unmodified SVGs from the frozen D2 bundle; fresh runtime per input, including all rendered errors.', cases,
}, null, 2) + '\n'));

// Observe the exported original helper both without validation and with the
// Empheq left/right allowlist, without enabling an additional package.
const rawInputs = [
  ['empty', ''], ['left-right', 'left={(},right={)}'], ['booleans', 'left=false,right'],
  ['nested', 'left={{(}},right={{)}}'], ['duplicates', 'right=x,left=y,right=z'],
  ['bare-proto', '__proto__'], ['true-proto', '__proto__=true'], ['false-proto', '__proto__=false'],
  ['string-proto', '__proto__=x'], ['empty-proto', '__proto__='], ['braces-proto', '__proto__={}'],
  ['braced-proto-key', '{{__proto__}}=x'], ['proto-before-left', '__proto__=true,left={(}'],
  ['proto-after-right', 'right={)},__proto__=false'], ['proto-duplicate', '__proto__=x,__proto__=false,left={(}'],
  ['constructor', 'constructor=x'], ['toString', 'toString=x'], ['hasOwnProperty', 'hasOwnProperty=x'],
  ['__defineGetter__', '__defineGetter__=x'], ['prototype', 'prototype=x'],
  ['proto-malformed-open', '__proto__={'], ['unknown-malformed-open', 'unknown=x,__proto__={'],
  ['proto-unknown', '__proto__=x,unknown=y'], ['unknown-proto', 'unknown=x,__proto__=y'],
  ['numeric-sort', 'unknown=x,10=y,2=z,0=w'], ['numeric-duplicate', 'unknown=x,10=y,2=z,10=w'],
  ['numeric-boundary', 'unknown=x,4294967295=y,4294967294=z'],
  ['numeric-braced', 'unknown=x,{{2}}=z'], ['numeric-space', 'unknown=x, 2 = z'],
  ['proto-numeric', '__proto__=x,unknown=y,2=z'], ['left-numeric', 'left={(},unknown=x,2=z'],
  ['empty-key-numeric', '=x,2=y'], ['numeric-empty-key', '2=y,=x'],
];
for (const key of ['0', '2', '10', '4294967294', '00', '01', '-0', '+2', '2.0', '2e0', '4294967295', '4294967296']) {
  rawInputs.push([`numeric-${key}`, `unknown=x,${key}=y`]);
}
const observations = [];
let keyvalMethod;
for (const [name, raw] of rawInputs) {
  for (const validated of [false, true]) {
    const context = fresh();
    const method = context.MathJax._.input.tex.ParseUtil.default.keyvalOptions;
    keyvalMethod = method.toString();
    let options = null, keys = null, error = null;
    try {
      options = method(raw, validated ? {left: 1, right: 1} : null, true);
      keys = Object.keys(options);
    } catch (e) { error = {id: e.id, message: e.message}; }
    observations.push({name: `${name}-${validated ? 'empheq' : 'unvalidated'}`, raw, validated, options, keys, error});
  }
}
fs.writeFileSync(path.join(__dirname, '../internal/tex/testdata/option_keyval_object_mathjax_3_2_2.json'), JSON.stringify({
  ...base, scope: 'Direct original ParseUtil.keyvalOptions with and without the Empheq allowlist; fresh frozen runtime per call.',
  keyvalMethod, cases: observations,
}, null, 2) + '\n');
console.log(JSON.stringify({svgCases: cases.length, validSVGs: cases.filter(c => !c.svg.includes('data-mjx-error')).length,
  errorSVGs: cases.filter(c => c.svg.includes('data-mjx-error')).length, keyvalCases: observations.length,
  successfulKeyvals: observations.filter(c => c.error === null).length, keyvalErrors: observations.filter(c => c.error !== null).length}));
