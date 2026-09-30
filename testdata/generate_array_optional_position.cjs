// SPDX-License-Identifier: Apache-2.0
// Original-only draft. Run only after the execution plan is reviewed.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const zlib = require('node:zlib');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');

const PIN = 'ad8f5c21cb810236551da8c6512ba733e67357ee';
const BASE = '16801f1f0e8e5c0b1582224cce85140b1093ed38';
const FILE = 'array_optional_position_mathjax_3_2_2.json.gz';
const FILE_SHA256 = '3a3360481e88ab9d7afa4da9a1fa3609e08c6296a5ed224d5bb5cbb79e0d14e4';
const NODE_SHA256 = '27db838bb204ef7c21df2931f5656e4c8fb32e6e947f363a402b49714d32b5b1';
const ASSETS = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'
};
const sha = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const clone = value => JSON.parse(JSON.stringify(value));
const encode = value => JSON.stringify(value, null, 2) + '\n';

async function main() {
  if (process.argv.length !== 4) throw new Error('usage: node generate_array_optional_position.cjs PINNED_ASSET_DIR NEW_OUTPUT_DIR');
  const assetDir = path.resolve(process.argv[2]);
  const output = path.resolve(process.argv[3]);
  assert.equal(fs.existsSync(output), false, 'output must be a fresh directory');
  assert.equal(sha(fs.readFileSync(process.execPath)), NODE_SHA256, 'pinned Node binary');
  const scripts = Object.entries(ASSETS).map(([name, digest]) => {
    const bytes = fs.readFileSync(path.join(assetDir, name));
    assert.equal(sha(bytes), digest, 'pinned original asset ' + name);
    return new vm.Script(bytes.toString('utf8'), {filename: name});
  });
  const archive = fs.readFileSync(path.join(__dirname, FILE));
  assert.equal(sha(archive), FILE_SHA256, 'untouched full original archive');
  const fixture = JSON.parse(zlib.gunzipSync(archive));
  assert.equal(fixture.originalPin, PIN);
  assert.equal(fixture.actualMergedBase, BASE);
  assert.equal(fixture.cases.length, 106);
  const seen = new Set(), names = new Set();
  let valid = 0, errors = 0, retained = 0, fresh = 0;
  for (const [index, row] of fixture.cases.entries()) {
    assert.equal(row.publicIndex, index);
    assert.equal(typeof row.tex, 'string'); assert.equal(typeof row.display, 'boolean');
    assert(row.name && !names.has(row.name)); names.add(row.name);
    const key = JSON.stringify([row.tex, row.display]); assert(!seen.has(key)); seen.add(key);
    assert.deepEqual(Object.keys(row.original), ['svg']);
    assert.equal(row.sourceXMLValid, true);
    assert.equal(sha(row.original.svg), row.originalSVGHash);
    assert.equal(row.inputMetadata.tex, row.tex); assert.equal(row.inputMetadata.display, row.display);
    if (row.original.svg.includes('data-mjx-error=')) { assert.equal(row.sourceKind, 'error-svg'); errors++; }
    else { assert.equal(row.sourceKind, 'svg'); valid++; }
    if (row.origin === 'exact-retained-original-reuse') {
      const r = row.provenance.completeReuseRecord;
      assert.equal(r.tex, row.tex); assert.equal(r.display, row.display);
      assert.deepEqual(r.original, row.original);
      assert.deepEqual(r.completeCurrentPublishedResidual.original, row.original);
      assert.equal(r.completeCurrentPublishedResidual.partition, 'held-svg');
      assert.equal(r.completeCurrentPublishedResidual.reason, 'AlignedArray optional setup (explicitly separate)');
      assert.equal(r.currentSource.commit, BASE);
      assert.equal(r.currentSource.path, 'testdata/ordinary_array_owner_residuals.json.gz');
      retained++;
    } else { assert.equal(row.origin, 'fresh-frozen-default-original'); fresh++; }
  }
  assert.deepEqual({valid, errors, retained, fresh}, {valid: 96, errors: 10, retained: 48, fresh: 58});
  fs.mkdirSync(output);
  const stream = path.join(output, 'fresh-observations.jsonl');
  fs.writeFileSync(stream, '', {flag: 'wx'});
  let conversions = 0, peakHeapUsed = 0, peakRSS = 0;
  try {
    for (const row of fixture.cases) {
      const request = {tex: row.tex, display: row.display, options: {Em: 16, Ex: 8, Display: row.display, FontCache: 0}};
      let context = vm.createContext({console}); context.globalThis = context;
      for (const script of scripts) script.runInContext(context, {timeout: 10000});
      let result;
      try {
        result = {svg: context.adaptor.innerHTML(context.html.convert(row.tex, {display: row.display, em: 16, ex: 8}))};
      } catch (error) { result = {error: String(error.stack || error)}; }
      result = clone(result);
      // Preserve the complete detached API object before any assertion/yield.
      fs.appendFileSync(stream, JSON.stringify({profile: 'public-default', publicIndex: row.publicIndex, request, result}) + '\n');
      conversions++;
      const usage = process.memoryUsage(); peakHeapUsed = Math.max(peakHeapUsed, usage.heapUsed); peakRSS = Math.max(peakRSS, usage.rss);
      context = null;
      await new Promise(resolve => setImmediate(resolve));
      assert.deepEqual(result, row.original, 'complete original API at ' + row.publicIndex);
    }
    assert.equal(conversions, 106);
    // Keep all historical fields, complete published residuals, and provenance
    // byte-identical after every fresh original API object has been compared.
    fs.writeFileSync(path.join(output, FILE), archive, {flag: 'wx'});
    assert.equal(sha(fs.readFileSync(path.join(output, FILE))), FILE_SHA256);
    const receipt = {status: 'PASS_ORIGINAL_ONLY', originalPin: PIN, actualMergedBase: BASE, assets: ASSETS, nodeSHA256: NODE_SHA256, conversions, fullStreamRecords: conversions, originalKinds: {svg: valid, errorSVG: errors, runtime: 0, invalidXML: 0}, historicalOrigins: {retained, fresh}, byteIdenticalContainers: {[FILE]: FILE_SHA256}, fixtureHistoricalMetadataUnchanged: true, candidateOutcomeClaim: false, publicationNewnessClaim: false, peakHeapUsed, peakRSS};
    fs.writeFileSync(path.join(output, 'regeneration.json'), encode(receipt), {flag: 'wx'});
    process.stdout.write(JSON.stringify(receipt) + '\n');
  } catch (error) {
    fs.writeFileSync(path.join(output, 'failure.json'), encode({status: 'FAILED', conversions, error: String(error.stack || error), peakHeapUsed, peakRSS}), {flag: 'wx'});
    throw error;
  }
}
main().catch(error => { process.stderr.write(String(error.stack || error) + '\n'); process.exitCode = 1; });
