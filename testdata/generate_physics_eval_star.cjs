// SPDX-License-Identifier: Apache-2.0
// Usage: node --jitless testdata/generate_physics_eval_star.cjs PINNED_ASSETS
// References come only from the hash-verified original; this never runs Go.
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto'),vm=require('node:vm'),{spawnSync}=require('node:child_process');
const publicFile=path.join(__dirname,'physics_eval_star_mathjax_3_2_2.json');
const methodFile=path.join(__dirname,'../internal/tex/testdata/eval_getstar_mathjax_3_2_2.json');
const data=JSON.parse(fs.readFileSync(publicFile,'utf8')),methods=JSON.parse(fs.readFileSync(methodFile,'utf8')),assets=process.argv[2];
const pin='ad8f5c21cb810236551da8c6512ba733e67357ee';
if(!assets||data.mathjaxGitCommit!==pin||methods.mathjaxGitCommit!==pin)throw Error('unbound original inputs');
for(const [name,hash]of Object.entries(data.originalAssets))if(crypto.createHash('sha256').update(fs.readFileSync(path.join(assets,name))).digest('hex')!==hash||methods.assetsSHA256[name]!==hash)throw Error('unverified '+name);
for(let start=0;start<data.cases.length;start+=24){
 const cases=data.cases.slice(start,start+24);
 const input=cases.map(c=>JSON.stringify({tex:c.tex,display:c.display,options:{display:c.display,Display:c.display}}).replace(/\u2028/g,'\\u2028').replace(/\u2029/g,'\\u2029')).join('\n')+'\n';
 const run=spawnSync(process.execPath,['--jitless',path.join(__dirname,'differential/oracle.mjs'),'--asset-dir',assets],{input,encoding:'utf8',maxBuffer:16*1024*1024});
 if(run.status!==0)throw Error(run.stderr);
 const rows=run.stdout.replace(/\n$/,'').split('\n').map(JSON.parse);
 if(rows.length!==cases.length)throw Error(`incomplete original capture at ${start}: ${rows.length}/${cases.length}`);
 rows.forEach((o,i)=>{if(typeof o.svg!=='string'||o.error)throw Error('original changed to runtime');cases[i].svg=o.svg;});
}
const c=vm.createContext({console});c.globalThis=c;
for(const name of Object.keys(data.originalAssets))new vm.Script(fs.readFileSync(path.join(assets,name),'utf8'),{filename:name}).runInContext(c);
c.requests=methods.cases.map(({source,point})=>point===undefined?{source}:{source,point});
const observed=vm.runInContext(`(()=>{const p=MathJax._.input.tex.TexParser.default.prototype;return {methods:{GetStar:p.GetStar.toString(),GetNext:p.GetNext.toString(),nextIsSpace:p.nextIsSpace.toString()},cases:requests.map(r=>{const x=Object.create(p);x.string=r.source;x.i=0;const value=p.GetStar.call(x);return {...r,value,cursorCodeUnits:x.i,consumedBytes:unescape(encodeURIComponent(x.string.slice(0,x.i))).length,remaining:x.string.slice(x.i)}})}})()`,c);
methods.methods=observed.methods;methods.cases=observed.cases;
fs.writeFileSync(publicFile,JSON.stringify(data,null,2)+'\n');
fs.writeFileSync(methodFile,JSON.stringify(methods,null,2)+'\n');
