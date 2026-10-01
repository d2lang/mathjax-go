// SPDX-License-Identifier: Apache-2.0
// Capture finite original BBox arithmetic and complete SVGs in fresh frozen VMs.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const zlib = require('node:zlib');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_bbox_arithmetic.cjs PINNED_ASSETS');
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
function fresh() {
  const context = vm.createContext({console});
  context.globalThis = context;
  for (const script of scripts) script.runInContext(context);
  return context;
}
const shapes = [
  {w: .572, h: .442, d: .011, L: 0, R: 0},
  {w: .9018046532517446, h: .6755411972844539, d: .205, L: 0, R: 0},
  {w: 1.1419467888754755, h: .6762483040656405, d: .011, L: 0, R: 0},
  {w: 1.2926721856198917, h: .8906797615707167, d: .011, L: .1, R: .2},
  {w: .465, h: .442, d: .011, L: .0556, R: .0278},
];
const scales = [.7071067811865475, 1.4142135623730951, .4999999999999999, .9, 1.2];
const offsets = [[0, 0], [.1, .413], [.3, .3625], [.572, .25], [1.056, -.3625], [-.2, -.413]];
const bboxCases = [];
let combineSource, appendSource;
for (const [shapeIndex, shape] of shapes.entries()) {
  for (const [scaleIndex, rscale] of scales.entries()) {
    for (const [offsetIndex, [x, y]] of offsets.entries()) {
      for (const method of ['combine', 'append']) {
        const context = fresh();
        const BBox = context.MathJax._.util.BBox.BBox;
        combineSource = BBox.prototype.combine.toString();
        appendSource = BBox.prototype.append.toString();
        const parent = {w: method === 'append' ? x : 0, h: -1000000, d: -1000000};
        const child = {...shape, rscale};
        const target = Object.assign(new BBox(), parent);
        const argument = Object.assign(new BBox(), child);
        if (method === 'append') target.append(argument);
        else target.combine(argument, x, y);
        const expected = {w: target.w, h: target.h, d: target.d};
        if (![...Object.values(expected), x, y, rscale].every(Number.isFinite)) throw new Error('non-finite capture');
        bboxCases.push({name: `${method}-${shapeIndex}-${scaleIndex}-${offsetIndex}`, method, parent, child, x, y, expected});
      }
    }
  }
}
const body = String.raw`{\displaystyle x^{y^z}}\\{\textstyle z^{w^v}}`;
const forms = [
  ['crampedsubstack', `\\crampedsubstack{${body}}`],
  ['crampedsubarray', `\\begin{crampedsubarray}{c}${body}\\end{crampedsubarray}`],
  ['subarray', `\\begin{subarray}{c}${body}\\end{subarray}`],
];
const svgCases = [];
for (const [formName, form] of forms) {
  for (const [contextName, tex] of [
    ['plain', form], ['bold', `\\mathbf{${form}}`], ['sans', `\\mathsf{${form}}`],
    ['cal', `\\mathcal{${form}}`], ['bold-normal', `\\bf\\mathnormal{${form}}`],
    ['cramped-textstyle', `\\cramped[\\textstyle]{${form}}`], ['denominator', `\\frac{x}{${form}}`],
    ['numerator', `\\frac{${form}}{x}`], ['root', `\\sqrt{${form}}`], ['fenced', `\\left(${form}\\right)`],
    ['subscript', `x_{${form}}`], ['superscript', `x^{${form}}`], ['large', '\\Large ' + form],
    ['small', '\\tiny ' + form], ['color', '\\color{red}' + form], ['boxed', `\\boxed{${form}}`],
  ]) {
    for (const display of [false, true]) {
      const context = fresh();
      const svg = context.adaptor.innerHTML(context.html.convert(tex, {em: 16, ex: 8, display}));
      if (svg.includes('data-mml-node="merror"')) throw new Error(`original rejected ${contextName}`);
      svgCases.push({name: `${formName}-${contextName}-${display ? 'display' : 'inline'}`, tex, display, svg});
    }
  }
}
fs.writeFileSync(path.join(__dirname, 'bbox_arithmetic_mathjax_3_2_2.json.gz'), zlib.gzipSync(JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee', assets: hashes,
  scope: 'Finite original BBox.combine/append observations and complete SVGs; every case uses a fresh frozen D2 runtime. Shape values include original nested-script bounding boxes.',
  combineSource, appendSource, bboxCases, svgCases,
}, null, 2) + '\n'));
