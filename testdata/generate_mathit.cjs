// SPDX-License-Identifier: Apache-2.0
// Frozen references from unmodified D2 v0.8.1 MathJax 3.2.2.
const fs = require('node:fs'), vm = require('node:vm'), path = require('node:path'), crypto = require('node:crypto');
const base = process.argv[2], output = process.argv[3];
if (!base || !output) throw Error('usage: node generate_mathit.cjs /path/to/d2latex output-directory');
fs.mkdirSync(output, {recursive:true});
const sha = b => crypto.createHash('sha256').update(b).digest('hex');
const assetsSHA256 = {
  'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
const assets = Object.entries(assetsSHA256).map(([name, digest]) => {
  const bytes = fs.readFileSync(path.join(base, name));
  if (sha(bytes) !== digest) throw Error('Unpinned asset: ' + name);
  return [name, bytes];
});
const inputs = [
  ['lowercase', String.raw`\mathit{x}`],
  ['uppercase', String.raw`\mathit{A}`],
  ['multiple', String.raw`\mathit{xyz}`],
  ['digits', String.raw`\mathit{123}`],
  ['greek', String.raw`\mathit{\alpha+\Gamma}`],
  ['scripts', String.raw`\mathit{x_i^2}`],
  ['fraction', String.raw`\mathit{\frac{x}{y}}`],
  ['radical', String.raw`\mathit{\sqrt{x}}`],
  ['operator', String.raw`\mathit{\sum x}`],
  ['expression', String.raw`\mathit{a+b}`],
  ['empty', String.raw`\mathit{}`],
  ['unbraced', String.raw`\mathit x`],
  ['scope', String.raw`\mathit{x}+y`],
  ['plain-control', 'x+y'],
  ['bold-control', String.raw`\mathbf{x}`],
  ['roman-control', String.raw`\mathrm{x}`],
  ['declaration-control', String.raw`{\it x}`],
  ['sin-control', String.raw`\sin x`],
  ['sans-control', String.raw`\mathsf{x}`],
];
const cases = [];
for (const [name, tex] of inputs) for (const display of [true, false]) {
  const context = vm.createContext({console:{log(){},warn(){},error(){}}});
  for (const [file, bytes] of assets) vm.runInContext(bytes.toString(), context, {filename:file});
  context.request = {tex, display};
  const svg = vm.runInContext('adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}))', context);
  if (svg.includes('data-mml-node="merror"')) throw Error('Unexpected oracle error: ' + name);
  const dimensions = svg.match(/<svg[^>]+width="([0-9.]+)ex" height="([0-9.]+)ex"/);
  if (!dimensions) throw Error('Missing serialized dimensions: ' + name);
  const label = name + (display ? '-display' : '-inline');
  fs.writeFileSync(path.join(output, label + '.svg'), svg);
  cases.push({name:label,tex,display,sha256:sha(svg),width:Math.ceil(Number(dimensions[1])*8),height:Math.ceil(Number(dimensions[2])*8)});
}
fs.writeFileSync(path.join(output, 'mathit_mathjax_3_2_2.json'), JSON.stringify({mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256,cases},null,2)+'\n');
console.log(cases.length + ' exact SVG references');
