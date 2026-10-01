// SPDX-License-Identifier: Apache-2.0
// Capture unmodified D2 MathJax 3.2.2 output in a fresh runtime per expression.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const zlib = require('node:zlib');
const {spawnSync} = require('node:child_process');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_physics_number_whitespace.cjs PINNED_ASSETS');
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
if (process.argv[3] === '--worker') {
  const cases = JSON.parse(fs.readFileSync(0, 'utf8')).map(input => {
    const context = vm.createContext({console});
    context.globalThis = context;
    for (const script of scripts) script.runInContext(context);
    const svg = context.adaptor.innerHTML(context.html.convert(input.tex, {em: 16, ex: 8, display: input.display}));
    return {...input, svg};
  });
  fs.writeFileSync(1, JSON.stringify(cases));
  process.exit(0);
}
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
    for (const alias of ['imat', 'identitymatrix']) {
      for (const [numberName, number] of [['two', '2']]) {
        const command = `\\${alias}{${before}${number}${after}}`;
        for (const [contextName, context] of [
          ['pmatrix', x => String.raw`\begin{pmatrix}${x}\end{pmatrix}`],
          ['quantity', x => String.raw`\mqty(${x})`],
        ]) inputs.push([`${name}-${position}-${alias}-${numberName}-${contextName}`, context(command)]);
      }
    }
    inputs.push(
      [`xmat-strict-${name}-${position}`, String.raw`\begin{pmatrix}\xmat{x}{${before}2${after}}{1}\end{pmatrix}`],
      [`zmat-strict-${name}-${position}`, String.raw`\begin{pmatrix}\zmat{1}{${before}2${after}}\end{pmatrix}`],
    );
  }
}
inputs.push(
  ['identity-control', String.raw`\begin{pmatrix}\imat{2}\end{pmatrix}`],
  ['identity-zero-control', String.raw`\begin{pmatrix}\imat{0}\end{pmatrix}`],
  ['identity-prefix-control', String.raw`\begin{pmatrix}\imat{2x}\end{pmatrix}`],
  ['identity-no-digits-control', String.raw`\begin{pmatrix}\imat{bad}\end{pmatrix}`],
  ['xmatrix-control', String.raw`\begin{pmatrix}\xmat{x}{2}{2}\end{pmatrix}`],
  ['xmatrix-star-control', String.raw`\begin{pmatrix}\xmat*{x}{2}{2}\end{pmatrix}`],
  ['xmatrix-negative-control', String.raw`\begin{pmatrix}\xmat{x}{-2}{1}\end{pmatrix}`],
  ['xmatrix-leading-zero-control', String.raw`\begin{pmatrix}\xmat{x}{02}{1}\end{pmatrix}`],
);
const numberInputs = [
  '0', '1', '2', '3', '-1', '-2', '+2', '02', '2x', '2.9', '2e3',
  '0x10', '-0x10', '0b10', '+0x10', '-9007199254740992', '-9007199254740993',
  '-9007199254740994', '-1000000000000000128', '-1000000000000000100',
  '-9223372036854775808', '-9223372036854776000', '-18446744073709552000',
  '-100000000000000000000', '-1000000000000000000000',
  '-' + '9'.repeat(500), '-' + '9'.repeat(500) + 'suffix', '+', '--2', 'bad', '  -3 ',
];
for (const [index, number] of numberInputs.entries()) {
  for (const alias of ['imat', 'identitymatrix']) {
    inputs.push([`number-${index}-${alias}`, String.raw`\begin{pmatrix}${'\\' + alias}{${number}}\end{pmatrix}`]);
  }
  for (const alias of ['xmat', 'xmatrix', 'zmat', 'zeromatrix']) {
    for (const position of ['rows', 'columns']) {
      const rows = position === 'rows' ? number : '1';
      const columns = position === 'columns' ? number : '1';
      const entry = alias[0] === 'x' ? '{x}' : '';
      inputs.push([`number-${index}-${alias}-${position}`,
        String.raw`\begin{pmatrix}${'\\' + alias}${entry}{${rows}}{${columns}}\end{pmatrix}`]);
    }
  }
}
const unique = new Map();
for (const [name, tex] of inputs) {
  for (const display of [false, true]) {
    const key = JSON.stringify([tex, display]);
    if (!unique.has(key)) unique.set(key, {name: `${name}-${display ? 'display' : 'inline'}`, tex, display});
  }
}
const pending = [...unique.values()];
const cases = [];
// Bound compiled runtime retention while retaining a fresh VM for every case.
for (let start = 0; start < pending.length; start += 256) {
  const result = spawnSync(process.execPath, [__filename, assets, '--worker'], {
    input: JSON.stringify(pending.slice(start, start + 256)), encoding: 'utf8', maxBuffer: 16 * 1024 * 1024,
  });
  if (result.error || result.status !== 0) throw result.error || new Error(result.stderr || `oracle worker ${result.status}`);
  cases.push(...JSON.parse(result.stdout));
}
fs.writeFileSync(path.join(__dirname, 'physics_number_whitespace_mathjax_3_2_2.json.gz'), zlib.gzipSync(JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee',
  assets: hashes,
  scope: 'JavaScript parseInt leading whitespace, decimal-prefix Number rounding, and strict dimension strings in Physics matrix sizes; complete unmodified SVGs from a fresh frozen D2 MathJax runtime per case. Large positive allocations and original runtime failures are excluded.',
  cases,
}, null, 2) + '\n'));
