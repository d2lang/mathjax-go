// SPDX-License-Identifier: Apache-2.0
// node --jitless testdata/generate_getstar_bra.cjs PINNED_ASSETS SCRATCH_OUTPUT
// Reads explicit preserved inputs and uses only the hash-verified original.
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const zlib = require('node:zlib'), {spawnSync} = require('node:child_process');
const assets = process.argv[2], output = process.argv[3];
if (!assets || !output) throw Error('usage: generate_getstar_bra.cjs PINNED_ASSETS SCRATCH_OUTPUT');
if (path.resolve(output) === path.resolve(__dirname)) throw Error('use a separate scratch output directory');
fs.mkdirSync(output, {recursive:true});
const sha = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const pin = 'ad8f5c21cb810236551da8c6512ba733e67357ee';
const firstLine = "TypeError: Cannot read properties of null (reading '4')";
const files = ['getstar_bra_mathjax_3_2_2.json.gz', 'getstar_bra_residuals.json.gz'];
const seen = new Set(), reports = [], runtimeObservations = [];
for (const name of files) {
  const inputBytes = fs.readFileSync(path.join(__dirname, name));
  const data = JSON.parse(zlib.gunzipSync(inputBytes));
  if (data.mathjaxGitCommit !== pin) throw Error('unbound original source');
  for (const [asset, hash] of Object.entries(data.originalAssets)) {
    if (sha(fs.readFileSync(path.join(assets, asset))) !== hash) throw Error('unverified ' + asset);
  }
  for (const c of data.cases) {
    const key = JSON.stringify([c.tex,c.display]);
    if (seen.has(key) || typeof c.display !== 'boolean') throw Error('duplicate or ambiguous input');
    seen.add(key);
  }
  let svg = 0, runtime = 0;
  for (let start = 0; start < data.cases.length; start += 24) {
    const cases = data.cases.slice(start,start+24);
    const input = cases.map(c => JSON.stringify({tex:c.tex,display:c.display,options:{display:c.display,Display:c.display}})
      .replace(/\u2028/g,'\\u2028').replace(/\u2029/g,'\\u2029')).join('\n')+'\n';
    const run = spawnSync(process.execPath,['--jitless',path.join(__dirname,'differential/oracle.mjs'),'--asset-dir',assets],
      {input,encoding:'utf8',maxBuffer:64*1024*1024});
    if (run.status !== 0) throw Error('original process failed: '+run.stderr);
    const rows = run.stdout.replace(/\n$/,'').split('\n').map(JSON.parse);
    if (rows.length !== cases.length) throw Error('incomplete original capture');
    rows.forEach((fresh,i) => {
      const c = cases[i], original = c.original;
      if (typeof original.svg === 'string') {
        if (fresh.error || fresh.svg !== original.svg || Object.keys(fresh).join() !== 'svg') {
          throw Error('complete original SVG changed: '+c.name);
        }
        // The replacement is freshly captured original output, never a Go result.
        c.original = fresh;
        svg++;
      } else {
        if (c.partition !== 'runtime-bounded-nel' || !c.tex.includes('\u0085') || original.svg || fresh.svg ||
            typeof original.error !== 'string' || typeof fresh.error !== 'string' ||
            original.error.split('\n')[0] !== firstLine || fresh.error.split('\n')[0] !== firstLine) {
          throw Error('original runtime contract changed: '+c.name);
        }
        // Stack locations depend on the capture runner. Preserve the complete
        // historical exception and bind the fresh complete exception separately.
        runtimeObservations.push({name:c.name,tex:c.tex,display:c.display,
          preservedErrorSHA256:sha(original.error),fresh});
        runtime++;
      }
    });
  }
  const bytes = zlib.gzipSync(JSON.stringify(data,null,2)+'\n',{level:9});
  fs.writeFileSync(path.join(output,name),bytes);
  reports.push({file:name,cases:data.cases.length,svg,runtime,inputSHA256:sha(inputBytes),
    regeneratedSHA256:sha(bytes),byteIdentical:bytes.equals(inputBytes)});
  console.log(JSON.stringify(reports[reports.length-1]));
}
if (seen.size !== 5170 || runtimeObservations.length !== 682) throw Error('unbound union/runtime count');
const receipt = {mathjaxGitCommit:pin,uniqueInputs:seen.size,source:'hash-verified frozen original only',
  completeSVGComparison:'exact unmodified bytes, including retained raw invalid-XML originals',
  runtimePolicy:'historical complete objects unchanged; fresh complete original exceptions retained below',
  files:reports,runtimeObservations};
fs.writeFileSync(path.join(output,'getstar_bra_regeneration.json.gz'),zlib.gzipSync(JSON.stringify(receipt,null,2)+'\n',{level:9}));
if (reports.some(r => !r.byteIdentical)) throw Error('fixture serialization changed');
