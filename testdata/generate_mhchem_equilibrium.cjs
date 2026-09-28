// SPDX-License-Identifier: Apache-2.0
// Capture unmodified D2 MathJax 3.2.2 output in a fresh runtime per expression.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_mhchem_equilibrium.cjs PINNED_ASSETS');
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
  ['equilibrium', String.raw`\ce{N2 + 3H2 <=> 2NH3}`],
  ['right-equilibrium', String.raw`\ce{HA <=>> H+ + A-}`],
  ['left-equilibrium', String.raw`\ce{HA <<=> H+ + A-}`],
  ['equilibrium-above', String.raw`\ce{A <=>[H2O] B}`],
  ['right-equilibrium-above', String.raw`\ce{A <=>>[H2O] B}`],
  ['left-equilibrium-above', String.raw`\ce{A <<=>[H2O] B}`],
  ['equilibrium-math-label', String.raw`\ce{A <=>[$k_1$] B}`],
  ['equilibrium-text-label', String.raw`\ce{A <=>[{catalyst}] B}`],
  ['multiple-equilibria', String.raw`\ce{A <=> B <=>> C <<=> D}`],
  ['equilibrium-script', String.raw`K_{\ce{A <=> B}}`],
  ['equilibrium-color', String.raw`\color{blue}\ce{A <=> B}`],
  ['equilibrium-font', String.raw`\mathbf{\ce{A <=> B}}`],
  ['equilibrium-inline-style', String.raw`\textstyle\ce{A <=> B}`],
  ['direct-equilibrium', String.raw`A\longrightleftharpoons B`],
  ['direct-right-equilibrium', String.raw`A\longRightleftharpoons B`],
  ['direct-left-equilibrium', String.raw`A\longLeftrightharpoons B`],
  ['direct-annotated', String.raw`A\overset{k}{\longrightleftharpoons}B`],
  ['direct-superscript', String.raw`\longrightleftharpoons^n`],
  ['direct-subscript', String.raw`\longRightleftharpoons_n`],
  ['direct-combined-scripts', String.raw`\longLeftrightharpoons_n^m`],
  ['override-equilibrium', String.raw`\DeclareMathOperator{\longrightleftharpoons}{eq}\ce{A <=> B}`],
  ['override-right-equilibrium', String.raw`\DeclareMathOperator{\longRightleftharpoons}{eq}\ce{A <=>> B}`],
  ['override-left-equilibrium', String.raw`\DeclareMathOperator{\longLeftrightharpoons}{eq}\ce{A <<=> B}`],
  ['ordinary-reaction-control', String.raw`\ce{2H2 + O2 -> 2H2O}`],
  ['labelled-reaction-control', String.raw`\ce{A ->[H2O] B}`],
  ['backward-reaction-control', String.raw`\ce{A <- B}`],
  ['two-headed-reaction-control', String.raw`\ce{A <-> B}`],
  ['unit-control', String.raw`\pu{1.2e3 kJ}`],
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
fs.writeFileSync(path.join(__dirname, 'mhchem_equilibrium_mathjax_3_2_2.json'), JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee',
  assets: hashes,
  scope: 'Complete unmodified SVGs from the frozen D2 MathJax bundle; each case uses a fresh runtime.',
  cases,
}, null, 2) + '\n');
