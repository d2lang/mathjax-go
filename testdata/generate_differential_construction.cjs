// SPDX-License-Identifier: Apache-2.0
// Capture unmodified D2 MathJax 3.2.2 output in a fresh runtime per expression.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_differential_construction.cjs PINNED_ASSETS');
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
  ['dd', String.raw`\dd{x}`],
  ['differential', String.raw`\differential{x}`],
  ['var', String.raw`\var{x}`],
  ['variation', String.raw`\variation{x}`],
  ['leading-superscript', String.raw`\dd{^2x}`],
  ['leading-subscript', String.raw`\dd{_x}`],
  ['leading-prime', String.raw`\dd{'}x`],
  ['paired-scripts', String.raw`\dd[2]{_x}f(x)`],
  ['variation-scripts', String.raw`\var[2]{_x}S`],
  ['empty-order', String.raw`\dd[]{x}`],
  ['order', String.raw`\dd[2]{x}`],
  ['variation-order', String.raw`\var[2]{x}`],
  ['fraction-order', String.raw`\dd[\frac{a}{b}]{x}`],
  ['declared-differential', String.raw`\DeclareMathOperator{\diffd}{D}\dd{x}`],
  ['declared-variation', String.raw`\DeclareMathOperator{\delta}{D}\var{x}`],
  ['declared-differential-order', String.raw`\DeclareMathOperator{\diffd}{D}\differential[2]{x}`],
  ['declared-variation-parens', String.raw`\DeclareMathOperator{\delta}{D}\variation(x)`],
  ['bold-scope', String.raw`{\bf\dd[2]{x}}+y`],
  ['operand-font', String.raw`\dd{\bf x}+y`],
  ['shared-declaration', String.raw`\dd{\DeclareMathOperator{\foo}{F}x}\foo y`],
  ['empty-braces', String.raw`a\dd{}b`],
  ['empty-tail', String.raw`a\dd`],
  ['unbraced', String.raw`\dd x+y`],
  ['unbraced-command', String.raw`\dd\alpha`],
  ['unary-operand', String.raw`\dd{-x}`],
  ['function-operand', String.raw`\dd{\sin x}`],
  ['fraction-operand', String.raw`\dd{\frac{x}{y}}`],
  ['parentheses', String.raw`\dd[2](g)+Z`],
  ['nested-parentheses', String.raw`\dd((x))`],
  ['left-right-parentheses', String.raw`\dd(\left(x\right))`],
  ['parentheses-font', String.raw`\dd(\bf x)+y`],
  ['script-recipient', String.raw`x^\dd[2](y)+Z`],
  ['duplicate-superscript', String.raw`\dd[2]{^3x}`],
  ['missing-bracket', String.raw`\dd[2`],
  ['missing-parenthesis', String.raw`\dd(x`],
  ['unbraced-missing-argument', String.raw`\dd\frac{x}{y}`],
];
const cases = [];
for (const [name, tex] of inputs) {
  for (const display of [false, true]) {
    const context = vm.createContext({console});
    context.globalThis = context;
    for (const script of scripts) script.runInContext(context);
    const svg = context.adaptor.innerHTML(context.html.convert(tex, {em: 16, ex: 8, display}));
    const width = Math.ceil(Number(svg.match(/ width="([\d.]+)ex"/)[1]) * 8);
    const height = Math.ceil(Number(svg.match(/ height="([\d.]+)ex"/)[1]) * 8);
    cases.push({name: `${name}-${display ? 'display' : 'inline'}`, tex, display, svg, width, height});
  }
}
fs.writeFileSync(path.join(__dirname, 'differential_construction_mathjax_3_2_2.json'), JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee',
  assets: hashes,
  scope: 'Complete unmodified SVGs from the frozen D2 MathJax bundle; each case uses a fresh runtime.',
  cases,
}, null, 2) + '\n');
