// SPDX-License-Identifier: Apache-2.0
// Usage: node --jitless testdata/generate_enclosure_bbox.cjs PINNED_ASSETS
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const {spawnSync}=require('node:child_process');
const assets=process.argv[2];
if(!assets)throw Error('expected pinned D2 asset directory');
const hashes={
 'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
 'mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
 'setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
for(const[file,expected]of Object.entries(hashes)){
 if(crypto.createHash('sha256').update(fs.readFileSync(path.join(assets,file))).digest('hex')!==expected)throw Error(`unverified ${file}`);
}
function run(cases,observer){
 const requests=cases.map(c=>JSON.stringify({tex:c.tex,display:c.display,options:{display:c.display},notation:c.notation})).join('\n')+'\n';
 const args=observer?[path.join(__dirname,observer),assets]:[path.join(__dirname,'differential/oracle.mjs'),'--asset-dir',assets];
 const p=spawnSync(process.execPath,['--jitless',...args],{input:requests,encoding:'utf8',maxBuffer:16*1024*1024});
 if(p.status!==0)throw Error(p.stderr||String(p.error));
 const values=p.stdout.trim().split('\n').map(JSON.parse);
 if(values.length!==cases.length)throw Error('incomplete original output');
 return values;
}
for(const[name,kind,observer]of[
 ['enclosure_bbox_mathjax_3_2_2.json','svg',null],
 ['enclosure_bbox_residuals.json','raw',null],
 ['enclosure_notations_mathjax_3_2_2.json','notation','enclosure_notation_observer.mjs'],
 ['enclosure_lifecycle_mathjax_3_2_2.json','lifecycle','enclosure_lifecycle_observer.mjs'],
]){
 const file=path.join(__dirname,name),fixture=JSON.parse(fs.readFileSync(file,'utf8'));
 for(let i=0;i<fixture.cases.length;i+=24){
  const batch=fixture.cases.slice(i,i+24),values=run(batch,observer);
  const unmodified=kind==='lifecycle'?run(batch,null):null;
  values.forEach((value,j)=>{
   const c=batch[j];
   if(kind==='svg'){if(typeof value.svg!=='string'||value.error)throw Error(JSON.stringify(value));c.svg=value.svg;}
   else if(kind==='lifecycle'){
    if(value.svg!==unmodified[j].svg)throw Error('observer changed original SVG');
    c.original=unmodified[j];c.lifecycle=value.lifecycle;
   }else c.original=value;
  });
 }
 fs.writeFileSync(file,JSON.stringify(fixture,null,2)+'\n');
}
