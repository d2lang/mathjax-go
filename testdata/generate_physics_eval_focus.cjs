// SPDX-License-Identifier: Apache-2.0
// Usage: node --jitless testdata/generate_physics_eval_focus.cjs PINNED_ASSETS
// Only the hash-verified original creates reference SVGs; no Go process.
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto'),{spawnSync}=require('node:child_process');
const file=path.join(__dirname,'physics_eval_focus_mathjax_3_2_2.json'),data=JSON.parse(fs.readFileSync(file,'utf8')),assets=process.argv[2];
if(!assets||data.mathjaxGitCommit!=='ad8f5c21cb810236551da8c6512ba733e67357ee')throw Error('unbound original inputs');
for(const [name,hash]of Object.entries(data.originalAssets))if(crypto.createHash('sha256').update(fs.readFileSync(path.join(assets,name))).digest('hex')!==hash)throw Error('unverified '+name);
for(let start=0;start<data.cases.length;start+=24){
 const cases=data.cases.slice(start,start+24);
 const input=cases.map(c=>JSON.stringify({tex:c.tex,display:c.display,options:{display:c.display,Display:c.display}})).join('\n')+'\n';
 const run=spawnSync(process.execPath,['--jitless',path.join(__dirname,'differential/oracle.mjs'),'--asset-dir',assets],{input,encoding:'utf8',maxBuffer:16*1024*1024});
 if(run.status!==0)throw Error(run.stderr);
 const rows=run.stdout.replace(/\n$/,'').split('\n').map(JSON.parse);
 if(rows.length!==cases.length)throw Error('incomplete original capture');
 rows.forEach((o,i)=>{if(typeof o.svg!=='string'||o.error)throw Error('original changed to runtime');cases[i].svg=o.svg;});
}
fs.writeFileSync(file,JSON.stringify(data,null,2)+'\n');
