// SPDX-License-Identifier: Apache-2.0
// node --jitless testdata/generate_physics_quantity.cjs PINNED_ASSETS
// All complete reference outcomes come from the original; no Go process.
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto'),zlib=require('node:zlib'),{spawnSync}=require('node:child_process');
const fixtureName=process.argv[3]||'physics_quantity_mathjax_3_2_2.json.gz';
const counts={'physics_quantity_mathjax_3_2_2.json.gz':2776,'physics_final_item_siblings_mathjax_3_2_2.json.gz':312,'physics_quantity_residuals.json':4};
if(!counts[fixtureName])throw Error('unknown original corpus');
const file=path.join(__dirname,fixtureName);
const raw=fs.readFileSync(file);
const data=JSON.parse(file.endsWith('.gz')?zlib.gunzipSync(raw):raw),assets=process.argv[2];
if(!assets||data.mathjaxGitCommit!=='ad8f5c21cb810236551da8c6512ba733e67357ee'||data.cases.length!==counts[fixtureName])throw Error('unbound original inputs');
for(const [name,hash]of Object.entries(data.originalAssets))if(crypto.createHash('sha256').update(fs.readFileSync(path.join(assets,name))).digest('hex')!==hash)throw Error('unverified '+name);
for(let start=0;start<data.cases.length;start+=24){
 const cases=data.cases.slice(start,start+24);
 const input=cases.map(c=>JSON.stringify({tex:c.tex,display:c.display,options:{display:c.display,Display:c.display}}).replace(/\u2028/g,'\\u2028').replace(/\u2029/g,'\\u2029')).join('\n')+'\n';
 const run=spawnSync(process.execPath,['--jitless',path.join(__dirname,'differential/oracle.mjs'),'--asset-dir',assets],{input,encoding:'utf8',maxBuffer:64*1024*1024});
 if(run.status!==0)throw Error(run.stderr);
 const rows=run.stdout.replace(/\n$/,'').split('\n').map(JSON.parse);
 if(rows.length!==cases.length)throw Error('incomplete original capture');
 rows.forEach((original,i)=>{if(typeof original.svg!=='string'&&typeof original.error!=='string')throw Error('missing original outcome');cases[i].original=original;});
}
const output=JSON.stringify(data,null,2)+'\n';
fs.writeFileSync(file,file.endsWith('.gz')?zlib.gzipSync(output):output);
console.log(JSON.stringify({cases:data.cases.length,svg:data.cases.filter(c=>c.original.svg).length,runtime:data.cases.filter(c=>c.original.error).length}));
