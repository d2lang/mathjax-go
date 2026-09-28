// SPDX-License-Identifier: Apache-2.0
// Usage: node --jitless testdata/generate_numcases_dynamic_helper.cjs PINNED_ASSETS
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const {spawnSync} = require('node:child_process');
const assets = process.argv[2];
if (!assets) throw Error('expected pinned D2 asset directory');
const hashes = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
for (const [file, expected] of Object.entries(hashes)) {
  const hash = crypto.createHash('sha256').update(fs.readFileSync(path.join(assets,file))).digest('hex');
  if (hash !== expected) throw Error(`unverified ${file}`);
}
for (const name of (process.argv[3] ? [process.argv[3]] : ['numcases_dynamic_helper_mathjax_3_2_2.json','numcases_dynamic_helper_residuals.json'])) {
  const file = path.join(__dirname,name);
  const fixture = JSON.parse(fs.readFileSync(file,'utf8'));
  const requests = [];
  for (const c of fixture.cases) {
    requests.push({c, tex:c.tex, key:name.includes('residuals') ? 'original' : 'svg'});
    if (c.canonicalControl) requests.push({c:c.canonicalControl, tex:c.canonicalControl.tex, key:'original'});
  }
  for (let start=0; start<requests.length; start+=24) {
    const batch=requests.slice(start,start+24);
    const input=batch.map(({c,tex})=>JSON.stringify({tex,options:{Display:c.display}})
      .replace(/[\u0080-\uffff]/g,c=>'\\u'+c.charCodeAt(0).toString(16).padStart(4,'0'))).join('\n')+'\n';
    const run=spawnSync(process.execPath,['--jitless',path.join(__dirname,'differential/oracle.mjs'),'--asset-dir',assets],
      {input,encoding:'utf8',maxBuffer:16*1024*1024});
    if (run.status!==0) throw Error(run.stderr || String(run.error));
    const results=run.stdout.replace(/\n$/, '').split('\n').map(line=>JSON.parse(line));
    if (results.length!==batch.length) throw Error('incomplete original output');
    results.forEach((result,i)=> {
      const {c,key}=batch[i];
      if (key==='svg') {
        if (typeof result.svg!=='string' || result.error) throw Error(JSON.stringify(result));
        c.svg=result.svg;
      } else c[key]=result;
    });
  }
  fs.writeFileSync(file,JSON.stringify(fixture,null,2)+'\n');
}

// These metadata observations are separate from the complete-SVG assertions.
if (!process.argv[3]) {
  const file = path.join(__dirname, 'numcases_child_environment_proof.json');
  const fixture = JSON.parse(fs.readFileSync(file, 'utf8'));
  const input = fixture.cases.map(c => JSON.stringify({tex:c.tex, display:c.display})
    .replace(/[\u0080-\uffff]/g,c=>'\\u'+c.charCodeAt(0).toString(16).padStart(4,'0'))).join('\n')+'\n';
  const run = spawnSync(process.execPath, ['--jitless', path.join(__dirname,'observe_numcases_environment.cjs'), assets],
    {input,encoding:'utf8',maxBuffer:32*1024*1024});
  if (run.status !== 0) throw Error(run.stderr || String(run.error));
  const results=run.stdout.replace(/\n$/, '').split('\n').map(line=>JSON.parse(line));
  if (results.length !== fixture.cases.length) throw Error('incomplete original environment observations');
  const originals = new Map(JSON.parse(fs.readFileSync(path.join(__dirname,'numcases_dynamic_helper_mathjax_3_2_2.json'),'utf8')).cases.map(c=>[JSON.stringify([c.tex,c.display]),c.svg]));
  results.forEach((result, i) => {
    if (result.original.svg !== originals.get(JSON.stringify([fixture.cases[i].tex,fixture.cases[i].display]))) throw Error('environment observer changed original SVG');
    if (typeof result.original.svg !== 'string' || result.original.error) throw Error(JSON.stringify(result.original));
    fixture.cases[i].observed = result.observed;
    fixture.cases[i].originalSVG_SHA256 = crypto.createHash('sha256').update(result.original.svg).digest('hex');
  });
  fs.writeFileSync(file,JSON.stringify(fixture,null,2)+'\n');
}
