// SPDX-License-Identifier: Apache-2.0
// Observe the original exported helper with error=false and error=true.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node --jitless generate_filtered_keyval.cjs PINNED_ASSETS');
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
// ENCLOSE_OPTIONS in the original EncloseConfiguration.ts, also used by Cancel.
const allowedKeys = ['data-arrowhead', 'color', 'mathcolor', 'background', 'mathbackground', 'data-padding', 'data-thickness'];
const rawInputs = [
  ['empty', ''], ['unknown', 'unknown=x'], ['bare-unknown', 'unknown'],
  ['unknown-first-color', 'unknown=x,mathcolor=red'], ['unknown-last-color', 'mathcolor=red,unknown=x'],
  ['proto', '__proto__=true'], ['proto-color', '__proto__=true,mathcolor=red'],
  ['braced-proto', '{{__proto__}}=true,mathcolor=red'], ['bom-proto', '\uFEFF__proto__\uFEFF=true'],
  ['constructor', 'constructor=x'], ['toString', 'toString=x'], ['hasOwnProperty', 'hasOwnProperty=x'],
  ['numeric-sort', '10=x,2=y,0=z,mathcolor=red'], ['numeric-boundary', '4294967295=x,4294967294=y'],
  ['noncanonical-numeric', '01=x,+2=y,-0=z,mathcolor=red'],
  ['empty-key', '=red,mathcolor=red'], ['empty-entries', ',,,mathcolor=red,,'],
  ['nested-comma', 'unknown={{a,b=c}},mathcolor={red}'], ['escaped-comma', String.raw`unknown={a\,b},mathcolor=red`],
  ['bool-colors', 'mathcolor=false,color=true,mathbackground=false,background'],
  ['boolean-strings', 'mathcolor={false},mathbackground={{true}}'],
  ['bom-color', 'mathcolor=\uFEFFred\uFEFF'], ['nel-color', 'mathcolor=\u0085red\u0085'],
  ['blank-color', 'mathcolor='], ['empty-braces-color', 'mathcolor={}'],
  ['duplicates', 'unknown=x,mathcolor=blue,mathcolor=red,unknown=y'],
  ['malformed-open', 'unknown={'], ['malformed-close', 'unknown=x}'],
  ['invalid-before-malformed', 'unknown=x,mathcolor={'], ['valid-before-malformed', 'mathcolor=red,unknown={'],
];
const cases = [];
let keyvalMethod;
for (const [name, raw] of rawInputs) {
  for (const [mode, allowed, errorOnUnknown] of [['unvalidated', false, false], ['filtered', true, false], ['strict', true, true]]) {
    const context = fresh();
    const method = context.MathJax._.input.tex.ParseUtil.default.keyvalOptions;
    keyvalMethod = method.toString();
    let options = null, keys = null, error = null;
    try {
      options = method(raw, allowed ? Object.fromEntries(allowedKeys.map(key => [key, 1])) : null, errorOnUnknown);
      keys = Object.keys(options);
    } catch (e) { error = {id: e.id, message: e.message}; }
    cases.push({name: `${name}-${mode}`, raw, allowed, errorOnUnknown, options, keys, error});
  }
}
fs.writeFileSync(path.join(__dirname, '../internal/tex/testdata/filtered_keyval_mathjax_3_2_2.json'), JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee', assets: hashes, allowedKeys, keyvalMethod,
  scope: 'Direct complete original keyvalOptions results with no allowlist, filtered Cancel allowlist, and strict allowlist; fresh frozen runtime per call.', cases,
}, null, 2) + '\n');
console.log(JSON.stringify({cases: cases.length, successful: cases.filter(c => c.error === null).length,
  errors: cases.filter(c => c.error !== null).length}));
