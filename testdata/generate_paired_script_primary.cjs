// SPDX-License-Identifier: Apache-2.0
// Capture unmodified frozen MathJax 3.2.2 output in a fresh VM per input.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const zlib = require('node:zlib');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_paired_script_primary.cjs PINNED_ASSETS');
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
const inputs = [];
const bases = [['letter', 'x'], ['sum', '\\sum'], ['sum-nolimits', '\\sum\\nolimits'], ['lim', '\\lim'], ['int', '\\int']];
const subs = [
  ['letter', 'y'], ['fraction', '\\frac{a}{b}'], ['display-fraction', '\\dfrac{a}{b}'],
  ['substack', '\\substack{x^2\\\\y^3}'], ['crampedsubstack', '\\crampedsubstack{x^2\\\\y^3}'],
  ['fenced-fraction', '\\left(\\frac ab\\right)'], ['phantom', '\\vphantom{\\dfrac ab}i'],
  ['root', '\\sqrt{y}'], ['smash', '\\smash{\\dfrac ab}'],
];
const sups = [['letter', 'n'], ['fraction', '\\frac ab'], ['phantom', '\\vphantom{\\dfrac ab}n']];
for (const [baseName, base] of bases) {
  for (const [subName, sub] of subs) {
    for (const [supName, sup] of sups) {
      inputs.push([`${baseName}-${subName}-${supName}`, `${base}_{${sub}}^{${sup}}x`]);
    }
  }
}
const script = String.raw`x_{\crampedsubstack{a\\b\\c}}^n`;
const sum = String.raw`\sum_{\crampedsubstack{a\\b\\c}}^n x`;
for (const [name, tex] of [
  ['three-row', script], ['textstyle-sum', '\\textstyle ' + sum],
  ['textstyle-product', String.raw`\textstyle \prod_{\crampedsubstack{a\\b\\c}}^n x`],
  ['bold', `\\mathbf{${script}}`], ['sans', `\\mathsf{${script}}`],
  ['color', `\\color{red}${script}`], ['root', `\\sqrt{${script}}`],
  ['denominator', `\\frac{1}{${script}}`], ['numerator', `\\frac{${script}}{2}`],
  ['superscript', `z^{${script}}`], ['subscript', `z_{${script}}`],
  ['scriptstyle', '\\scriptstyle ' + script], ['scriptscriptstyle', '\\scriptscriptstyle ' + script],
  ['large', '\\Large ' + script], ['small', '\\tiny ' + script],
  ['reversed', String.raw`x^n_{\crampedsubstack{a\\b\\c}}`],
  ['sup-only-control', String.raw`x^{\crampedsubstack{a\\b\\c}}`],
  ['sub-only-control', String.raw`x_{\crampedsubstack{a\\b\\c}}`],
  ['display-limit-control', String.raw`\displaystyle \sum_{\crampedsubstack{a\\b\\c}}^n x`],
  ['d2-witness', script + String.raw`\quad\textstyle ` + sum],
]) inputs.push([name, tex]);
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
fs.writeFileSync(path.join(__dirname, 'paired_script_primary_mathjax_3_2_2.json.gz'), zlib.gzipSync(JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee', assets: hashes,
  scope: 'Paired scripts inherit CommonScriptbase.scriptChild (child 1) for their initial superscript shift; all complete unmodified SVGs use a fresh frozen D2 MathJax runtime.',
  cases,
}, null, 2) + '\n'));
