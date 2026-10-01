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
const scriptsCases = [
  ['both', ['a', 'b', 'C']],
  ['sup-only', ['a', '', 'C']],
  ['sub-only', ['', 'b', 'C']],
  ['neither', ['', '', 'C']],
  ['blank', [' ', ' ', 'C']],
  ['control-space', ['\\ ', '', 'C']],
  ['phantom', [String.raw`\vphantom{\frac12}`, 'b', 'C']],
  ['empty-group', ['{}', 'b', 'C']],
  ['empty-base', ['a', 'b', '']],
];
const formats = [
  ['default', ''],
  ['sup-bold', String.raw`\mathtoolsset{prescript-sup-format=\mathbf}`],
  ['sub-sans', String.raw`\mathtoolsset{prescript-sub-format=\mathsf}`],
  ['base-roman', String.raw`\mathtoolsset{prescript-arg-format=\mathrm}`],
  ['colors', String.raw`\mathtoolsset{prescript-sub-format=\color{red},prescript-sup-format=\color{blue},prescript-arg-format=\color{green}}`],
];
const contexts = [
  ['plain', x => x],
  ['group', x => '{' + x + '}'],
  ['script', x => 'x^{' + x + '}'],
  ['fraction', x => String.raw`\frac{${x}}{z}`],
  ['font', x => String.raw`\mathsf{${x}}`],
  ['vector', x => String.raw`\va{${x}}`],
];
const inputs = [];
for (const [scriptName, args] of scriptsCases) {
  for (const [formatName, format] of formats) {
    const tex = format + String.raw`\prescript{${args[0]}}{${args[1]}}{${args[2]}}`;
    for (const [contextName, context] of contexts) inputs.push([`${scriptName}-${formatName}-${contextName}`, context(tex)]);
  }
}
inputs.push(
  ['format-disabled', String.raw`\mathtoolsset{prescript-sup-format=false}\prescript{a}{b}{C}`],
  ['format-toggle', String.raw`\mathtoolsset{prescript-sup-format=\mathbf}\prescript{a}{b}{C}\mathtoolsset{prescript-sup-format=false}\prescript{a}{b}{C}`],
  ['format-macro', String.raw`\newcommand{\fmt}[1]{\color{red}\mathbf{#1}}\mathtoolsset{prescript-sup-format=\fmt}\prescript{a}{b}{C}`],
  ['format-lap', String.raw`\mathtoolsset{prescript-sup-format=\mathclap}\prescript{abcdef}{b}{C}`],
  ['terminal-control-space', String.raw`\prescript{a\ }{b}{C}`],
  ['script-declaration', String.raw`\prescript{\color{red}}{b}{C}`],
  ['both-declarations', String.raw`\prescript{\relax}{\relax}{C}`],
  ['missing-base', String.raw`\prescript{a}{b}`],
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
fs.writeFileSync(path.join(__dirname, 'mathtools_prescript_mathjax_3_2_2.json.gz'), zlib.gzipSync(JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee',
  assets: hashes,
  scope: 'Prescript authored arguments and their formatting; complete unmodified SVGs from a fresh frozen D2 MathJax runtime per case.',
  cases,
}, null, 2) + '\n'));
