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
const whitespace = [
  ['space', ' '], ['tab', '\t'], ['line-feed', '\n'], ['vertical-tab', '\v'],
  ['form-feed', '\f'], ['carriage-return', '\r'], ['nbsp', '\u00A0'],
  ['ogham', '\u1680'], ['en-quad', '\u2000'], ['em-quad', '\u2001'],
  ['en-space', '\u2002'], ['em-space', '\u2003'], ['third-em', '\u2004'],
  ['quarter-em', '\u2005'], ['sixth-em', '\u2006'], ['figure-space', '\u2007'],
  ['punctuation-space', '\u2008'], ['thin-space', '\u2009'], ['hair-space', '\u200A'],
  ['line-separator', '\u2028'], ['paragraph-separator', '\u2029'],
  ['narrow-nbsp', '\u202F'], ['medium-space', '\u205F'], ['ideographic-space', '\u3000'],
  ['bom', '\uFEFF'], ['nel-control', '\u0085'], ['mvs-control', '\u180E'], ['zwsp-control', '\u200B'],
];
const inputs = [];
for (const [name, space] of whitespace) {
  for (const [position, before, after] of [
    ['leading', space, ''], ['trailing', '', space], ['both', space, space],
  ]) {
    const number = before + '2' + after;
    const tag = before + 'round' + after;
    inputs.push(
      [`strut-height-${name}-${position}`, String.raw`\left(\xmathstrut{${number}}x\right)`],
      [`strut-depth-${name}-${position}`, String.raw`\left(\xmathstrut[${number}]{1}x\right)`],
      [`declare-tag-${name}-${position}`, String.raw`\newtagform{${tag}}{[}{]}\usetagform{round}\begin{equation}x\tag{1}\end{equation}`],
      [`select-tag-${name}-${position}`, String.raw`\newtagform{round}{[}{]}\usetagform{${tag}}\begin{equation}x\tag{1}\end{equation}`],
      [`matching-tag-${name}-${position}`, String.raw`\newtagform{${tag}}{[}{]}\usetagform{${tag}}\begin{equation}x\tag{1}\end{equation}`],
    );
  }
  inputs.push(
    [`empty-tag-${name}`, String.raw`\newtagform{${space}}{[}{]}x`],
    [`reset-tag-${name}`, String.raw`\newtagform{round}{[}{]}\usetagform{round}\usetagform{${space}}\begin{equation}x\tag{1}\end{equation}`],
  );
}
inputs.push(
  ['strut-height-control', String.raw`\left(\xmathstrut{2}x\right)`],
  ['strut-depth-control', String.raw`\left(\xmathstrut[2]{1}x\right)`],
  ['strut-number-error-control', String.raw`\xmathstrut{bad}x`],
  ['tag-selection-control', String.raw`\newtagform{round}{[}{]}\usetagform{round}\begin{equation}x\tag{1}\end{equation}`],
  ['tag-reset-control', String.raw`\newtagform{round}{[}{]}\usetagform{round}\usetagform{}\begin{equation}x\tag{1}\end{equation}`],
  ['tag-renew-bom', String.raw`\newtagform{round}{[}{]}\renewtagform{\uFEFFround\uFEFF}{<}{>}\usetagform{round}\begin{equation}x\tag{1}\end{equation}`.replaceAll('\\uFEFF', '\uFEFF')],
);
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
fs.writeFileSync(path.join(__dirname, 'mathtools_argument_trim_mathjax_3_2_2.json.gz'), zlib.gzipSync(JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee',
  assets: hashes,
  scope: 'JavaScript trim in Mathtools number and tag-form arguments; complete unmodified SVGs from a fresh frozen D2 MathJax runtime per case.',
  cases,
}, null, 2) + '\n'));
