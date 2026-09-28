// SPDX-License-Identifier: Apache-2.0
// Capture unmodified D2 MathJax 3.2.2 output in a fresh runtime per expression.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_mhchem_partial_bonds.cjs PINNED_ASSETS');
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
const forms = [
  ['single', '~'],
  ['one-and-half', '~-'],
  ['two-and-half', '~='],
  ['two-and-half-alias', '~--'],
  ['two-and-half-centered', '-~-'],
];
const inputs = [];
for (const [name, bond] of forms) {
  const chem = String.raw`\ce{C\bond{${bond}}O}`;
  for (const [context, tex] of [
    ['plain', chem],
    ['group', '{' + chem + '}'],
    ['formula', String.raw`\ce{H3C\bond{${bond}}CH2^{+}}`],
    ['charge', String.raw`\ce{[Fe\bond{${bond}}O]^{2-}}`],
    ['subscript', 'K_{' + chem + '}'],
    ['superscript', 'x^{' + chem + '}'],
    ['nested-script', 'x^{y^{' + chem + '}}'],
    ['numerator', String.raw`\frac{${chem}}{2}`],
    ['denominator', String.raw`\frac{1}{${chem}}`],
    ['scriptstyle', String.raw`\scriptstyle ${chem}`],
    ['scriptscriptstyle', String.raw`\scriptscriptstyle ${chem}`],
    ['large', String.raw`\Large ${chem}`],
    ['tiny', String.raw`\tiny ${chem}`],
    ['huge', String.raw`\Huge ${chem}`],
    ['bold', String.raw`\mathbf{${chem}}`],
    ['roman', String.raw`\mathrm{${chem}}`],
    ['sans', String.raw`\mathsf{${chem}}`],
    ['color', String.raw`\color{red}${chem}`],
    ['textcolor', String.raw`\textcolor{blue}{${chem}}`],
    ['boxed', String.raw`\boxed{${chem}}`],
    ['matrix', String.raw`\begin{matrix}${chem}&x\\y&z\end{matrix}`],
    ['root', String.raw`\sqrt{${chem}}`],
    ['root-index', String.raw`\sqrt[${chem}]{x}`],
    ['raised', String.raw`\raise.5em{${chem}}`],
    ['lowered', String.raw`\lower.2em{${chem}}`],
    ['reversible', String.raw`\ce{C\bond{${bond}}O <-->[$k_1$][$k_{-1}$] C=O}`],
  ]) inputs.push([`${name}-${context}`, tex]);
}
inputs.push(
  ['direct', String.raw`A\tripledash B`],
  ['direct-repeat', String.raw`A\tripledash\tripledash B`],
  ['direct-subscript', String.raw`A_{\tripledash}`],
  ['direct-as-subscript', String.raw`A_\tripledash`],
  ['direct-as-superscript', String.raw`A^\tripledash`],
  ['direct-follow-script', String.raw`A\tripledash^2B`],
  ['direct-tiny', String.raw`\tiny A\tripledash B`],
  ['direct-font-restores', String.raw`A\tripledash B\kern1em C`],
  ['direct-color', String.raw`\color{red}A\tripledash B`],
  ['direct-fraction', String.raw`\frac{\tripledash}{x}`],
  ['macro-override', String.raw`\DeclareMathOperator{\tripledash}{Q}\ce{C\bond{~-}O}`],
  ['all-forms', String.raw`\ce{C\bond{~}C\bond{~-}C\bond{~=}C\bond{~--}C\bond{-~-}C}`],
  ['bond-outside-chemistry', String.raw`\bond{~}`],
  ['missing-bond-argument', String.raw`\ce{C\bond}`],
  ['missing-chemistry-argument', String.raw`\ce`],
  ['ordinary-bond-controls', String.raw`\ce{C\bond{-}C\bond{=}C\bond{#}C}`],
  ['other-bond-controls', String.raw`\ce{C\bond{...}C\bond{....}C\bond{->}C\bond{<-}C\bond{<}C\bond{>}C}`],
  ['numbered-bond-controls', String.raw`\ce{C\bond{1}C\bond{2}C\bond{3}C}`],
  ['tilde-control', String.raw`\ce{C~C}`],
  ['physical-units-control', String.raw`\pu{\tripledash}`],
  ['phantom-control', String.raw`\vphantom{-}A`],
  ['raise-control', String.raw`\raise2mu{\tiny\text{-}}`],
);
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
fs.writeFileSync(path.join(__dirname, 'mhchem_partial_bonds_mathjax_3_2_2.json'), JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee',
  assets: hashes,
  scope: 'Complete unmodified SVGs from the frozen D2 MathJax bundle; each case uses a fresh runtime.',
  cases,
}, null, 2) + '\n');
