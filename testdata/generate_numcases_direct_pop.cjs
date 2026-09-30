// SPDX-License-Identifier: Apache-2.0
// node --jitless generate_numcases_direct_pop.cjs PINNED_ASSET_DIR NEW_OUTPUT_DIR ORIGINAL_ORACLE_PATH
// Original-only regeneration. No Go binary or candidate output is read.
const fs = require('node:fs');
const path = require('node:path');
const zlib = require('node:zlib');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const {spawnSync} = require('node:child_process');

const PIN = 'ad8f5c21cb810236551da8c6512ba733e67357ee';
const BASE = 'c24058161bb5df6af3f269ce65030c94079b952a';
const UNION = 'a4fda78dd875df1e1ffbdd452a48dd98d2602c3f5ac865d4bb39ec964c8650e3';
const INITIAL_UNION = 'e09328fc33b63a604c95c93a9b29db3d8fe72854e425ed11a8d83ea06b1d5655';
const NODE = '27db838bb204ef7c21df2931f5656e4c8fb32e6e947f363a402b49714d32b5b1';
const ORACLE = 'ddc9ff73cefa65f0c7caccf9a84426fe14524a0cdabfd833e556cba21767930e';
const ASSETS = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'
};
const ARCHIVES = {
  'numcases_direct_pop_mathjax_3_2_2.json.gz': '83dc943dffb364b534b363dd08ef6bde5452cdab6f953ae8c7ec25a1c59dffae',
  'numcases_direct_pop_original_runtime_mathjax_3_2_2.json.gz': 'ff61cca875471472028b067b9c854efb6780fc72365c83f4f1ce1fc110a1f195'
};
const sha = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const encode = value => JSON.stringify(value, null, 2) + '\n';

function main() {
  if (process.argv.length !== 5) throw Error('usage: node --jitless generate_numcases_direct_pop.cjs PINNED_ASSET_DIR NEW_OUTPUT_DIR ORIGINAL_ORACLE_PATH');
  const assetDir = path.resolve(process.argv[2]), output = path.resolve(process.argv[3]), oraclePath = path.resolve(process.argv[4]);
  assert.equal(fs.existsSync(output), false, 'output must be a nonexistent directory');
  assert.equal(sha(fs.readFileSync(process.execPath)), NODE, 'pinned Node binary');
  assert.equal(sha(fs.readFileSync(oraclePath)), ORACLE, 'unchanged original oracle script');
  for (const [name, digest] of Object.entries(ASSETS)) assert.equal(sha(fs.readFileSync(path.join(assetDir, name))), digest, 'frozen original asset ' + name);

  const archives = Object.entries(ARCHIVES).map(([name, digest]) => {
    const bytes = fs.readFileSync(path.join(__dirname, name));
    assert.equal(sha(bytes), digest, 'untouched complete original archive ' + name);
    return {name, bytes, fixture: JSON.parse(zlib.gunzipSync(bytes))};
  });
  const [svgFixture, runtimeFixture] = archives.map(a => a.fixture);
  for (const fixture of [svgFixture, runtimeFixture]) {
    assert.equal(fixture.originalPin, PIN); assert.equal(fixture.actualMergedBase, BASE);
    const initialRuntimeArchive = fixture === runtimeFixture;
    assert.equal(fixture.sourceOriginalUnionSHA256, initialRuntimeArchive ? INITIAL_UNION : UNION);
    assert.equal(fixture.sourceCaptureCount, initialRuntimeArchive ? 56 : 112);
    assert.equal(fixture.originalAPIOnly, true); assert.equal(fixture.originalTreeExpectations, false);
    assert.equal(fixture.originalOracleSHA256, ORACLE); assert.equal(fixture.originalNodeSHA256, NODE);
    assert.equal(oraclePath, fixture.originalOraclePath, 'preserve exact captured oracle path for complete runtime stacks');
    assert.deepEqual(fixture.originalAssets, ASSETS);
    assert.deepEqual(fixture.options, {Em: 16, Ex: 8, Display: 'explicit per input', FontCache: 0});
  }
  assert.equal(svgFixture.partition, 'strict-svg'); assert.equal(svgFixture.cases.length, 100);
  assert.equal(runtimeFixture.partition, 'original-runtime-preservation-only'); assert.equal(runtimeFixture.cases.length, 12);
  const rows = [...svgFixture.cases, ...runtimeFixture.cases].sort((a, b) => a.sourceUnionIndex - b.sourceUnionIndex);
  const names = new Set(), keys = new Set();
  let svg = 0, errorSVG = 0, retained = 0, freshSVG = 0, copyRuntime = 0, spreadRuntime = 0;
  for (const [index, row] of rows.entries()) {
    assert.equal(row.sourceUnionIndex, index); assert.equal(typeof row.tex, 'string'); assert.equal(typeof row.display, 'boolean');
    assert(row.name && !names.has(row.name)); names.add(row.name);
    const key = JSON.stringify([row.tex, row.display]); assert(!keys.has(key)); keys.add(key);
    assert.deepEqual(row.inputMetadata, {name: row.name, tex: row.tex, display: row.display});
    if (row.sourceKind === 'runtime') {
      assert.deepEqual(Object.keys(row.original), ['error']); assert.equal(row.sourceXMLValid, false);
      assert.equal(sha(row.original.error), row.originalErrorHash);
      assert(row.original.error.startsWith(row.originalFirstLine + '\n'));
      assert.equal(row.origin, 'fresh-frozen-default-original');
      assert.deepEqual(row.publicationProvenance, {category: 'first-runtime-input', firstInput: true});
      if (row.originalFirstLine === "TypeError: Cannot read properties of undefined (reading 'copy')") copyRuntime++;
      else { assert.equal(row.originalFirstLine, "TypeError: Cannot read properties of undefined (reading 'match')"); spreadRuntime++; }
    } else {
      assert.deepEqual(Object.keys(row.original), ['svg']); assert.equal(row.sourceXMLValid, true);
      assert.equal(sha(row.original.svg), row.originalSVGHash);
      if (row.original.svg.includes('data-mjx-error=')) { assert.equal(row.sourceKind, 'error-svg'); errorSVG++; }
      else { assert.equal(row.sourceKind, 'svg'); svg++; }
      if (row.origin === 'exact-retained-original-reuse') {
        assert.deepEqual(row.publicationProvenance, {category: 'prior-held-svg-assertion-promotion', firstInput: false});
        const p = row.provenance, held = p.completeCurrentPublishedResidual;
        assert.equal(p.currentSource.commit, BASE); assert.equal(p.currentSource.path, 'testdata/ordinary_array_owner_residuals.json.gz');
        assert.equal(p.currentSource.sha256, '7877905982204984a0a19b69b3f6129d0df23a4d97524d8d77dbe9f5323dd209');
        assert.equal(held.name, row.name); assert.equal(held.tex, row.tex); assert.equal(held.display, row.display);
        assert.equal(held.partition, 'held-svg'); assert.equal(held.reason, 'Existing NumCases/SubNumCases direct Pop finalization');
        assert.deepEqual(held.original, row.original); retained++;
      } else {
        assert.equal(row.origin, 'fresh-frozen-default-original');
        assert.deepEqual(row.publicationProvenance, {category: 'first-svg-input', firstInput: true});
        assert(row.provenance.newnessReceipt && !row.provenance.completeCurrentPublishedResidual); freshSVG++;
      }
    }
  }
  assert.deepEqual({svg, errorSVG, retained, freshSVG, copyRuntime, spreadRuntime}, {svg: 48, errorSVG: 52, retained: 8, freshSVG: 92, copyRuntime: 4, spreadRuntime: 8});

  fs.mkdirSync(output);
  const requests = rows.map(row => ({tex: row.tex, options: {Em: 16, Ex: 8, Display: row.display, FontCache: 0}}));
  const input = requests.map(request => JSON.stringify(request) + '\n').join('');
  fs.writeFileSync(path.join(output, 'requests.jsonl'), input, {flag: 'wx'});
  const argv = [oraclePath, '--asset-dir', assetDir];
  const envOverrides = {NODE_OPTIONS: '--jitless', NODE_NO_WARNINGS: '1'};
  fs.writeFileSync(path.join(output, 'started.json'), encode({status: 'RUNNING_ORIGINAL_ONLY', argv: [process.execPath, ...argv], envOverrides, originalPin: PIN, actualBase: BASE, assets: ASSETS, nodeSHA256: NODE, oracleSHA256: ORACLE, inputSHA256: sha(input), freshVMPerRequest: true, sourceCount: 112}), {flag: 'wx'});
  let conversions = 0;
  try {
    const child = spawnSync(process.execPath, argv, {input, encoding: null, env: {...process.env, ...envOverrides}, timeout: 180000, maxBuffer: 16 * 1024 * 1024});
    // Save every original byte before validating status, framing, or equality.
    fs.writeFileSync(path.join(output, 'original-stdout.bin'), child.stdout || Buffer.alloc(0), {flag: 'wx'});
    fs.writeFileSync(path.join(output, 'original-stderr.bin'), child.stderr || Buffer.alloc(0), {flag: 'wx'});
    if (child.error) throw child.error;
    assert.equal(child.status, 0, 'original process completed'); assert.equal(child.signal, null);
    const raw = child.stdout.toString('utf8'); assert(raw.endsWith('\n'), 'complete LF-framed original stdout');
    const lines = raw.slice(0, -1).split('\n'); assert.equal(lines.length, 112, 'complete original API record count');
    const stream = path.join(output, 'fresh-observations.jsonl'); fs.writeFileSync(stream, '', {flag: 'wx'});
    const results = lines.map(line => JSON.parse(line));
    for (const [index, result] of results.entries()) {
      // Record all detached API objects before the first parity assertion.
      fs.appendFileSync(stream, JSON.stringify({profile: 'public-default', sourceUnionIndex: index, request: requests[index], result}) + '\n'); conversions++;
    }
    for (const [index, result] of results.entries()) assert.deepEqual(result, rows[index].original, 'complete original API at sourceUnionIndex ' + index);
    for (const archive of archives) {
      fs.writeFileSync(path.join(output, archive.name), archive.bytes, {flag: 'wx'});
      assert.equal(sha(fs.readFileSync(path.join(output, archive.name))), ARCHIVES[archive.name]);
    }
    const receipt = {status: 'PASS_ORIGINAL_ONLY', originalPin: PIN, actualMergedBase: BASE, sourceOriginalUnionSHA256: UNION, conversions, originalExitCode: child.status, freshVMPerRequest: true, wholeOriginalAPIEquality: true, originalKinds: {svg, errorSVG, runtime: copyRuntime + spreadRuntime, invalidXML: 0}, historicalOrigins: {priorHeldPromotions: retained, newSVGInputs: freshSVG, newRuntimeInputs: copyRuntime + spreadRuntime}, byteIdenticalArchives: ARCHIVES, completeRuntimeErrorsUnnormalized: true, originalTreeExpectations: false, GoExecuted: false, candidateOutcomeClaim: false, files: Object.fromEntries(['requests.jsonl', 'started.json', 'original-stdout.bin', 'original-stderr.bin', 'fresh-observations.jsonl'].map(name => [name, sha(fs.readFileSync(path.join(output, name)))]))};
    fs.writeFileSync(path.join(output, 'regeneration.json'), encode(receipt), {flag: 'wx'});
    process.stdout.write(JSON.stringify(receipt) + '\n');
  } catch (error) {
    fs.writeFileSync(path.join(output, 'failure.json'), encode({status: 'FAILED_ORIGINAL_ONLY', conversions, error: String(error.stack || error), rawOriginalFilesPreserved: true}), {flag: 'wx'});
    throw error;
  }
}
try { main(); } catch (error) { process.stderr.write(String(error.stack || error) + '\n'); process.exitCode = 1; }
