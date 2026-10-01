// SPDX-License-Identifier: Apache-2.0
// Capture the capital Braket macro and lowercase Physics handlers independently.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const zlib = require('node:zlib');
const {spawnSync} = require('node:child_process');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_capital_ketbra.cjs PINNED_ASSETS');
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
if (process.argv[3] === '--worker') {
  const cases = JSON.parse(fs.readFileSync(0, 'utf8')).map(record => {
    const context = fresh();
    const svg = context.adaptor.innerHTML(context.html.convert(record.tex, {em: 16, ex: 8, display: record.display}));
    if (svg.includes('data-mml-node="merror"') !== record.renderedError) throw new Error(`original validity differs for ${record.name}`);
    return {...record, svg};
  });
  fs.writeFileSync(1, JSON.stringify(cases));
  process.exit(0);
}
const pending = [];
function add(name, tex, renderedError = false) {
  for (const display of [false, true]) pending.push({name: name + (display ? '-display' : '-inline'), tex, display, renderedError});
}
const pairs = [
  ['letters', 'x', 'y'], ['greek', '\\psi', '\\phi'], ['empty', '', ''],
  ['empty-ket', '', 'y'], ['empty-bra', 'x', ''],
  ['fraction', '\\frac{a}{b}', '\\frac{c}{d}'], ['radical', '\\sqrt{x}', '\\sqrt{y}'],
  ['tall', '\\dfrac{a}{b}', '\\dfrac{c}{d}'],
  ['nested', '\\Ketbra{x}{y}', '\\Bra{z}'], ['families', '\\mathrm{x}', '\\mathbf{y}'],
  ['bars', 'x|y', 'z\\|w'], ['matrix', '\\begin{pmatrix}a\\\\b\\end{pmatrix}', 'y'],
];
for (const [name, left, right] of pairs) {
  const form = `\\Ketbra{${left}}{${right}}`;
  for (const [context, tex] of [
    ['plain', form], ['bold', `\\mathbf{${form}}`], ['subscript', `z_{${form}}`], ['root', `\\sqrt{${form}}`],
  ]) add(name + '-' + context, tex);
}
for (const alias of ['ketbra', 'outerproduct', 'dyad', 'op']) {
  for (const [name, tail] of [
    ['pair', '{x}{y}'], ['empty-second', '{x}{}'], ['optional-second', '{x}'],
    ['star-pair', '*{x}{y}'], ['star-optional', '*{x}'], ['unbraced', ' x y'],
    ['nested-capital', '{\\Ketbra{x}{y}}{z}'], ['tall', '{\\dfrac{a}{b}}{\\dfrac{c}{d}}'],
  ]) add('physics-' + alias + '-' + name, '\\' + alias + tail);
}
for (const [name, tex, renderedError = false] of [
  ['tail', '\\Ketbra{x}{y}z'], ['adjacent', '\\Ketbra{x}{y}\\Ketbra{z}{w}'],
  ['unbraced', '\\Ketbra x y'], ['unbraced-control', '\\Ketbra\\frac{a}{b}x', true],
  ['unbraced-second', '\\Ketbra{x}y'], ['literal-star', '\\Ketbra*{x}{y}'],
  ['literal-brackets', '\\Ketbra[x]{y}'],
  ['paired-priority-control', '\\DeclarePairedDelimiter{\\Ketbra}{(}{)}\\Ketbra{x}'],
]) add(name, tex, renderedError);
for (const [name, tex] of [
  ['missing-both', '\\Ketbra'], ['missing-second', '\\Ketbra{x}'],
  ['unclosed-first', '\\Ketbra{x'], ['unclosed-second', '\\Ketbra{x}{'],
]) add(name, tex, true);
const context = fresh();
const handlers = context.html.inputJax[0].parseOptions.handlers.get('macro');
const capital = handlers.lookup('Ketbra'), lower = handlers.lookup('ketbra');
if (capital.func !== context.MathJax._.input.tex.base.BaseMethods.default.Macro || lower.func !== context.MathJax._.input.tex.physics.PhysicsMethods.default.KetBra) throw new Error('unexpected active command precedence');
const registration = {symbol: capital.symbol, args: capital.args, handler: 'BaseMethods.Macro', lowerHandler: 'PhysicsMethods.KetBra', macroSource: capital.func.toString()};
const cases = [];
for (let i = 0; i < pending.length; i += 32) {
  const worker = spawnSync(process.execPath, [__filename, assets, '--worker'], {input: JSON.stringify(pending.slice(i, i + 32)), encoding: 'utf8', maxBuffer: 16 * 1024 * 1024});
  if (worker.error || worker.status !== 0) throw worker.error || new Error(worker.stderr || `oracle worker ${worker.status}`);
  cases.push(...JSON.parse(worker.stdout));
}
fs.writeFileSync(path.join(__dirname, 'capital_ketbra_mathjax_3_2_2.json.gz'), zlib.gzipSync(JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee', assets: hashes, registration,
  scope: 'Complete original capital Braket macro SVGs, Physics lowercase controls, and separately marked rendered-error diagnostics. Every expression uses a fresh frozen D2 runtime.', cases,
}, null, 2) + '\n'));
