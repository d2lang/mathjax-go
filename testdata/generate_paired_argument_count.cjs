// SPDX-License-Identifier: Apache-2.0
// Bind paired declaration counts to unmodified original registrations and SVGs.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const zlib = require('node:zlib');
const {spawnSync} = require('node:child_process');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_paired_argument_count.cjs PINNED_ASSETS');
const hashes = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
let mathjaxSource;
const scripts = Object.entries(hashes).map(([file, hash]) => {
  const bytes = fs.readFileSync(path.join(assets, file));
  if (crypto.createHash('sha256').update(bytes).digest('hex') !== hash) throw new Error(`unverified ${file}`);
  if (file === 'mathjax.js') mathjaxSource = bytes.toString();
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
    const macro = context.html.inputJax[0].parseOptions.handlers.get('macro').lookup('pair');
    return {...record, svg, registration: macro ? macro.args : null,
      countType: macro ? typeof macro.args[3] : 'unregistered'};
  });
  fs.writeFileSync(1, JSON.stringify(cases));
  process.exit(0);
}
const pending = [];
function add(name, tex, renderedError = false) {
  for (const display of [false, true]) pending.push({name: name + (display ? '-display' : '-inline'), tex, display, renderedError});
}
const aliases = ['DeclarePairedDelimiterX', 'DeclarePairedDelimitersX', 'DeclarePairedDelimiterXPP', 'DeclarePairedDelimitersXPP'];
function declaration(alias, count, body) {
  const brackets = count === null ? '' : '[' + count + ']';
  return '\\' + alias + '{\\pair}' + brackets + (alias.endsWith('XPP') ? '{p}{(}{)}{r}' : '{(}{)}') + '{' + body + '}';
}
const tail = '{x}{y}{z}{a}{b}{c}{d}{e}{f}{g}';
const counts = [
  ['absent', null, false], ['empty', '', false], ['zero', '0', false], ['zeroes', '00', false],
  ['one', '1', false], ['two', '2', false], ['ten', '10', false],
  ['space-only', ' ', true], ['tab-only', '\t', true], ['signed-plus', '+1', true],
  ['signed-minus', '-1', true], ['decimal', '1.0', true], ['suffix', '1x', true],
  ['hex', '0x1', true], ['non-ascii-digit', '\u0661', true],
];
const whites = [
  ['space', ' '], ['tab', '\t'], ['line-feed', '\n'], ['vertical-tab', '\v'],
  ['form-feed', '\f'], ['carriage-return', '\r'], ['nbsp', '\u00a0'], ['ogham', '\u1680'],
  ['en-quad', '\u2000'], ['em-quad', '\u2001'], ['en-space', '\u2002'], ['em-space', '\u2003'],
  ['third-em', '\u2004'], ['quarter-em', '\u2005'], ['sixth-em', '\u2006'], ['figure-space', '\u2007'],
  ['punctuation-space', '\u2008'], ['thin-space', '\u2009'], ['hair-space', '\u200a'],
  ['line-separator', '\u2028'], ['paragraph-separator', '\u2029'], ['narrow-nbsp', '\u202f'],
  ['medium-space', '\u205f'], ['ideographic-space', '\u3000'], ['bom', '\ufeff'],
  ['nel-control', '\u0085'], ['mvs-control', '\u180e'], ['zwsp-control', '\u200b'],
];
for (const alias of aliases) {
  for (const [name, count, error] of counts) add(alias + '-' + name, declaration(alias, count, 'q') + '\\pair' + tail, error);
  for (const [name, white] of whites) add(alias + '-' + name, declaration(alias, white + '1' + white, '#1') + '\\pair{x}', name.endsWith('-control'));
  for (const [name, count] of [['absent', null], ['empty', ''], ['zero', '0']]) {
    add(alias + '-parameter-' + name, declaration(alias, count, '#1') + '\\pair{x}', name !== 'absent');
  }
  for (const [name, count] of [['absent', null], ['empty', ''], ['zero', '0'], ['two', '2']]) {
    add(alias + '-literal-hash-' + name, declaration(alias, count, '\\text{##}') + '\\pair' + tail);
  }
  add(alias + '-escaped-hash-control', declaration(alias, '0', '\\#') + '\\pair{x}');
  if (alias.endsWith('XPP')) {
    for (const [name, count] of [['empty', ''], ['zero', '0']]) {
      add(alias + '-outer-hash-' + name, '\\' + alias + '{\\pair}[' + count + ']{\\text{##}}{(}{)}{\\text{##}}{q}\\pair{x}');
    }
  }
  for (const [name, count] of [['absent', null], ['empty', ''], ['bom', '\ufeff1\ufeff']]) {
    const prefix = declaration(alias, count, 'q');
    for (const [context, invocation] of [['star', '\\pair*'], ['big', '\\pair[\\Big]'], ['bold', '\\mathbf{\\pair' + tail + '}']]) {
      add(alias + '-' + name + '-' + context, prefix + invocation + (context === 'bold' ? '' : tail));
    }
  }
}
add('d2-empty-count-witness', declaration('DeclarePairedDelimiterX', '', 'q') + '\\pair{\\dfrac{a+b}{c+d}}');
add('d2-bom-count-witness', declaration('DeclarePairedDelimiterX', '\ufeff1\ufeff', '#1') + '\\pair{\\dfrac{a+b}{c+d}}');
add('d2-zero-count-hash-witness', declaration('DeclarePairedDelimiterX', '0', '\\text{##}') + '\\pair{x}');
const match = mathjaxSource.match(/t\.GetArgCount=function\(t,e\)\{[\s\S]*?(?=,t\.GetTemplate=)/);
if (!match) throw new Error('missing pinned GetArgCount implementation');
const context = fresh();
context.html.convert(declaration(aliases[0], '1', 'q'), {em: 16, ex: 8, display: false});
const pairedMethod = context.html.inputJax[0].parseOptions.handlers.get('macro').lookup('pair').func;
const cases = [];
for (let i = 0; i < pending.length; i += 32) {
  const worker = spawnSync(process.execPath, [__filename, assets, '--worker'], {input: JSON.stringify(pending.slice(i, i + 32)), encoding: 'utf8', maxBuffer: 16 * 1024 * 1024});
  if (worker.error || worker.status !== 0) throw worker.error || new Error(worker.stderr || `oracle worker ${worker.status}`);
  cases.push(...JSON.parse(worker.stdout));
}
fs.writeFileSync(path.join(__dirname, 'paired_argument_count_mathjax_3_2_2.json.gz'), zlib.gzipSync(JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee', assets: hashes,
  getArgCountSource: match[0], pairedDelimitersSource: pairedMethod.toString(),
  scope: 'Complete fresh original SVGs and active macro registration arguments for finite paired declaration counts. Rendered-error diagnostics are marked separately; enormous allocation inputs are not exercised.', cases,
}, null, 2) + '\n'));
