// SPDX-License-Identifier: Apache-2.0
// Capture unmodified D2 MathJax 3.2.2 output in a fresh runtime per expression.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_delimiter_color.cjs PINNED_ASSETS');
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
  [
    "fence-0-body-0",
    "\\left(\\color{red}x\\right)"
  ],
  [
    "fence-0-body-1",
    "\\left(\\color{blue}\\frac{x}{y}\\right)"
  ],
  [
    "fence-0-body-2",
    "\\left(\\color{red}x\\color{blue}y\\right)"
  ],
  [
    "fence-0-body-3",
    "\\left(\\color{red}{\\color{blue}x}y\\right)"
  ],
  [
    "fence-1-body-0",
    "\\left[\\color{red}x\\right]"
  ],
  [
    "fence-1-body-1",
    "\\left[\\color{blue}\\frac{x}{y}\\right]"
  ],
  [
    "fence-1-body-2",
    "\\left[\\color{red}x\\color{blue}y\\right]"
  ],
  [
    "fence-1-body-3",
    "\\left[\\color{red}{\\color{blue}x}y\\right]"
  ],
  [
    "fence-2-body-0",
    "\\left\\{\\color{red}x\\right\\}"
  ],
  [
    "fence-2-body-1",
    "\\left\\{\\color{blue}\\frac{x}{y}\\right\\}"
  ],
  [
    "fence-2-body-2",
    "\\left\\{\\color{red}x\\color{blue}y\\right\\}"
  ],
  [
    "fence-2-body-3",
    "\\left\\{\\color{red}{\\color{blue}x}y\\right\\}"
  ],
  [
    "fence-3-body-0",
    "\\left\\langle\\color{red}x\\right\\rangle"
  ],
  [
    "fence-3-body-1",
    "\\left\\langle\\color{blue}\\frac{x}{y}\\right\\rangle"
  ],
  [
    "fence-3-body-2",
    "\\left\\langle\\color{red}x\\color{blue}y\\right\\rangle"
  ],
  [
    "fence-3-body-3",
    "\\left\\langle\\color{red}{\\color{blue}x}y\\right\\rangle"
  ],
  [
    "fence-4-body-0",
    "\\left|\\color{red}x\\right|"
  ],
  [
    "fence-4-body-1",
    "\\left|\\color{blue}\\frac{x}{y}\\right|"
  ],
  [
    "fence-4-body-2",
    "\\left|\\color{red}x\\color{blue}y\\right|"
  ],
  [
    "fence-4-body-3",
    "\\left|\\color{red}{\\color{blue}x}y\\right|"
  ],
  [
    "fence-5-body-0",
    "\\left.\\color{red}x\\right|"
  ],
  [
    "fence-5-body-1",
    "\\left.\\color{blue}\\frac{x}{y}\\right|"
  ],
  [
    "fence-5-body-2",
    "\\left.\\color{red}x\\color{blue}y\\right|"
  ],
  [
    "fence-5-body-3",
    "\\left.\\color{red}{\\color{blue}x}y\\right|"
  ],
  [
    "fence-6-body-0",
    "\\left(\\color{red}x\\right."
  ],
  [
    "fence-6-body-1",
    "\\left(\\color{blue}\\frac{x}{y}\\right."
  ],
  [
    "fence-6-body-2",
    "\\left(\\color{red}x\\color{blue}y\\right."
  ],
  [
    "fence-6-body-3",
    "\\left(\\color{red}{\\color{blue}x}y\\right."
  ],
  [
    "outer",
    "\\color{red}\\left(x\\right)"
  ],
  [
    "outer-switch",
    "\\color{green}\\left(\\color{red}x\\right)y"
  ],
  [
    "textcolor",
    "\\textcolor{red}{\\left(\\color{blue}x\\right)}"
  ],
  [
    "textcolor-local",
    "\\left(\\textcolor{red}{x}\\right)"
  ],
  [
    "group-local",
    "\\left({\\color{red}x}\\right)"
  ],
  [
    "siblings",
    "\\left(\\color{red}x\\right)\\left[y\\right]"
  ],
  [
    "siblings-switch",
    "\\left(\\color{red}x\\right)\\left[\\color{blue}y\\right]"
  ],
  [
    "nested",
    "\\left(\\color{red}x\\left[y\\right]\\right)"
  ],
  [
    "nested-switch",
    "\\left(\\color{red}x\\left[\\color{blue}y\\right]\\right)"
  ],
  [
    "nested-follows",
    "\\left(\\left[\\color{red}x\\right]\\color{blue}y\\right)"
  ],
  [
    "empty-colored",
    "\\left(\\color{red}\\right)"
  ],
  [
    "empty-reset",
    "\\left(\\color{red}\\left[\\right]\\right)"
  ],
  [
    "infix-color-numerator",
    "\\left(\\color{red}x\\over y\\right)"
  ],
  [
    "infix-color-denominator",
    "\\left(x\\over\\color{blue}y\\right)"
  ],
  [
    "infix-colors",
    "\\left(\\color{red}x\\over\\color{blue}y\\right)"
  ],
  [
    "font",
    "\\left(\\color{red}x\\rm y\\right)"
  ],
  [
    "font-first",
    "\\left(\\rm\\color{red}x\\right)"
  ],
  [
    "size",
    "\\left(\\Huge\\color{red}x\\right)"
  ],
  [
    "style",
    "\\left(\\scriptstyle\\color{red}x\\right)"
  ],
  [
    "index",
    "\\sqrt[\\left(\\color{red}n\\right)]{x}"
  ],
  [
    "script",
    "x_{\\left(\\color{red}y\\right)}"
  ],
  [
    "fraction",
    "\\frac{\\left(\\color{red}x\\right)}{\\left[\\color{blue}y\\right]}"
  ],
  [
    "raise-color",
    "\\left(\\raise1em\\color{red}x\\right)"
  ],
  [
    "color-raise-color",
    "\\left(\\color{red}\\raise1em\\color{blue}x\\right)"
  ],
  [
    "declared-color",
    "\\definecolor{c}{RGB}{20,60,150}\\left(\\color{c}x\\right)"
  ],
  [
    "model",
    "\\left(\\color[rgb]{0.8,0.1,0.2}x\\right)"
  ],
  [
    "hex",
    "\\left(\\color{#336699}x\\right)"
  ],
  [
    "matrix-scope",
    "\\begin{matrix}\\left(\\color{red}x\\right)&\\left[y\\right]\\end{matrix}"
  ],
  [
    "plain-control",
    "\\left(x\\right)"
  ],
  [
    "middle-control",
    "\\left(x\\middle|y\\right)"
  ],
  [
    "bad-right",
    "\\left(\\color{red}x\\right\\undefined"
  ],
  [
    "missing-right",
    "\\left(\\color{red}x"
  ],
  [
    "outer-siblings",
    "\\color{green}\\left(\\color{red}x\\right)\\left[y\\right]"
  ],
  [
    "outer-nested-switch",
    "\\color{green}\\left(\\color{red}x\\left[\\color{blue}y\\right]\\right)"
  ],
  [
    "outer-infix-colors",
    "\\color{green}\\left(\\color{red}x\\over\\color{blue}y\\right)"
  ],
  [
    "outer-matrix-scope",
    "\\color{green}\\begin{matrix}\\left(\\color{red}x\\right)&\\left[y\\right]\\end{matrix}"
  ]
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
fs.writeFileSync(path.join(__dirname, 'delimiter_color_mathjax_3_2_2.json'), JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee',
  assets: hashes,
  scope: 'Complete unmodified SVGs from the frozen D2 MathJax bundle; each case uses a fresh runtime.',
  cases,
}, null, 2) + '\n');
