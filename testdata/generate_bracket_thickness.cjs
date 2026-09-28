// SPDX-License-Identifier: Apache-2.0
// Capture unmodified D2 MathJax 3.2.2 output in a fresh runtime per expression.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_bracket_thickness.cjs PINNED_ASSETS');
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
    "overbracket-0",
    "\\overbracket[.1em][1ex]{x+y}"
  ],
  [
    "overbracket-1",
    "\\overbracket[2pt][1ex]{x+y}"
  ],
  [
    "overbracket-2",
    "\\overbracket[3px][1ex]{x+y}"
  ],
  [
    "overbracket-3",
    "\\overbracket[1ex][1ex]{x+y}"
  ],
  [
    "overbracket-4",
    "\\overbracket[2mu][1ex]{x+y}"
  ],
  [
    "overbracket-5",
    "\\overbracket[.25pc][1ex]{x+y}"
  ],
  [
    "overbracket-6",
    "\\overbracket[.03in][1ex]{x+y}"
  ],
  [
    "overbracket-7",
    "\\overbracket[2mm][1ex]{x+y}"
  ],
  [
    "overbracket-8",
    "\\overbracket[.1cm][1ex]{x+y}"
  ],
  [
    "overbracket-9",
    "\\overbracket[200%][1ex]{x+y}"
  ],
  [
    "overbracket-10",
    "\\overbracket[2][1ex]{x+y}"
  ],
  [
    "overbracket-11",
    "\\overbracket[][1ex]{x+y}"
  ],
  [
    "overbracket-12",
    "\\overbracket[0][1ex]{x+y}"
  ],
  [
    "overbracket-13",
    "\\overbracket[-.02em][1ex]{x+y}"
  ],
  [
    "overbracket-14",
    "\\overbracket[thin][1ex]{x+y}"
  ],
  [
    "overbracket-15",
    "\\overbracket[medium][1ex]{x+y}"
  ],
  [
    "overbracket-16",
    "\\overbracket[thick][1ex]{x+y}"
  ],
  [
    "overbracket-17",
    "\\overbracket[thinmathspace][1ex]{x+y}"
  ],
  [
    "overbracket-18",
    "\\overbracket[nope][1ex]{x+y}"
  ],
  [
    "overbracket-19",
    "\\overbracket[2ptgarbage][1ex]{x+y}"
  ],
  [
    "overbracket-20",
    "\\overbracket[1e2em][1ex]{x+y}"
  ],
  [
    "overbracket-21",
    "\\overbracket[ 2pt][1ex]{x+y}"
  ],
  [
    "overbracket-22",
    "\\overbracket[ 2pt][1ex]{x+y}"
  ],
  [
    "overbracket-23",
    "\\overbracket[﻿2pt][1ex]{x+y}"
  ],
  [
    "overbracket-24",
    "\\overbracket[2 pt][1ex]{x+y}"
  ],
  [
    "overbracket-25",
    "\\overbracket[pt][1ex]{x+y}"
  ],
  [
    "overbracket-26",
    "\\overbracket[.0006em][1ex]{x+y}"
  ],
  [
    "overbracket-27",
    "\\overbracket[.0015em][1ex]{x+y}"
  ],
  [
    "overbracket-28",
    "\\overbracket[.0125em][1ex]{x+y}"
  ],
  [
    "underbracket-0",
    "\\underbracket[.1em][1ex]{x+y}"
  ],
  [
    "underbracket-1",
    "\\underbracket[2pt][1ex]{x+y}"
  ],
  [
    "underbracket-2",
    "\\underbracket[3px][1ex]{x+y}"
  ],
  [
    "underbracket-3",
    "\\underbracket[1ex][1ex]{x+y}"
  ],
  [
    "underbracket-4",
    "\\underbracket[2mu][1ex]{x+y}"
  ],
  [
    "underbracket-5",
    "\\underbracket[.25pc][1ex]{x+y}"
  ],
  [
    "underbracket-6",
    "\\underbracket[.03in][1ex]{x+y}"
  ],
  [
    "underbracket-7",
    "\\underbracket[2mm][1ex]{x+y}"
  ],
  [
    "underbracket-8",
    "\\underbracket[.1cm][1ex]{x+y}"
  ],
  [
    "underbracket-9",
    "\\underbracket[200%][1ex]{x+y}"
  ],
  [
    "underbracket-10",
    "\\underbracket[2][1ex]{x+y}"
  ],
  [
    "underbracket-11",
    "\\underbracket[][1ex]{x+y}"
  ],
  [
    "underbracket-12",
    "\\underbracket[0][1ex]{x+y}"
  ],
  [
    "underbracket-13",
    "\\underbracket[-.02em][1ex]{x+y}"
  ],
  [
    "underbracket-14",
    "\\underbracket[thin][1ex]{x+y}"
  ],
  [
    "underbracket-15",
    "\\underbracket[medium][1ex]{x+y}"
  ],
  [
    "underbracket-16",
    "\\underbracket[thick][1ex]{x+y}"
  ],
  [
    "underbracket-17",
    "\\underbracket[thinmathspace][1ex]{x+y}"
  ],
  [
    "underbracket-18",
    "\\underbracket[nope][1ex]{x+y}"
  ],
  [
    "underbracket-19",
    "\\underbracket[2ptgarbage][1ex]{x+y}"
  ],
  [
    "underbracket-20",
    "\\underbracket[1e2em][1ex]{x+y}"
  ],
  [
    "underbracket-21",
    "\\underbracket[ 2pt][1ex]{x+y}"
  ],
  [
    "underbracket-22",
    "\\underbracket[ 2pt][1ex]{x+y}"
  ],
  [
    "underbracket-23",
    "\\underbracket[﻿2pt][1ex]{x+y}"
  ],
  [
    "underbracket-24",
    "\\underbracket[2 pt][1ex]{x+y}"
  ],
  [
    "underbracket-25",
    "\\underbracket[pt][1ex]{x+y}"
  ],
  [
    "underbracket-26",
    "\\underbracket[.0006em][1ex]{x+y}"
  ],
  [
    "underbracket-27",
    "\\underbracket[.0015em][1ex]{x+y}"
  ],
  [
    "underbracket-28",
    "\\underbracket[.0125em][1ex]{x+y}"
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
fs.writeFileSync(path.join(__dirname, 'bracket_thickness_mathjax_3_2_2.json'), JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee',
  assets: hashes,
  scope: 'Complete unmodified SVGs from the frozen D2 MathJax bundle; each case uses a fresh runtime.',
  cases,
}, null, 2) + '\n');
