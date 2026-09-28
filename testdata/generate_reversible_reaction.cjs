// SPDX-License-Identifier: Apache-2.0
// Capture unmodified D2 MathJax 3.2.2 output in a fresh runtime per expression.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_reversible_reaction.cjs PINNED_ASSETS');
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
const rule = String.raw`\Rule{1em}{.5em}{.2em}`;
const inputs = [
  ['reaction', String.raw`\ce{N2O4 <--> 2NO2}`],
  ['reaction-top-label', String.raw`\ce{A <-->[H2O] B}`],
  ['reaction-top-math', String.raw`\ce{A <-->[$k$] B}`],
  ['reaction-top-text', String.raw`\ce{A <-->[{catalyst}] B}`],
  ['reaction-sequence', String.raw`\ce{A -> B <--> C}`],
  ['reaction-color', String.raw`\color{blue}\ce{A <--> B}`],
  ['reaction-script', String.raw`K_{\ce{A <--> B}}`],
  ['direct-arrow', String.raw`A\longleftrightarrows B`],
  ['direct-arrow-limits', String.raw`A\overset{k}{\longleftrightarrows}B`],
  ['arrow-override', String.raw`\DeclareMathOperator{\longleftrightarrows}{rev}\ce{A <--> B}`],
  ['rule', 'a' + rule + 'b'],
  ['rule-zero-width', String.raw`a\Rule{0px}{.25em}{0px}b`],
  ['rule-depth', String.raw`a\Rule{1em}{0em}{.5em}b`],
  ['rule-zero', String.raw`a\Rule{0em}{0em}{0em}b`],
  ['rule-unbraced', String.raw`a\Rule1em .5em .2em b`],
  ['rule-no-spaces', String.raw`a\Rule1em.5em.2em b`],
  ['rule-physical-units', String.raw`a\Rule{1cm}{2mm}{3pt}b`],
  ['rule-ex-mu', String.raw`a\Rule{1ex}{2mu}{3mu}b`],
  ['rule-comma-decimals', String.raw`a\Rule{1,5em}{0,5em}{0,2em}b`],
  ['rule-negative-width', String.raw`a\Rule{-1em}{.5em}{.2em}b`],
  ['rule-negative-height', String.raw`a\Rule{1em}{-.5em}{.2em}b`],
  ['rule-negative-depth', String.raw`a\Rule{1em}{.5em}{-.2em}b`],
  ['rule-script', 'a_{' + rule + '}'],
  ['rule-large', String.raw`\Large a` + rule + 'b'],
  ['space', String.raw`a\Space{1em}{.5em}{.2em}b`],
  ['space-unbraced', String.raw`a\Space1em .5em .2em b`],
  ['space-color', String.raw`\color{red}a\Space{1em}{.5em}{.2em}b`],
  ['space-script', String.raw`x^{a\Space{1em}{.5em}{.2em}b}`],
  ['missing-all', String.raw`\Rule`],
  ['missing-height', String.raw`\Rule{1em}`],
  ['missing-depth', String.raw`\Rule{1em}{.5em}`],
  ['missing-width-units', String.raw`\Rule{1}{.5em}{.2em}`],
  ['invalid-height', String.raw`\Rule{1em}{bad}{.2em}`],
  ['invalid-depth', String.raw`\Rule{1em}{.5em}{bad}`],
  ['unclosed-dimension', String.raw`\Rule{1em`],
  ['space-error', String.raw`\Space{1em}{2}`],
  ['color-declaration', String.raw`\color{red}` + rule],
  ['textcolor', String.raw`\textcolor{blue}{` + rule + '}'],
  ['color-custom', String.raw`\definecolor{custom}{RGB}{12,34,56}\color{custom}` + rule],
  ['color-group-restores', '{' + String.raw`\color{red}` + rule + '}' + rule],
  ['color-nested-restores', String.raw`\color{red}{\color{blue}` + rule + '}' + rule],
  ['textcolor-restores', String.raw`\color{red}\textcolor{blue}{` + rule + '}' + rule],
  ['array-reset-restores', String.raw`\color{red}\begin{matrix}` + rule + String.raw`\end{matrix}` + rule],
  ['array-cell-reset', String.raw`\begin{matrix}\color{red}` + rule + '&' + rule + String.raw`\end{matrix}`],
  ['aligned-reset', String.raw`\color{red}\begin{aligned}` + rule + String.raw`\end{aligned}`],
  ['multlined-reset', String.raw`\color{red}\begin{multlined}` + rule + String.raw`\end{multlined}`],
  ['physics-array-reset', String.raw`\color{red}\mqty{` + rule + '}'],
  ['boxed-reset-restores', String.raw`\color{red}\boxed{` + rule + '}' + rule],
  ['internal-math-reset', String.raw`\color{red}\text{$` + rule + '$}'],
  ['fraction-inherits', String.raw`\color{red}\frac{` + rule + '}{x}'],
  ['infix-inherits', String.raw`\color{red}` + rule + String.raw`\over ` + rule],
  ['left-right-inherits', String.raw`\color{red}\left(` + rule + String.raw`\right)`],
  ['operator-inherits', String.raw`\color{red}\operatorname{` + rule + '}'],
  ['font-inherits', String.raw`\color{red}\mathrm{` + rule + '}'],
  ['vector-inherits', String.raw`\color{red}\vb{` + rule + '}'],
  ['rule-override', String.raw`\DeclareMathOperator{\Rule}{R}\Rule{1em}{.5em}{.2em}`],
  ['forward-reaction-control', String.raw`\ce{2H2 + O2 -> 2H2O}`],
  ['backward-reaction-control', String.raw`\ce{A <- B}`],
  ['two-headed-reaction-control', String.raw`\ce{A <-> B}`],
  ['horizontal-space-control', String.raw`a\kern1em b`],
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
fs.writeFileSync(path.join(__dirname, 'reversible_reaction_mathjax_3_2_2.json'), JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee',
  assets: hashes,
  scope: 'Complete unmodified SVGs from the frozen D2 MathJax bundle; each case uses a fresh runtime.',
  cases,
}, null, 2) + '\n');
