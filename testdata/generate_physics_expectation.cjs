// SPDX-License-Identifier: Apache-2.0
// node --jitless testdata/generate_physics_expectation.cjs PINNED_ASSETS SCRATCH_OUTPUT
// Only the hash-verified original supplies output; no Go process is run.
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const zlib=require('node:zlib'),vm=require('node:vm'),{spawnSync}=require('node:child_process');
const assets=process.argv[2],output=process.argv[3],sha=x=>crypto.createHash('sha256').update(x).digest('hex');
if(!assets||!output||path.resolve(output)===path.resolve(__dirname))throw Error('use PINNED_ASSETS and a separate SCRATCH_OUTPUT');
fs.mkdirSync(output,{recursive:true});
const name='physics_expectation_mathjax_3_2_2.json.gz',input=fs.readFileSync(path.join(__dirname,name));
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
if(svgCount!==1748||runtimeObservations.length!==48||freshByInput.size!==1796)throw Error('original partition changed');
const bytes=zlib.gzipSync(JSON.stringify(data,null,2)+'\n',{level:9});fs.writeFileSync(path.join(output,name),bytes);
const reports=[{file:name,inputs:1796,fullOriginalSVGs:svgCount,runtimeObjectsPreserved:48,inputSHA256:sha(input),regeneratedSHA256:sha(bytes),byteIdentical:input.equals(bytes)}];

// Reproduce passive source-child observations without re-evaluating mml or
// modifying the returned tree. The primary output must remain byte-identical.
const fontName='physics_expectation_font_observations.json.gz',fontBytes=fs.readFileSync(path.join(__dirname,fontName));
const fonts=JSON.parse(zlib.gunzipSync(fontBytes));
if(fonts.cases.length!==36||JSON.stringify(fonts.assetSHA256)!==JSON.stringify(data.originalAssets))throw Error('unbound font observations');
for(const request of fonts.cases){
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
 if(!fresh||JSON.stringify(fresh)!==JSON.stringify(request.original)||JSON.stringify(observed)!==JSON.stringify(request.observed))throw Error('source child observation changed: '+request.tex);
}
// These two gzip files retain their historical byte containers after their
// complete source observations have been re-established, without extra inputs.
fs.writeFileSync(path.join(output,fontName),fontBytes);
reports.push({file:fontName,observations:36,additionalPublicInputs:0,inputSHA256:sha(fontBytes),regeneratedSHA256:sha(fontBytes),byteIdentical:true});
const budgetName='physics_expectation_budget_observations.json.gz',budgetBytes=fs.readFileSync(path.join(__dirname,budgetName));
const budgets=JSON.parse(zlib.gunzipSync(budgetBytes));if(budgets.cases.length!==54)throw Error('unbound budget inventory');
let budgetErrors=0;
for(const c of budgets.cases){
 const fresh=freshByInput.get(JSON.stringify([c.tex,c.display]));
 if(!fresh||JSON.stringify(fresh)!==JSON.stringify(c.original))throw Error('source budget outcome changed');
 const match=fresh.svg.match(/data-mjx-error="([^"]*)"/),errorText=match?match[1]:null;
 if(errorText!==c.capturedErrorText)throw Error('source budget diagnostic changed');
 if(errorText)budgetErrors++;
}
if(budgetErrors!==12)throw Error('source budget boundary changed');
fs.writeFileSync(path.join(output,budgetName),budgetBytes);
reports.push({file:budgetName,observations:54,additionalPublicInputs:0,inputSHA256:sha(budgetBytes),regeneratedSHA256:sha(budgetBytes),byteIdentical:true});
fs.writeFileSync(path.join(output,'physics_expectation_regeneration.json.gz'),zlib.gzipSync(JSON.stringify({mathjaxGitCommit:data.mathjaxGitCommit,files:reports,uniquePublicInputs:1796,source:'frozen original only',runtimePolicy:'historical complete exceptions unchanged; fresh complete source failures retained below',runtimeObservations},null,2)+'\n',{level:9}));
console.log(JSON.stringify(reports,null,2));
if(reports.some(r=>!r.byteIdentical))throw Error('original fixture bytes changed');
