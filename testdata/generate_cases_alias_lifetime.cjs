// SPDX-License-Identifier: Apache-2.0
// Original-only verification: no Go binary or candidate result is read.
const fs = require('node:fs');
const path = require('node:path');
const zlib = require('node:zlib');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const {spawnSync} = require('node:child_process');

const PIN = 'ad8f5c21cb810236551da8c6512ba733e67357ee';
const BASE = '2a19a9626f9136931a5866ca14288d2d64998f0c';
const UNION = 'ce5a75b9828df72f980ae761c8100a40dc0b115984053e0d23f3ab23884317e8';
const NODE = '27db838bb204ef7c21df2931f5656e4c8fb32e6e947f363a402b49714d32b5b1';
const ORACLE = 'ddc9ff73cefa65f0c7caccf9a84426fe14524a0cdabfd833e556cba21767930e';
const ASSETS = {
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'
};
const ARCHIVES = {
  'cases_alias_lifetime_mathjax_3_2_2.json.gz': '00acc38cd1467cc4e8695a3593b774ebfa0bf69360811ef70603abae435fa8aa',
  'cases_alias_lifetime_original_runtime_mathjax_3_2_2.json.gz': '04977ed195a66881b5f98be3c6b9c19e5c4b903f57e9cd0b7dc564a2067e4765'
};
const sha = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const encode = value => JSON.stringify(value, null, 2) + '\n';
// Node readline splits literal U+2028/U+2029. Escape transport only; preserve
// every decoded TeX codepoint, including surrogate pairs, before conversion.
const ascii = value => JSON.stringify(value).replace(/[\u0080-\uffff]/g, c => '\\u' + c.charCodeAt(0).toString(16).padStart(4, '0'));

function main() {
  if (process.argv.length !== 5) throw Error('usage: node --jitless generate_cases_alias_lifetime.cjs PINNED_ASSET_DIR NEW_OUTPUT_DIR ORIGINAL_ORACLE_PATH');
  const assetDir = path.resolve(process.argv[2]), output = path.resolve(process.argv[3]), oraclePath = path.resolve(process.argv[4]);
  assert.equal(fs.existsSync(output), false, 'output must be nonexistent');
  assert.equal(sha(fs.readFileSync(process.execPath)), NODE);
  assert.equal(sha(fs.readFileSync(oraclePath)), ORACLE);
  for (const [name, digest] of Object.entries(ASSETS)) assert.equal(sha(fs.readFileSync(path.join(assetDir, name))), digest);
  const archives = Object.entries(ARCHIVES).map(([name, digest]) => {
    const bytes = fs.readFileSync(path.join(__dirname, name));
    assert.equal(sha(bytes), digest);
    return {name, bytes, fixture: JSON.parse(zlib.gunzipSync(bytes))};
  });
  for (const a of archives) {
    const f = a.fixture;
    assert.equal(f.originalPin, PIN); assert.equal(f.actualMergedBase, BASE);
    assert.equal(f.sourceOriginalUnionSHA256, UNION); assert.equal(f.sourceCaptureCount, 252);
    assert.equal(f.originalAPIOnly, true); assert.equal(f.originalTreeExpectations, false);
    assert.equal(f.originalOracleSHA256, ORACLE); assert.equal(f.originalNodeSHA256, NODE);
    assert.equal(f.originalOraclePath, oraclePath, 'exact original oracle path for runtime stacks');
    assert.deepEqual(f.originalAssets, ASSETS);
    assert.deepEqual(f.options, {Em: 16, Ex: 8, Display: 'explicit per input', FontCache: 0});
  }
  assert.equal(archives[0].fixture.partition, 'strict-svg'); assert.equal(archives[0].fixture.cases.length, 196);
  assert.equal(archives[1].fixture.partition, 'original-runtime-preservation-only'); assert.equal(archives[1].fixture.cases.length, 56);
  const rows = archives.flatMap(a => a.fixture.cases).sort((a, b) => a.sourceUnionIndex - b.sourceUnionIndex);
  const names = new Set(), keys = new Set();
  let normalSVG = 0, diagnosticSVG = 0, copyRuntime = 0, matchRuntime = 0, appendRuntime = 0;
  for (const [index, row] of rows.entries()) {
    assert.equal(row.sourceUnionIndex, index); assert.equal(typeof row.tex, 'string'); assert.equal(typeof row.display, 'boolean');
    assert(row.name && !names.has(row.name)); names.add(row.name);
    const key = JSON.stringify([row.tex, row.display]); assert(!keys.has(key)); keys.add(key);
    assert(row.sourceMembership.family && row.sourceMembership.requestSHA256.length === 64);
    if (row.sourceKind === 'original-runtime') {
      assert.deepEqual(Object.keys(row.original), ['error']); assert.equal(sha(row.original.error), row.originalErrorHash);
      if (row.original.error.startsWith("TypeError: Cannot read properties of undefined (reading 'copy')\n")) copyRuntime++;
      else if (row.original.error.startsWith("TypeError: Cannot read properties of undefined (reading 'match')\n")) matchRuntime++;
      else { assert(row.original.error.startsWith("TypeError: Cannot read properties of undefined (reading 'appendChild')\n")); appendRuntime++; }
    } else {
      assert.deepEqual(Object.keys(row.original), ['svg']); assert.equal(sha(row.original.svg), row.originalSVGHash);
      if (row.original.svg.includes('data-mjx-error=')) { assert.equal(row.sourceKind, 'diagnostic-svg'); diagnosticSVG++; }
      else { assert.equal(row.sourceKind, 'normal-svg'); normalSVG++; }
    }
  }
  assert.deepEqual({normalSVG, diagnosticSVG, copyRuntime, matchRuntime, appendRuntime}, {normalSVG: 32, diagnosticSVG: 164, copyRuntime: 44, matchRuntime: 4, appendRuntime: 8});
  const requests = rows.map(r => ({tex: r.tex, options: {Em: 16, Ex: 8, Display: r.display, FontCache: 0}}));
  const input = requests.map(r => ascii(r) + '\n').join('');
  assert.equal(Buffer.from(input).every(c => c < 128), true);
  assert.deepEqual(input.slice(0, -1).split('\n').map(s => JSON.parse(s)), requests);
  fs.mkdirSync(output);
  fs.writeFileSync(path.join(output, 'requests.jsonl'), input, {flag: 'wx'});
  const argv = ['--jitless', oraclePath, '--asset-dir', assetDir];
  const envOverrides = {NODE_OPTIONS: '--jitless', NODE_NO_WARNINGS: '1'};
  fs.writeFileSync(path.join(output, 'started.json'), encode({status: 'RUNNING_ORIGINAL_ONLY', argv: [process.execPath, ...argv], envOverrides, originalPin: PIN, actualBase: BASE, assets: ASSETS, nodeSHA256: NODE, oracleSHA256: ORACLE, inputSHA256: sha(input), freshVMPerRequest: true, sourceCount: 252}), {flag: 'wx'});
  let conversions = 0;
  try {
    const child = spawnSync(process.execPath, argv, {input, encoding: null, env: {...process.env, ...envOverrides}, timeout: 180000, maxBuffer: 64 * 1024 * 1024});
    fs.writeFileSync(path.join(output, 'original-stdout.bin'), child.stdout || Buffer.alloc(0), {flag: 'wx'});
    fs.writeFileSync(path.join(output, 'original-stderr.bin'), child.stderr || Buffer.alloc(0), {flag: 'wx'});
    if (child.error) throw child.error;
    assert.equal(child.status, 0); assert.equal(child.signal, null);
    assert.equal(child.stderr.length, 0);
    const raw = child.stdout.toString('utf8'); assert(raw.endsWith('\n'));
    const lines = raw.slice(0, -1).split('\n'); assert.equal(lines.length, 252); assert(lines.every(Boolean));
    const results = lines.map(line => JSON.parse(line));
    const observations = path.join(output, 'fresh-observations.jsonl');
    fs.writeFileSync(observations, '', {flag: 'wx'});
    for (const [index, result] of results.entries()) {
      fs.appendFileSync(observations, JSON.stringify({profile: 'public-default', sourceUnionIndex: index, request: requests[index], result}) + '\n'); conversions++;
    }
    for (const [index, result] of results.entries()) assert.deepEqual(result, rows[index].original, 'whole original API at sourceUnionIndex ' + index);
    for (const a of archives) fs.writeFileSync(path.join(output, a.name), a.bytes, {flag: 'wx'});
    const receipt = {status: 'PASS_ORIGINAL_ONLY', originalPin: PIN, actualMergedBase: BASE, sourceOriginalUnionSHA256: UNION, conversions, originalExitCode: child.status, freshVMPerRequest: true, wholeOriginalAPIEquality: true, originalKinds: {normalSVG, diagnosticSVG, runtime: copyRuntime + matchRuntime + appendRuntime, copyRuntime, matchRuntime, appendRuntime}, byteIdenticalArchives: ARCHIVES, fullRuntimeErrorsUnnormalized: true, originalTreeExpectations: false, GoExecuted: false, candidateOutcomeClaim: false, files: Object.fromEntries(['requests.jsonl', 'started.json', 'original-stdout.bin', 'original-stderr.bin', 'fresh-observations.jsonl'].map(name => [name, sha(fs.readFileSync(path.join(output, name)))]))};
    fs.writeFileSync(path.join(output, 'regeneration.json'), encode(receipt), {flag: 'wx'});
    process.stdout.write(JSON.stringify(receipt) + '\n');
  } catch (error) {
    fs.writeFileSync(path.join(output, 'failure.json'), encode({status: 'FAILED_ORIGINAL_ONLY', conversions, error: String(error.stack || error), rawOriginalFilesPreserved: true}), {flag: 'wx'});
    throw error;
  }
}
try { main(); } catch (error) { process.stderr.write(String(error.stack || error) + '\n'); process.exitCode = 1; }
