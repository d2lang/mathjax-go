// SPDX-License-Identifier: Apache-2.0
// Usage: node --jitless testdata/generate_runtime_environment_end.cjs PINNED_ASSETS
// Only the frozen JavaScript original produces reference output. No Go process.
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const {spawnSync}=require('node:child_process');
const assets=process.argv[2];
if(!assets)throw Error('expected pinned D2 asset directory');
const expected={
  'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
for(const [name,hash]of Object.entries(expected))if(crypto.createHash('sha256').update(fs.readFileSync(path.join(assets,name))).digest('hex')!==hash)throw Error('unverified '+name);
const wire=cases=>cases.map(c=>JSON.stringify({tex:c.tex,display:c.display,options:{Display:c.display,display:c.display}}).replace(/[\u0080-\uffff]/g,c=>'\\u'+c.charCodeAt(0).toString(16).padStart(4,'0'))).join('\n')+'\n';
const run=(args,input)=>{const p=spawnSync(process.execPath,['--jitless',...args],{input,encoding:'utf8',maxBuffer:32*1024*1024});if(p.status!==0)throw Error(p.stderr||String(p.error));return p.stdout;};
for(const name of ['runtime_environment_end_mathjax_3_2_2.json','runtime_environment_end_residuals.json','runtime_environment_end_repair_proof.json']){
 const file=path.join(__dirname,name),data=JSON.parse(fs.readFileSync(file,'utf8'));
 for(let start=0;start<data.cases.length;start+=24){
  const batch=data.cases.slice(start,start+24),results=run([path.join(__dirname,'differential/oracle.mjs'),'--asset-dir',assets],wire(batch)).replace(/\n$/,'').split('\n').map(JSON.parse);
  if(results.length!==batch.length)throw Error('incomplete original output');
  results.forEach((o,i)=>{
   const c=batch[i];
   if(name.endsWith('mathjax_3_2_2.json')){if(typeof o.svg!=='string'||o.error)throw Error(JSON.stringify(o));c.svg=o.svg;}
   else{
    // Keep complete initial exception stacks. A changed stack must be reviewed,
    // never silently replaced or normalized into an exact assertion.
    if(c.original.error&&JSON.stringify(c.original)!==JSON.stringify(o))throw Error('original runtime changed '+c.name);
    c.original=o;
   }
  });
 }
 fs.writeFileSync(file,JSON.stringify(data,null,2)+'\n');
}
{
 const file=path.join(__dirname,'runtime_environment_end_cleanup_proof.json'),data=JSON.parse(fs.readFileSync(file,'utf8'));
 const out=run([path.join(__dirname,'observe_runtime_script_cleanup.cjs'),assets],wire(data.cases)).replace(/\n$/,'').split('\n').map(JSON.parse);
 if(out.length!==data.cases.length)throw Error('incomplete cleanup observations');
 out.forEach((o,i)=>{const c=data.cases[i];if(c.original.svg?o.original.svg!==c.original.svg:o.original.error.split('\n')[0]!==c.original.error.split('\n')[0])throw Error('observer changed original outcome');if(JSON.stringify(o)!==JSON.stringify(c.observed))throw Error('source cleanup metadata changed');c.observed=o;});
 fs.writeFileSync(file,JSON.stringify(data,null,2)+'\n');
}
{
 const file=path.join(__dirname,'runtime_environment_end_spread_proof.json');
 const fresh=JSON.parse(run([path.join(__dirname,'observe_runtime_spread_values.cjs'),assets],'')),old=JSON.parse(fs.readFileSync(file,'utf8'));
 // The original method's stack includes the observer filename. Preserve its
 // first-capture full stack while checking every other field and first line.
 fresh.rows.forEach((r,i)=>{if(r.error){if(!old.rows[i].error||r.error.split('\n')[0]!==old.rows[i].error.split('\n')[0])throw Error('source spread error changed');r.error=old.rows[i].error;}});
 if(JSON.stringify(fresh)!==JSON.stringify(old))throw Error('source spread observations changed');
 fs.writeFileSync(file,JSON.stringify(fresh,null,2)+'\n');
}
