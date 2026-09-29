// SPDX-License-Identifier: Apache-2.0
// node --jitless testdata/generate_physics_ketbra.cjs PINNED_ASSETS SCRATCH_OUTPUT
// Only the hash-verified original supplies output; no Go process is run.
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const zlib=require('node:zlib'),vm=require('node:vm'),{spawnSync}=require('node:child_process');
const assets=process.argv[2],output=process.argv[3],sha=x=>crypto.createHash('sha256').update(x).digest('hex');
if(!assets||!output||path.resolve(output)===path.resolve(__dirname))throw Error('use PINNED_ASSETS and a separate SCRATCH_OUTPUT');
fs.mkdirSync(output,{recursive:true});
const name='physics_ketbra_mathjax_3_2_2.json.gz',input=fs.readFileSync(path.join(__dirname,name));
const data=JSON.parse(zlib.gunzipSync(input));
if(data.mathjaxGitCommit!=='ad8f5c21cb810236551da8c6512ba733e67357ee'||data.cases.length!==1796)throw Error('unbound original inputs');
const scripts=Object.entries(data.originalAssets).map(([name,hash])=>{
 const bytes=fs.readFileSync(path.join(assets,name));if(sha(bytes)!==hash)throw Error('unverified '+name);
 return new vm.Script(bytes.toString(),{filename:name});
});
const firstLine="TypeError: Cannot read properties of null (reading '4')",freshByInput=new Map(),runtimeObservations=[];
let svgCount=0;
for(let start=0;start<data.cases.length;start+=24){
 const cases=data.cases.slice(start,start+24);
 const wire=cases.map(c=>JSON.stringify({tex:c.tex,display:c.display,options:{display:c.display,Display:c.display}})
  .replace(/\u2028/g,'\\u2028').replace(/\u2029/g,'\\u2029')).join('\n')+'\n';
 const run=spawnSync(process.execPath,['--jitless',path.join(__dirname,'differential/oracle.mjs'),'--asset-dir',assets],{input:wire,encoding:'utf8',maxBuffer:64*1024*1024});
 if(run.status!==0)throw Error('original process failed: '+run.stderr);
 const rows=run.stdout.replace(/\n$/,'').split('\n').map(JSON.parse);if(rows.length!==cases.length)throw Error('incomplete original capture');
 rows.forEach((fresh,i)=>{
  const c=cases[i],key=JSON.stringify([c.tex,c.display]);if(freshByInput.has(key))throw Error('duplicate input');freshByInput.set(key,fresh);
  if(c.partition==='strict-svg'){
   if(fresh.error||fresh.svg!==c.original.svg||Object.keys(fresh).join()!=='svg')throw Error('complete original SVG changed: '+c.name);
   c.original=fresh;svgCount++;
  }else{
   if(c.partition!=='runtime-bounded-nel'||!c.tex.includes('\u0085')||c.original.svg||fresh.svg||
    typeof fresh.error!=='string'||c.original.error.split('\n')[0]!==firstLine||fresh.error.split('\n')[0]!==firstLine)throw Error('runtime source contract changed: '+c.name);
   runtimeObservations.push({name:c.name,tex:c.tex,display:c.display,preservedErrorSHA256:sha(c.original.error),fresh});
  }
 });
}
if(svgCount!==1764||runtimeObservations.length!==32||freshByInput.size!==1796)throw Error('original partition changed');
const bytes=zlib.gzipSync(JSON.stringify(data,null,2)+'\n',{level:9});fs.writeFileSync(path.join(output,name),bytes);
const reports=[{file:name,inputs:1796,fullOriginalSVGs:svgCount,runtimeObjectsPreserved:32,inputSHA256:sha(input),regeneratedSHA256:sha(bytes),byteIdentical:input.equals(bytes)}];

// Both evidence sets are passive original parser-entry/return observations.
// Each original method runs exactly once; its result remains unchanged.
function verifyPassive(file,count){
 const bytes=fs.readFileSync(path.join(__dirname,file)),dataSet=JSON.parse(zlib.gunzipSync(bytes));
 if(dataSet.cases.length!==count||dataSet.mathjaxGitCommit!==data.mathjaxGitCommit||JSON.stringify(dataSet.assetSHA256)!==JSON.stringify(data.originalAssets))throw Error('unbound passive evidence: '+file);
 for(const request of dataSet.cases){
  const context=vm.createContext({console,request,events:[]});context.globalThis=context;
  for(const script of scripts)script.runInContext(context);
  new vm.Script(`(()=>{
   const p=MathJax._.input.tex.TexParser.default.prototype,mml=p.mml,parse=p.Parse;
   p.Parse=function(){const env={};for(const key of Object.keys(this.stack.env)){const value=this.stack.env[key];env[key]=value instanceof RegExp?String(value):value;}events.push({phase:'parse-entry',source:this.string,cursor:this.i,macroCount:this.macroCount,env});return parse.apply(this,arguments);};
   const tree=n=>({kind:n.kind,inferred:!!n.isInferred,text:n.kind==='text'?n.getText():undefined,children:(n.childNodes||[]).map(tree)});
   p.mml=function(){const result=mml.apply(this,arguments);const env={};for(const key of Object.keys(this.stack.env)){const value=this.stack.env[key];env[key]=value instanceof RegExp?String(value):value;}events.push({source:this.string,cursor:this.i,macroCount:this.macroCount,env,tree:tree(result)});return result;};
  })()`).runInContext(context);
  const observed=new vm.Script('({output:{svg:adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}))},events})').runInContext(context);
  const fresh=freshByInput.get(JSON.stringify([request.tex,request.display]));
  if(!fresh||JSON.stringify(fresh)!==JSON.stringify(request.original)||JSON.stringify(observed)!==JSON.stringify(request.observed))throw Error('passive source observation changed: '+file+' '+request.tex);
 }
 // Preserve the original evidence container after re-establishing every event
 // and complete output; these inputs already occur in the public fixture.
 fs.writeFileSync(path.join(output,file),bytes);
 reports.push({file,observations:count,additionalPublicInputs:0,inputSHA256:sha(bytes),regeneratedSHA256:sha(bytes),byteIdentical:true});
}
verifyPassive('physics_ketbra_font_observations.json.gz',48);
verifyPassive('physics_ketbra_budget_observations.json.gz',16);
fs.writeFileSync(path.join(output,'physics_ketbra_regeneration.json.gz'),zlib.gzipSync(JSON.stringify({mathjaxGitCommit:data.mathjaxGitCommit,files:reports,uniquePublicInputs:1796,source:'frozen original only',runtimePolicy:'historical complete exceptions unchanged; fresh complete source failures retained below',runtimeObservations},null,2)+'\n',{level:9}));
console.log(JSON.stringify(reports,null,2));
if(reports.some(r=>!r.byteIdentical))throw Error('original fixture bytes changed');
