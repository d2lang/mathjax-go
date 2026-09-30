// SPDX-License-Identifier: Apache-2.0
// node --jitless testdata/generate_ordinary_array_owner.cjs PINNED_ASSETS NEW_OUTPUT OFFICIAL_NEWCOMMAND
// Every output comes from the frozen original. This generator never invokes Go.
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const zlib = require('node:zlib'), vm = require('node:vm'), {spawnSync} = require('node:child_process');
const PIN = 'ad8f5c21cb810236551da8c6512ba733e67357ee';
const UNION = '816903dc9f2beeece895f5a668c58bf4b8128fe163c71be8bccf7b7e5af2155e';
const ASSETS = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'
};
const SUPPLEMENT = 'a41669a4ae924ab83cbc3d08f95ae90490a33cd649ed865e83675011b8708ab9';
const PRIMARY = 'e59040261f8024d42d1c1c1e5d893ab807f8d0625790d5242d215fc285a59432';
const sha = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const equal = (a,b) => JSON.stringify(a) === JSON.stringify(b);
const key = c => JSON.stringify([c.tex,c.display]);
const encode = obj => zlib.gzipSync(JSON.stringify(obj,null,2)+'\n',{level:9});
const readFixture = name => JSON.parse(zlib.gunzipSync(fs.readFileSync(path.join(__dirname,name))));
function assetsAt(directory) {
  return Object.fromEntries(Object.entries(ASSETS).map(([name,hash]) => {
    const bytes = fs.readFileSync(path.join(directory,name));
    if (sha(bytes) !== hash) throw Error('unverified original asset: '+name);
    return [name,bytes.toString()];
  }));
}

// Preserved source observer observe-owner178.cjs SHA256 97e879e06073348ddd527a50f34848b14ea62809d8c4c3c067a53ec7461a4b79
const OWNER_OBSERVER = new vm.Script(String.raw`(()=>{
   let next=0;const ids=new WeakMap();const id=o=>{if(!o)return null;if(!ids.has(o))ids.set(o,++next);return ids.get(o);};
   const snap=o=>!o?null:{id:id(o),kind:o.kind,name:o.getName?o.getName():null,nodes:o.nodes?o.nodes.map(n=>n.kind):[],row:o.row?o.row.length:null,table:o.table?o.table.length:null,envKeys:o.env?Object.keys(o.env).sort():[]};
   const stack=s=>(s.stack||[]).map(snap);
   const B=MathJax._.input.tex.base.BaseItems;
   const relevant=new Set(['end','stop','cell','close','right','over']);
   for(const [label,proto]of[['ArrayItem',B.ArrayItem.prototype],['BeginItem',B.BeginItem.prototype],['OpenItem',B.OpenItem.prototype],['LeftItem',B.LeftItem.prototype],['AutoOpen',MathJax._.input.tex.physics.PhysicsItems.AutoOpen.prototype]]){
    const method=proto.checkItem;
    proto.checkItem=function(item){
     if(!relevant.has(item.kind))return method.apply(this,arguments);
     const event={event:'checkItem',owner:label,before:snap(this),incoming:snap(item)};events.push(event);
     try{const result=method.apply(this,arguments);event.after=snap(this);event.result={success:result?result[1]:null,items:result&&result[0]?result[0].map(snap):null};return result;}
     catch(error){event.error={id:error.id||null,message:String(error.message)};throw error;}
    };
   }
   for(const name of ['EndEntry','EndRow','EndTable']){
    const method=B.ArrayItem.prototype[name];
    B.ArrayItem.prototype[name]=function(){const event={event:name,before:snap(this)};events.push(event);const result=method.apply(this,arguments);event.after=snap(this);return result;};
   }
   const S=MathJax._.input.tex.Stack.default.prototype,pop=S.Pop;
   S.Pop=function(){
    const before=stack(this),observe=before.some(o=>o.kind==='array'||o.kind==='eqnarray'||o.kind==='begin');
    const result=pop.apply(this,arguments);
    if(observe)events.push({event:'Pop',before,popped:snap(result),after:stack(this)});
    return result;
   };
   const C=MathJax._.input.tex.SymbolMap.CommandMap.prototype,parse=C.parse;
   const names=new Set(['begin','end','discard','close','open','qty','eval','text','phantom']);
   C.parse=function(input){
    const[p,name]=input;if(!names.has(name)||!this.contains(name))return parse.apply(this,arguments);
    const event={event:'command',name,map:this.name,beforeStack:stack(p.stack),source:p.string,cursor:p.i,macroCount:p.macroCount};events.push(event);
    try{const result=parse.apply(this,arguments);event.afterStack=stack(p.stack);event.afterMacroCount=p.macroCount;event.afterCursor=p.i;return result;}
    catch(error){event.error={id:error.id||null,message:String(error.message)};throw error;}
   };
   const U=MathJax._.input.tex.ParseUtil.default,charge=U.checkMaxMacros;
   U.checkMaxMacros=function(p,macro=true){
    const observe=p.macroCount>=997;if(!observe)return charge.apply(this,arguments);
    const event={event:'charge',before:p.macroCount,macro,currentCS:p.currentCS,stack:stack(p.stack)};events.push(event);
    try{const result=charge.apply(this,arguments);event.after=p.macroCount;return result;}
    catch(error){event.after=p.macroCount;event.error={id:error.id||null,message:String(error.message)};throw error;}
   };
  })()`);

// Preserved source observer observe-corridor178.cjs SHA256 1d1f4be3bb11359bfb365a420fa6b8f4fb2e60b45eecd1bf7928673548526072
const CORRIDOR_OBSERVER = new vm.Script(String.raw`(()=>{
   let next=0;const ids=new WeakMap();const id=o=>{if(!o)return null;if(!ids.has(o))ids.set(o,++next);return ids.get(o);};
   const snap=o=>!o?null:{id:id(o),kind:o.kind,name:o.getName?o.getName():null,nodes:o.nodes?o.nodes.map(n=>n.kind):[],nodeText:o.nodes?o.nodes.map(n=>typeof n.getText==='function'?n.getText():null):[],row:o.row?o.row.length:null,table:o.table?o.table.length:null,envKeys:o.env?Object.keys(o.env).sort():[],env:o.env?Object.fromEntries(Object.entries(o.env).filter(([k,v])=>v===null||['string','number','boolean'].includes(typeof v))):{}};
   const stack=s=>(s.stack||[]).map(snap);
   const B=MathJax._.input.tex.base.BaseItems;
   const relevant=new Set(['end','stop','cell','close','right','over','mml']);
   for(const [label,proto]of[['ArrayItem',B.ArrayItem.prototype],['BeginItem',B.BeginItem.prototype],['OpenItem',B.OpenItem.prototype],['LeftItem',B.LeftItem.prototype],['AutoOpen',MathJax._.input.tex.physics.PhysicsItems.AutoOpen.prototype]]){
    const method=proto.checkItem;
    proto.checkItem=function(item){
     if(!relevant.has(item.kind))return method.apply(this,arguments);
     const event={event:'checkItem',owner:label,before:snap(this),incoming:snap(item)};events.push(event);
     try{const result=method.apply(this,arguments);event.after=snap(this);event.result={success:result?result[1]:null,items:result&&result[0]?result[0].map(snap):null};return result;}
     catch(error){event.error={id:error.id||null,message:String(error.message)};throw error;}
    };
   }
   for(const name of ['EndEntry','EndRow','EndTable']){
    const method=B.ArrayItem.prototype[name];
    B.ArrayItem.prototype[name]=function(){const event={event:name,before:snap(this)};events.push(event);const result=method.apply(this,arguments);event.after=snap(this);return result;};
   }
   const S=MathJax._.input.tex.Stack.default.prototype,pop=S.Pop;
   S.Pop=function(){
    const before=stack(this),observe=before.some(o=>o.kind==='array'||o.kind==='eqnarray'||o.kind==='begin');
    const result=pop.apply(this,arguments);
    if(observe)events.push({event:'Pop',before,popped:snap(result),after:stack(this)});
    return result;
   };
   const C=MathJax._.input.tex.SymbolMap.CommandMap.prototype,parse=C.parse;
   const names=new Set(['begin','end','discard','close','open','qty','eval','text','phantom']);
   C.parse=function(input){
    const[p,name]=input;if(!names.has(name)||!this.contains(name))return parse.apply(this,arguments);
    const event={event:'command',name,map:this.name,beforeStack:stack(p.stack),source:p.string,cursor:p.i,macroCount:p.macroCount};events.push(event);
    try{const result=parse.apply(this,arguments);event.afterStack=stack(p.stack);event.afterMacroCount=p.macroCount;event.afterCursor=p.i;return result;}
    catch(error){event.error={id:error.id||null,message:String(error.message)};throw error;}
   };
   const U=MathJax._.input.tex.ParseUtil.default,charge=U.checkMaxMacros;
   U.checkMaxMacros=function(p,macro=true){
    const observe=p.macroCount>=997;if(!observe)return charge.apply(this,arguments);
    const event={event:'charge',before:p.macroCount,macro,currentCS:p.currentCS,stack:stack(p.stack)};events.push(event);
    try{const result=charge.apply(this,arguments);event.after=p.macroCount;return result;}
    catch(error){event.after=p.macroCount;event.error={id:error.id||null,message:String(error.message)};throw error;}
   };
  })()`);

function observe(request, scripts, instrumentation) {
  const context = vm.createContext({console,request,events:[]});
  context.globalThis = context;
  for (const script of scripts) script.runInContext(context);
  if (instrumentation) instrumentation.runInContext(context);
  const output = new vm.Script('({svg:adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}))})').runInContext(context,{timeout:15000});
  return {output,events:context.events};
}
function worker(mode, directory, supplementPath) {
  const source = assetsAt(directory), requests = JSON.parse(fs.readFileSync(0,'utf8'));
  const scripts = Object.entries(source).map(([name,text]) => new vm.Script(text,{filename:name}));
  let rows;
  if (mode === 'augmented') {
    const component = fs.readFileSync(supplementPath), primary = fs.readFileSync(path.join(__dirname,'macro_boundary_primary.js'));
    if (sha(component) !== SUPPLEMENT || sha(primary) !== PRIMARY) throw Error('unbound augmented component/harness');
    function run(request,passive) {
      const c = vm.createContext({console,request,passive});
      for (const name of ['polyfills.js','mathjax.js']) new vm.Script(source[name],{filename:name}).runInContext(c);
      new vm.Script(component.toString(),{filename:'official-3.2.2/newcommand.js'}).runInContext(c);
      if (source['setup.js'].split("'physics'").length !== 2) throw Error('augmented setup anchor changed');
      new vm.Script(source['setup.js'].replace("'physics'","'physics', 'newcommand'"),{filename:'newcommand-augmented-setup.js'}).runInContext(c);
      return new vm.Script(primary.toString(),{filename:'macro_boundary_primary.js'}).runInContext(c,{timeout:15000});
    }
    rows = requests.map(request => ({request,plain:run(request,false),observed:run(request,true)}));
  } else {
    if (!['owner','corridor'].includes(mode)) throw Error('unknown observer');
    rows = requests.map(request => {
      const plain = observe(request,scripts,null);
      const observed = observe(request,scripts,mode==='owner'?OWNER_OBSERVER:CORRIDOR_OBSERVER);
      if (!equal(plain.output,observed.output)) throw Error('observer changed complete output');
      return {...request,original:plain.output,observed};
    });
  }
  process.stdout.write(JSON.stringify(rows));
}
if (process.argv[2] === '--worker') {
  worker(process.argv[3],process.argv[4],process.argv[5]);
} else {
  main();
}

function main() {
  const [directory,output,supplementPath] = process.argv.slice(2);
  if (!directory || !output || !supplementPath || fs.existsSync(output)) throw Error('use PINNED_ASSETS, a nonexistent NEW_OUTPUT directory, and OFFICIAL_NEWCOMMAND');
  assetsAt(directory);
  if (sha(fs.readFileSync(supplementPath)) !== SUPPLEMENT || sha(fs.readFileSync(path.join(__dirname,'macro_boundary_primary.js'))) !== PRIMARY) throw Error('unverified augmented source');
  fs.mkdirSync(output);
  const mainName = 'ordinary_array_owner_mathjax_3_2_2.json.gz', residualName = 'ordinary_array_owner_residuals.json.gz';
  const fixtures = [mainName,residualName].map(name => ({name,value:readFixture(name)}));
  for (const {value} of fixtures) if (value.mathjaxGitCommit !== PIN || value.sourceUnionSHA256 !== UNION || !equal(value.originalAssets,ASSETS)) throw Error('unbound public fixture');
  const cases = fixtures.flatMap(f => f.value.cases).sort((a,b) => a.sourceIndex-b.sourceIndex);
  if (cases.length !== 6124 || fixtures[0].value.cases.length !== 5852 || fixtures[1].value.cases.length !== 272) throw Error('public container counts changed');
  const freshByInput = new Map(), runtimeObservations = [], partitions = {}, names = new Set(), reports = [];
  let originalSVGs = 0, sourceMemberships = 0;
  function run(args,input,label) {
    const result = spawnSync(process.execPath,['--jitless',...args],{input,encoding:'utf8',maxBuffer:128*1024*1024,timeout:60000});
    if (result.status !== 0 || result.error) throw Error(label+': '+result.error+' '+result.stderr);
    return result.stdout;
  }
  for (let start=0;start<cases.length;start+=24) {
    const batch = cases.slice(start,start+24);
    const wire = batch.map(c => JSON.stringify({tex:c.tex,options:{Display:c.display,Em:16,Ex:8}}).replace(/\u2028/g,'\\u2028').replace(/\u2029/g,'\\u2029')).join('\n')+'\n';
    const stdout = run([path.join(__dirname,'differential/oracle.mjs'),'--asset-dir',directory],wire,'original runtime');
    const fresh = stdout.replace(/\n$/,'').split('\n').map(JSON.parse);
    if (fresh.length !== batch.length) throw Error('incomplete original capture');
    fresh.forEach((row,j) => {
      const c = batch[j], k = key(c);
      if (c.sourceIndex !== start+j || names.has(c.name) || freshByInput.has(k) || typeof c.display !== 'boolean') throw Error('duplicate/unbound original case');
      names.add(c.name);freshByInput.set(k,row);sourceMemberships += c.sourceMemberships.length;
      partitions[c.partition]=(partitions[c.partition]||0)+1;
      if (c.original.svg) {
        if (!['strict-svg','safe-xml','held-svg'].includes(c.partition) || !equal(row,c.original)) throw Error('complete original SVG changed: '+c.name);
        c.original = row;
        originalSVGs++;
      } else {
        if (c.partition !== 'runtime' || row.svg || typeof row.error !== 'string' || Object.keys(row).length !== 1 ||
            row.error.split('\n')[0] !== c.contract.originalFirstLine || c.original.error.split('\n')[0] !== c.contract.originalFirstLine) throw Error('original runtime contract changed: '+c.name);
        runtimeObservations.push({name:c.name,tex:c.tex,display:c.display,contract:c.contract,
          historical:c.original,historicalObjectSHA256:sha(JSON.stringify(c.original)),fresh:row,fullObjectEqual:equal(row,c.original)});
      }
    });
    if ((start+batch.length)%240===0 || start+batch.length===cases.length) console.log(JSON.stringify({stage:'public originals',completed:start+batch.length,total:cases.length}));
  }
  if (originalSVGs !== 5976 || runtimeObservations.length !== 148 || sourceMemberships !== 6244 ||
      partitions['strict-svg'] !== 5608 || partitions['safe-xml'] !== 244 || partitions['held-svg'] !== 124 || partitions.runtime !== 148) throw Error('public source inventory changed');
  function writeVerified(name,value,metadata) {
    const input = fs.readFileSync(path.join(__dirname,name)), regenerated = encode(value);
    fs.writeFileSync(path.join(output,name),regenerated,{flag:'wx'});
    const report = {file:name,inputSHA256:sha(input),regeneratedSHA256:sha(regenerated),byteIdentical:input.equals(regenerated),...metadata};
    reports.push(report);
    if (!report.byteIdentical) throw Error('original regenerated bytes changed: '+name);
  }
  for (const {name,value} of fixtures) writeVerified(name,value,{cases:value.cases.length});
  const streams = [
    ['ordinary_array_owner_observations.json.gz','owner',222],
    ['ordinary_array_owner_corridor_observations.json.gz','corridor',406],
    ['ordinary_array_owner_repair_observations.json.gz','corridor',48],
    ['ordinary_array_owner_literal_observations.json.gz','corridor',96]
  ];
  const publicPassiveInputs = new Set();
  let publicPassiveObjects = 0;
  for (const [name,mode,count] of streams) {
    const fixture = readFixture(name), seen = new Set();
    if (fixture.sourcePin !== PIN || !equal(fixture.assets,ASSETS) || fixture.cases.length !== count) throw Error('unbound source observations');
    for (let start=0;start<count;start+=12) {
      const batch = fixture.cases.slice(start,start+12);
      const requests = batch.map(c => Object.fromEntries(Object.entries(c).filter(([k]) => !['original','observed'].includes(k))));
      const repeated = JSON.parse(run([__filename,'--worker',mode,directory],JSON.stringify(requests),'source observer'));
      if (repeated.length !== batch.length) throw Error('incomplete passive capture');
      repeated.forEach((row,j) => {
        const c=batch[j],k=key(c);
        if (seen.has(k) || !equal(freshByInput.get(k),c.original) || !equal(row,c)) throw Error('full passive source object changed: '+name+' '+(start+j));
        seen.add(k);publicPassiveInputs.add(k);publicPassiveObjects++;
        fixture.cases[start+j]=row;
      });
    }
    writeVerified(name,fixture,{observations:count,additionalPublicInputs:0});
    console.log(JSON.stringify({stage:'public source observations',file:name,completeObjects:count}));
  }
  if (publicPassiveObjects !== 772 || publicPassiveInputs.size !== 770) throw Error('public observer overlap changed');
  const augmentedName='ordinary_array_owner_augmented_methods.json.gz', augmented=readFixture(augmentedName);
  if (augmented.cases.length !== 48 || !equal(augmented.pins,ASSETS) || augmented.supplementSHA256 !== SUPPLEMENT || augmented.primarySHA256 !== PRIMARY) throw Error('unbound augmented method profile');
  const augmentedOutcomes = {};
  for (let start=0;start<48;start+=12) {
    const batch=augmented.cases.slice(start,start+12), requests=batch.map(c=>c.request);
    const repeated=JSON.parse(run([__filename,'--worker','augmented',directory,supplementPath],JSON.stringify(requests),'augmented original methods'));
    if (repeated.length !== batch.length) throw Error('incomplete augmented capture');
    repeated.forEach((row,j) => {
      if (!equal(row,batch[j])) throw Error('complete augmented method record changed: '+(start+j));
      const plain={...row.plain}, observed={...row.observed};delete plain.trace;delete observed.trace;
      if (!equal(plain,observed)) throw Error('augmented observer changed complete output/tree');
      const diagnostic=row.plain.formattedErrors.length?row.plain.formattedErrors[0].id:'valid';
      augmentedOutcomes[diagnostic]=(augmentedOutcomes[diagnostic]||0)+1;
      augmented.cases[start+j]=row;
    });
  }
  if (augmentedOutcomes.valid!==12 || augmentedOutcomes.UnknownEnv!==18 || augmentedOutcomes.MaxMacroSub2!==18) throw Error('augmented outcome inventory changed');
  writeVerified(augmentedName,augmented,{methodObservations:48,additionalPublicInputs:0,additionalPublicPassiveObservations:0,profile:augmented.profile});
  const receipt={source:'frozen original only; no Go invocation',mathjaxGitCommit:PIN,originalAssets:ASSETS,nodeVersion:process.version,
    generatorSHA256:sha(fs.readFileSync(__filename)),sourceUnionSHA256:UNION,uniquePublicInputs:6124,completeOriginalSVGs:originalSVGs,
    partitions,sourceMemberships,publicPassiveObservationObjects:publicPassiveObjects,publicPassiveUniqueExistingInputs:publicPassiveInputs.size,
    separateAugmentedMethodObservations:48,augmentedOutcomes,augmentedCounterPolicy:'Official component with source method count seeded by the pinned primary harness; no public input claim.',
    historicalRuntimeObjectsPreserved:runtimeObservations.length,freshRuntimeObjectsFullyEqual:runtimeObservations.filter(r=>r.fullObjectEqual).length,
    runtimePolicy:'Complete historical objects remain unchanged. Fresh full exceptions are retained separately; only the explicitly bound first-line runtime class tolerates differing stack callsites.',
    runtimeObservations,files:reports};
  fs.writeFileSync(path.join(output,'ordinary_array_owner_regeneration.json.gz'),encode(receipt),{flag:'wx'});
  console.log(JSON.stringify({status:'PASS',files:reports,originalSVGs,runtime:runtimeObservations.length,publicPassiveObjects,augmentedMethodObjects:48},null,2));
}
