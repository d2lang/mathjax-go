// SPDX-License-Identifier: Apache-2.0
// Fresh, unmodified D2 MathJax 3.2.2 references; no Go output is consulted.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const zlib = require('node:zlib');
const assets = process.argv[2];
if (!assets) throw Error('usage: node generate_mathtools_paired_option.cjs PINNED_ASSETS');
const hashes = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
const scripts = Object.entries(hashes).map(([name, digest]) => {
  const bytes = fs.readFileSync(path.join(assets, name));
  if (crypto.createHash('sha256').update(bytes).digest('hex') !== digest) throw Error('asset ' + name);
  return new vm.Script(bytes.toString(), {filename: name});
});
function fresh() {
  const context = vm.createContext({console});
  context.globalThis = context;
  for (const script of scripts) script.runInContext(context);
  return context;
}
const forms = [
  ['bare', 'pairedDelimiters'], ['empty', 'pairedDelimiters='],
  ['empty-braces', 'pairedDelimiters={}'], ['blank-braces', 'pairedDelimiters={ }'],
  ['true', 'pairedDelimiters=true'], ['false', 'pairedDelimiters=false'],
  ['upper-true', 'pairedDelimiters=TRUE'], ['upper-false', 'pairedDelimiters=FALSE'],
  ['zero', 'pairedDelimiters=0'], ['one', 'pairedDelimiters=1'],
  ['other-string', 'pairedDelimiters=x'], ['nested', 'pairedDelimiters={{x}}'],
  ['comma-in-braces', 'pairedDelimiters={left,right}'],
  ['documented-tuple-string', 'pairedDelimiters={example: [left,right,body,1,pre,post]}'],
  ['duplicate-last-false', 'pairedDelimiters=x,pairedDelimiters=false'],
  ['duplicate-last-string', 'pairedDelimiters=false,pairedDelimiters=x'],
];
const pair = String.raw`\DeclarePairedDelimitersX{\pair}[1]{(}{)}{#1}`;
const contexts = [
  ['plain', raw => String.raw`\mathtoolsset{${raw}}a=b`],
  ['existing-delimiter', raw => pair + String.raw`\mathtoolsset{${raw}}\pair*{\frac{1}{x^2+1}}`],
  ['later-delimiter', raw => String.raw`\mathtoolsset{${raw}}` + pair + String.raw`\pair[\Bigg]{\frac{1}{x^2+1}}`],
  ['scalar-after', raw => String.raw`\mathtoolsset{${raw},centercolon=true,use-unicode=true}a:=b\coloneqq c`],
  ['scalar-before', raw => String.raw`\mathtoolsset{centercolon=true,use-unicode=true,${raw}}a:=b\coloneqq c`],
  ['fraction', raw => String.raw`\frac{\mathtoolsset{${raw}}a+b}{c+d}`],
  ['script', raw => String.raw`x_{\mathtoolsset{${raw}}a+b}^{c+d}`],
  ['alignment', raw => String.raw`\begin{align*}\mathtoolsset{${raw}}a&=b\\c&=d\end{align*}`],
];
const inputs = [];
for (const [name, raw] of forms) {
  for (const [contextName, make] of contexts) inputs.push([name + '-' + contextName, make(raw)]);
}
for (const [name, tex] of [
  ['plain-control', 'a=b'], ['delimiter-control', pair + String.raw`\pair*{\frac{1}{x^2+1}}`],
  ['scalar-control', String.raw`\mathtoolsset{centercolon=true,use-unicode=true}a:=b\coloneqq c`],
  ['empty-options-control', String.raw`\mathtoolsset{}a=b`],
  ['delimiter-body-control', String.raw`\DeclarePairedDelimitersX{\pair}[1]{(}{)}{\mathtoolsset{pairedDelimiters={}}#1}\pair{x}`],
  ['delimsize-control', String.raw`\mathtoolsset{pairedDelimiters=false}\DeclarePairedDelimiterX{\pair}[2]{\langle}{\rangle}{#1\delimsize\vert#2}\pair*{\frac12}{y}`],
  ['xpp-control', String.raw`\mathtoolsset{pairedDelimiters={foo}}\DeclarePairedDelimitersXPP{\pair}[1]{P}{(}{)}{_2}{#1}\pair{x}`],
  ['reset-does-not-rebuild', pair + String.raw`\mathtoolsset{pairedDelimiters=false}\pair{x}\mathtoolsset{pairedDelimiters={other}}\pair{y}`],
  ['tagform-independent', String.raw`\mathtoolsset{pairedDelimiters={}}\newtagform{brackets}{[}{]}\usetagform{brackets}\begin{equation}a=b\tag{2}\end{equation}`],
  ['forbidden-tagforms', String.raw`\mathtoolsset{tagforms={}}a=b`],
  ['forbidden-allow', String.raw`\mathtoolsset{allow-mathtoolsset=false}a=b`],
  ['typo-remains-unknown', String.raw`\mathtoolsset{pariedDelimiters={}}a=b`],
  ['unknown-remains-unknown', String.raw`\mathtoolsset{other={}}a=b`],
  ['unknown-after-valid', String.raw`\mathtoolsset{pairedDelimiters={},other=true}a=b`],
]) inputs.push([name, tex]);
const cases = [];
for (const [name, tex] of inputs) {
  for (const display of [false, true]) {
    const context = fresh();
    const svg = context.adaptor.innerHTML(context.html.convert(tex, {em: 16, ex: 8, display}));
    cases.push({name: name + (display ? '-display' : '-inline'), tex, display, svg,
      diagnostic: svg.includes('data-mml-node="merror"')});
  }
}
let setOptionsSource;
const observations = [];
for (const [name, raw] of [...forms, ['forbidden-tagforms', 'tagforms={}'], ['forbidden-allow', 'allow-mathtoolsset=false'], ['typo', 'pariedDelimiters={}']]) {
  const context = fresh();
  const jax = context.html.inputJax[0];
  const method = jax.parseOptions.handlers.get('macro').lookup('mathtoolsset').func;
  setOptionsSource = method.toString();
  const options = JSON.parse(JSON.stringify(jax.parseOptions.options.mathtools));
  let error = null;
  try { method({options: {mathtools: options}, GetArgument: () => raw}, '\\mathtoolsset'); }
  catch (e) { error = {id: e.id, message: e.message}; }
  observations.push({name, raw, error, pairedValue: options.pairedDelimiters, pairedType: typeof options.pairedDelimiters});
}
fs.writeFileSync(path.join(__dirname, 'mathtools_paired_option_mathjax_3_2_2.json.gz'), zlib.gzipSync(JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee', assets: hashes,
  scope: 'Full unmodified original SVGs, fresh runtime per input; actual JS delimiter initialization remains a separate API boundary.', cases,
}, null, 2) + '\n'));
fs.writeFileSync(path.join(__dirname, '../internal/tex/testdata/mathtools_paired_option_observations.json'), JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee', assets: hashes,
  setOptionsSource, scope: 'Direct unmodified SetOptions on original default options, fresh runtime per observation. Accepted pairedDelimiters values are stored without reparsing/rebuilding already registered delimiter definitions.',
  cases: observations,
}, null, 2) + '\n');
console.log(JSON.stringify({svgCases: cases.length, valid: cases.filter(c => !c.diagnostic).length,
  diagnostics: cases.filter(c => c.diagnostic).length, directObservations: observations.length}));
