// SPDX-License-Identifier: Apache-2.0
// Regenerate only original observations. Historical Go receipts stay untouched.
// Usage: node --jitless testdata/generate_operator_primes.cjs PINNED_ASSETS
const fs = require('node:fs');
const path = require('node:path');
const {spawnSync} = require('node:child_process');
const assets = process.argv[2];
if (!assets) throw new Error('expected the pinned D2 MathJax asset directory');
function run(batch, observe = false, inputKey = 'tex') {
  // Keep Unicode line separators inside the JSON string on the JSONL wire.
  const input = batch.map(c => JSON.stringify({tex: c[inputKey], display: c.display, options: {Display: c.display}})
    .replace(/[\u007f-\uffff]/g, c => '\\u' + c.charCodeAt(0).toString(16).padStart(4, '0'))).join('\n') + '\n';
  const script = observe ? 'operator_primes_observer.mjs' : 'differential/oracle.mjs';
  const args = observe ? [assets] : ['--asset-dir', assets];
  const processResult = spawnSync(process.execPath, ['--jitless', path.join(__dirname, script), ...args],
    {input, encoding: 'utf8', maxBuffer: 16 * 1024 * 1024});
  if (processResult.status !== 0) throw new Error(processResult.stderr || String(processResult.error));
  const results = processResult.stdout.trim().split('\n').map(line => JSON.parse(line));
  if (results.length !== batch.length) throw new Error('incomplete original output');
  return results;
}
const exactFile = path.join(__dirname, 'operator_primes_mathjax_3_2_2.json');
const exact = JSON.parse(fs.readFileSync(exactFile, 'utf8'));
for (let start = 0; start < exact.cases.length; start += 24) {
  const batch = exact.cases.slice(start, start + 24);
  run(batch, true).forEach((original, index) => {
    const record = batch[index];
    record.svg = original.svg;
    if (Object.hasOwn(record, 'operators')) record.operators = original.operators;
  });
}
fs.writeFileSync(exactFile, JSON.stringify(exact, null, 2) + '\n');
const residualFile = path.join(__dirname, 'operator_primes_residuals.json');
const residuals = JSON.parse(fs.readFileSync(residualFile, 'utf8'));
for (let start = 0; start < residuals.cases.length; start += 24) {
  const batch = residuals.cases.slice(start, start + 24);
  run(batch).forEach((original, index) => { batch[index].original = original; });
}
const literals = residuals.cases.filter(record => Object.hasOwn(record, 'literalTeX'));
for (let start = 0; start < literals.length; start += 24) {
  const batch = literals.slice(start, start + 24);
  run(batch, false, 'literalTeX').forEach((original, index) => { batch[index].originalLiteral = original; });
}
fs.writeFileSync(residualFile, JSON.stringify(residuals, null, 2) + '\n');
