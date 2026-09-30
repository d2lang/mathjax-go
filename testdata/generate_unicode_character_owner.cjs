// SPDX-License-Identifier: Apache-2.0
// node --jitless testdata/generate_unicode_character_owner.cjs PINNED_ASSETS NEW_OUTPUT
// Original-only: fresh frozen MathJax state for every conversion, never Go output.
'use strict';
const fs = require('node:fs'), path = require('node:path'), vm = require('node:vm');
const zlib = require('node:zlib'), crypto = require('node:crypto'), assert = require('node:assert/strict');
const PIN = 'ad8f5c21cb810236551da8c6512ba733e67357ee';
const UNION = '4a5d99ff13db6c0ca8dcd349ce42c6d8638db0598536a270b8b058092c097ac1';
const NODE_SHA = '27db838bb204ef7c21df2931f5656e4c8fb32e6e947f363a402b49714d32b5b1';
const ASSETS = {
  'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
const FILES = [
  'unicode_character_owner_mathjax_3_2_2.json.gz',
  'unicode_character_owner_residuals.json.gz',
  'unicode_character_owner_runtime_originals.json.gz',
  'unicode_character_owner_dispatch_observations.json.gz',
  'unicode_character_owner_range_observations.json.gz',
  'unicode_character_owner_prime_supplement.json.gz',
];
const sha = b => crypto.createHash('sha256').update(b).digest('hex');
const key = r => JSON.stringify([r.tex,r.display]);
const clone = value => JSON.parse(JSON.stringify(value));
// Shared by the source exporter and regenerator. Node writes fixed mtime=0 headers.
const encode = value => zlib.gzipSync(JSON.stringify(value,null,2)+'\n',{level:9});
const PILOT_OBSERVER_BODY = "(()=>{\n   const M=MathJax._.input.tex.MapHandler.MapHandler,letter=M.getMap('letter'),digit=M.getMap('digit');\n   globalThis.dispatch={letterPattern:String(letter._regExp),digitPattern:String(digit._regExp),letterMatches:letter.contains(request.characterText),digitMatches:digit.contains(request.characterText)};\n   const P=MathJax._.input.tex.TexParser.default.prototype,create=P.create;\n   P.create=function(kind,...args){\n    const event=kind==='token'?{kind:args[0],attributes:JSON.parse(JSON.stringify(args[1]||{})),text:args[2],source:this.string,cursor:this.i,font:this.stack.env.font||null}:null;\n    const result=create.apply(this,arguments);\n    if(event){event.createdKind=result.kind;event.createdText=result.getText();event.createdAttributes=result.attributes.getAllAttributes();event.properties=result.getAllProperties();events.push(event);}\n    return result;\n   };\n  })()";
const RANGE_OBSERVER_BODY = "(()=>{\n const T=MathJax._.input.tex.TexParser.default.prototype;\n const D=MathJax._.core.MmlTree.OperatorDictionary;\n const ids=new WeakMap(),wrappedMaps=new WeakSet(),wrappedFallbacks=new WeakSet(),wrappedConfigurations=new WeakSet();let next=0,activeParser=null;\n const id=o=>{if(!o||typeof o!=='object')return null;if(!ids.has(o))ids.set(o,++next);return ids.get(o);};\n const encode=(value,seen=new Map())=>{\n  if(value===undefined)return{type:'undefined'};\n  if(value===null||['string','number','boolean'].includes(typeof value))return value;\n  if(typeof value==='function')return{type:'function',name:value.name,source:Function.prototype.toString.call(value)};\n  if(Object.prototype.toString.call(value)==='[object RegExp]')return{type:'RegExp',source:value.source,flags:value.flags,lastIndex:value.lastIndex};\n  if(seen.has(value))return{type:'reference',path:seen.get(value)};\n  seen.set(value,seen.size);\n  if(Array.isArray(value))return value.map(x=>encode(x,seen));\n  const result={};for(const key of Object.keys(value).sort())result[key]=encode(value[key],seen);return result;\n };\n const envSnapshot=env=>{const effective={};for(const key in env)effective[key]=env[key];return{own:encode(env),effective:encode(effective)};};\n const nodeSnapshot=n=>{\n  if(!n)return null;\n  const attrs={};if(n.attributes){for(const name of ['getAllAttributes','getAllInherited','getAllDefaults'])if(typeof n.attributes[name]==='function')attrs[name]=encode(n.attributes[name]());}\n  return{id:id(n),kind:n.kind,text:typeof n.getText==='function'?n.getText():null,attributes:attrs,properties:typeof n.getAllProperties==='function'?encode(n.getAllProperties()):null,children:(n.childNodes||[]).map(nodeSnapshot)};\n };\n const parserSnapshot=p=>({id:id(p),source:p.string,cursorUTF16:p.i,remaining:p.string.slice(p.i),currentCS:p.currentCS,macroCount:p.macroCount,env:envSnapshot(p.stack.env),stack:p.stack.stack.map(item=>({id:id(item),kind:item.kind,name:typeof item.getName==='function'?item.getName():null,nodes:(item.nodes||[]).map(nodeSnapshot)}))});\n const range=D.getRange;\n D.getRange=function(char){const result=range.apply(this,arguments);if(activeParser)events.push({event:'range',character:char,range:encode(result),parser:parserSnapshot(activeParser)});return result;};\n const install=p=>{\n  const config=p.configuration,handler=config.handlers.get('character');\n  if(!wrappedConfigurations.has(config)){\n   wrappedConfigurations.add(config);const addNode=config.addNode;\n   config.addNode=function(name,node){const event={event:'addNode',list:name,nodeBefore:nodeSnapshot(node),parser:activeParser?parserSnapshot(activeParser):null};events.push(event);const result=addNode.apply(this,arguments);event.nodeAfter=nodeSnapshot(node);return result;};\n  }\n  for(const {item:map,priority}of handler._configuration){\n   if(wrappedMaps.has(map))continue;wrappedMaps.add(map);const parse=map.parse;\n   map.parse=function(input){const [parser,char]=input;const event={event:'mapAttempt',map:this.name,priority,character:char,before:parserSnapshot(parser)};events.push(event);try{const result=parse.apply(this,arguments);event.result=encode(result);event.after=parserSnapshot(parser);return result;}catch(error){event.error={name:error.name,id:error.id||null,message:String(error.message)};throw error;}};\n  }\n  for(const item of handler._fallback){\n   if(wrappedFallbacks.has(item))continue;wrappedFallbacks.add(item);const fallback=item.item;\n   item.item=function(parser,char){const event={event:'fallback',priority:item.priority,character:char,before:parserSnapshot(parser)};events.push(event);try{const result=fallback.apply(this,arguments);event.result=encode(result);event.after=parserSnapshot(parser);return result;}catch(error){event.error={name:error.name,id:error.id||null,message:String(error.message)};throw error;}};\n  }\n  return handler;\n };\n const parse=T.parse;\n T.parse=function(kind,input){\n  if(kind!=='character')return parse.apply(this,arguments);\n  const handler=install(this),event={event:'characterDispatch',character:input[1],before:parserSnapshot(this),maps:Array.from(handler._configuration,({item,priority})=>({name:item.name,priority,pattern:item._regExp?String(item._regExp):null,parser:encode(item.parser)})),fallbacks:Array.from(handler._fallback,x=>({priority:x.priority}))};events.push(event);\n  const old=activeParser;activeParser=this;\n  try{const result=parse.apply(this,arguments);event.result=encode(result);event.after=parserSnapshot(this);return result;}catch(error){event.error={name:error.name,id:error.id||null,message:String(error.message)};throw error;}finally{activeParser=old;}\n };\n const create=T.create;\n T.create=function(kind,...args){\n  if(kind!=='token')return create.apply(this,arguments);\n  const factory=this.configuration.nodeFactory;\n  const event={event:'tokenFactory',kind,arguments:encode(args),before:parserSnapshot(this),factoryConstructor:factory.constructor.name};events.push(event);\n  try{const result=create.apply(this,arguments);event.created=nodeSnapshot(result);event.after=parserSnapshot(this);return result;}catch(error){event.error={name:error.name,id:error.id||null,message:String(error.message)};throw error;}\n };\n const push=T.Push;\n T.Push=function(item){\n  if(!item||!item.attributes)return push.apply(this,arguments);\n  const event={event:'pushNode',nodeBefore:nodeSnapshot(item),before:parserSnapshot(this)};events.push(event);\n  try{const result=push.apply(this,arguments);event.nodeAfter=nodeSnapshot(item);event.after=parserSnapshot(this);return result;}catch(error){event.error={name:error.name,id:error.id||null,message:String(error.message)};throw error;}\n };\n globalThis.observerSchema={kind:'source-passive-state-v1',state:'full parser source/cursor/count, effective and own environment including RegExp/function values, stack owners and full retained node snapshots',node:'kind/text, explicit/inherited/default attributes, properties, child nodes',scope:'actual character map attempts/fallbacks, getRange calls during dispatch, token factory and Push, configuration.addNode',excluded:'JavaScript prototypes and arbitrary configuration internals are not serialized; no candidate output or runtime normalization'};\n})()\n";
const OBSERVER_HASHES = {"pilotBody": "8a5f727f721fd4610a5de3859bd00a6287d5310abe73ab1dd21dff363f76ac71", "rangeBody": "ad8bc2942cac0f09fa966d455d6e9cfac581a6de5fce30f8759cab0b19f96392"};
const PILOT_CONVERT = new vm.Script('({svg:adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}))})');
async function main() {
  const args=process.argv.slice(2); assert.equal(args.length,2,'Pass PINNED_ASSETS NEW_OUTPUT');
  const assets=path.resolve(args[0]),output=path.resolve(args[1]);
  assert.equal(sha(fs.readFileSync(process.execPath)),NODE_SHA,'Use the pinned Node executable');
  assert(process.execArgv.includes('--jitless'),'Use the recorded --jitless original profile');
  assert(!fs.existsSync(output),'Output directory must be new');
  assert.equal(sha(Buffer.from(PILOT_OBSERVER_BODY)),OBSERVER_HASHES.pilotBody);
  assert.equal(sha(Buffer.from(RANGE_OBSERVER_BODY)),OBSERVER_HASHES.rangeBody);
  const scripts=Object.entries(ASSETS).map(([name,digest])=>{
    const bytes=fs.readFileSync(path.join(assets,name));assert.equal(sha(bytes),digest,name);
    return new vm.Script(bytes.toString(),{filename:name});
  });
  const pilotObserver=new vm.Script(PILOT_OBSERVER_BODY);
  const rangeObserver=new vm.Script(RANGE_OBSERVER_BODY,{filename:'passive-observer180.js'});
  const bytes=FILES.map(name=>fs.readFileSync(path.join(__dirname,name)));
  const docs=bytes.map(b=>JSON.parse(zlib.gunzipSync(b)));
  assert.deepEqual(docs.map(d=>d.cases.length),[868,0,104,152,72,40]);
  const originals=new Map(),indices=new Set();
  for(const doc of docs.slice(0,3)) {
    assert.equal(doc.mathjaxGitCommit,PIN);assert.equal(doc.sourceUnionSHA256,UNION);
    for(const row of doc.cases) {
      assert.equal(typeof row.tex,'string');assert.equal(typeof row.display,'boolean');
      assert(!originals.has(key(row))&&!indices.has(row.sourceIndex));
      originals.set(key(row),row.original);indices.add(row.sourceIndex);
      assert.deepEqual(row.original,row.sourceRecord.original);
      assert.equal(row.tex,row.sourceRecord.tex);assert.equal(row.display,row.sourceRecord.display);
    }
  }
  assert.equal(originals.size,972);assert.deepEqual([...indices].sort((a,b)=>a-b),Array.from({length:972},(_,i)=>i));
  for(const doc of docs.slice(3,5)) { assert.equal(doc.sourcePin,PIN);assert.deepEqual(doc.assets,ASSETS); }
  const passiveKeys=new Set();
  for(const doc of docs.slice(3,5)) for(const row of doc.cases) {
    const request=row.request||row;assert(!passiveKeys.has(key(request)));passiveKeys.add(key(request));
    assert.deepEqual(row.original,originals.get(key(request)));assert.deepEqual(row.observed.output,row.original);
  }
  assert.equal(passiveKeys.size,224);
  const supplement=docs[5],supplementIdentities=new Set(),supplementProfiles={};
  assert.equal(supplement.mathjaxGitCommit,PIN);
  assert.equal(supplement.actualMergedBase,'6d88373dd7479949265afdeae982b1bd5fcd5613');
  assert.equal(supplement.partition,'strict-svg-supplement');
  for(const row of supplement.cases) {
    const member=row.sourceMembership,record=member.record,binding=row.sourceGitBinding;
    assert.equal(typeof row.tex,'string');assert.equal(typeof row.display,'boolean');
    assert(!originals.has(key(row)));originals.set(key(row),row.original);
    const identity=JSON.stringify([member.sourcePath,member.jsonPointer]);
    assert(!supplementIdentities.has(identity));supplementIdentities.add(identity);
    assert.equal(record.tex,row.tex);assert.equal(record.display,row.display);
    assert.equal(binding.head,supplement.actualMergedBase);assert.equal(binding.path,member.sourcePath);
    assert.equal(binding.gitBlob,member.gitBlob);assert.equal(binding.sha256,member.sourceSHA256);
    assert.equal(row.profile,member.profile);
    const index=Number(member.jsonPointer.replace(/^\/cases\//,''));
    if(row.profile==='raw-prime-residual') {
      assert.equal(member.sourcePath,'testdata/operator_primes_residuals.json');
      assert((index>=121&&index<=142)||(index>=189&&index<=204));
      assert.equal(binding.gitBlob,'8933dac9387a943943ef6eeaa67a441a19ccfdfa');
      assert.equal(binding.sha256,'4ff8888780f5f22057ce8d48b9a2e47a23b0890b15a9d1a3e68609fd7a2b3003');
      assert.deepEqual(row.original,record.original);assert.equal(row.qualifiedChange,'fix');
      assert.equal(row.publicationCategory,'firstStrictPromotionFromExistingRaw');
      assert.equal(row.expectedField,'sourceMembership.record.original.svg');
    } else {
      assert.equal(row.profile,'strict-phantom-control');
      assert.equal(member.sourcePath,'testdata/operator_primes_mathjax_3_2_2.json');
      assert([3589,3590].includes(index));
      assert.equal(binding.gitBlob,'e8d161e4e38aabe67fdd921c709bc316a7dd573b');
      assert.equal(binding.sha256,'2261e9ddd241d11005c4706be360e756d8c8ed867036baee75655ec98ec39b17');
      assert.deepEqual(row.original,{svg:record.svg});assert.equal(row.qualifiedChange,'control');
      assert.equal(row.publicationCategory,'priorStrictCompleteSVG');
      assert.equal(row.expectedField,'sourceMembership.record.svg');
    }
    assert.deepEqual(Object.keys(row.original),['svg']);assert(row.original.svg);
    supplementProfiles[row.profile]=(supplementProfiles[row.profile]||0)+1;
  }
  assert.deepEqual(supplementProfiles,{'raw-prime-residual':38,'strict-phantom-control':2});
  assert.equal(originals.size,1012);

  fs.mkdirSync(output,{recursive:true});
  let conversions=0;
  const runtimeRepeats=[];
  const observationPath=path.join(output,'fresh-observations.jsonl');
  fs.writeFileSync(observationPath,'',{flag:'wx'});
  let observationCount=0,peakHeapUsed=0,peakRSS=0;
  function recordFresh(entry) {
    // Persist complete detached outputs before yielding or starting another VM.
    // A fatal V8 abort cannot run catch/finally, so evidence is streamed eagerly.
    fs.appendFileSync(observationPath,JSON.stringify(entry)+'\n');observationCount++;
    const usage=process.memoryUsage();peakHeapUsed=Math.max(peakHeapUsed,usage.heapUsed);peakRSS=Math.max(peakRSS,usage.rss);
  }
  async function afterConversion(convert) {
    const result=convert();
    // Conversion and observation stay synchronous. Drain pending startup promise
    // jobs only after their complete result has been copied and recorded, so the
    // previous context is collectible before the next fresh context is created.
    await new Promise(resolve=>setImmediate(resolve));
    return result;
  }
  async function sequentialMap(rows,convert) {
    const results=[];
    for(const row of rows)results.push(await convert(row));
    return results;
  }
  function context(request) {
    const c=vm.createContext({console,request,events:[]});c.globalThis=c;
    for(const script of scripts)script.runInContext(c,{timeout:10000});
    return c;
  }
  function original(request) {
    const c=context(request);conversions++;
    let result;
    try { const node=c.html.convert(request.tex,{display:request.display,em:16,ex:8});result={svg:c.adaptor.innerHTML(node)}; }
    catch(error) { result={error:String(error?.stack??error)}; }
    recordFresh({profile:'plain-original',request:{tex:request.tex,display:request.display},result});
    return result;
  }
  function observe(request,profile,enabled) {
    const c=context(request);
    if(enabled)(profile==='pilot'?pilotObserver:rangeObserver).runInContext(c,{timeout:10000});
    conversions++;
    if(profile==='pilot') {
      const output=PILOT_CONVERT.runInContext(c,{timeout:10000});
      const result=clone({output,events:c.events,dispatch:c.dispatch});
      recordFresh({profile,enabled,request,result});return result;
    }
    const node=c.html.convert(request.tex,{em:16,ex:8,display:request.display});
    const result=clone({output:{svg:c.adaptor.innerHTML(node)},events:c.events,schema:c.observerSchema||null});
    recordFresh({profile,enabled,request,result});return result;
  }
  try {
    const fresh=[];
    for(const doc of docs.slice(0,3)) {
      const cases=await sequentialMap(doc.cases,async row=>{
        const result=await afterConversion(()=>original(row));
        if(row.original.svg) {
          assert.deepEqual(result,row.original);
          return {...row,original:result};
        }
        // Full historical exceptions remain unchanged. Record the complete fresh
        // object before comparing source frames, so changed host frames are visible.
        const historical=row.original,entry={sourceIndex:row.sourceIndex,tex:row.tex,display:row.display,historical,fresh:result,fullObjectEqual:JSON.stringify(result)===JSON.stringify(historical)};
        runtimeRepeats.push(entry);
        fs.writeFileSync(path.join(output,'runtime-'+row.sourceIndex+'.json'),JSON.stringify(entry,null,2)+'\n',{flag:'wx'});
        assert.deepEqual(Object.keys(result),['error']);assert.deepEqual(Object.keys(historical),['error']);
        assert.equal(result.error.split('\n')[0],row.contract.originalFirstLine);
        const frames=s=>s.split('\n').filter(line=>line.includes('mathjax.js:'));
        assert.deepEqual(frames(result.error),frames(historical.error),'Original source frames differ at '+row.sourceIndex);
        assert.equal(frames(historical.error).length,10,'Unexpected historical runtime stack profile');
        return row;
      });
      fresh.push({...doc,cases});
    }
    const pilot=await sequentialMap(docs[3].cases,async row=>{
      const {original:historical,observed:retained,...request}=row;
      const plain=await afterConversion(()=>observe(request,'pilot',false));
      const observed=await afterConversion(()=>observe(request,'pilot',true));
      assert.deepEqual(plain,{output:historical,events:[]});assert.deepEqual(observed,retained);
      return {...request,original:plain.output,observed};
    });
    fresh.push({...docs[3],cases:pilot});
    const range=await sequentialMap(docs[4].cases,async row=>{
      const plain=await afterConversion(()=>observe(row.request,'range',false));
      const observed=await afterConversion(()=>observe(row.request,'range',true));
      assert.deepEqual(plain,row.plain);assert.deepEqual(observed,row.observed);assert.deepEqual(plain.output,row.original);
      return {...row,original:plain.output,plain,observed};
    });
    fresh.push({...docs[4],cases:range});
    const primes=await sequentialMap(docs[5].cases,async row=>{
      const result=await afterConversion(()=>original(row));assert.deepEqual(result,row.original);
      // Keep the entire historical source record, including old Go observations,
      // unchanged. Only the original object supplies the expected SVG.
      return {...row,original:result};
    });
    fresh.push({...docs[5],cases:primes});
    const files={};
    for(let i=0;i<FILES.length;i++) {
      assert.deepEqual(fresh[i],docs[i]);const regenerated=encode(fresh[i]);
      assert(regenerated.equals(bytes[i]),'Regenerated container differs: '+FILES[i]);
      fs.writeFileSync(path.join(output,FILES[i]),regenerated,{flag:'wx'});
      files[FILES[i]]={sha256:sha(regenerated),byteIdentical:true,completeObjectsEqual:true};
    }
    assert.equal(conversions,1460);assert.equal(observationCount,1460);assert.equal(runtimeRepeats.length,104);
    const receipt={status:'PASS original-only complete-object and deterministic regeneration',sourcePin:PIN,sourceUnionSHA256:UNION,
      assets:ASSETS,node:process.version,nodeSHA256:NODE_SHA,generatorSHA256:sha(fs.readFileSync(__filename)),
      originalSVGs:908,focusedOriginalSVGs:868,supplementalOriginalSVGs:40,heldSVGs:0,historicalRuntimeObjects:104,completeFreshRuntimeObjects:104,
      completeRuntimeObjectExact:runtimeRepeats.filter(r=>r.fullObjectEqual).length,
      passiveExistingInputs:224,passiveFreshConversions:448,conversions,files,noGo:true,
      completeFreshObservations:{path:observationPath,records:observationCount,sha256:sha(fs.readFileSync(observationPath)),framing:'JSON objects, one literal LF after each'},
      resourceProfile:{eventLoopYieldAfterEachDetachedConversion:true,heapLimitChangedByGenerator:false,peakHeapUsed,peakRSS},
      historicalExceptionObjectsUnchanged:true,completeSupplementSourceRecordsUnchanged:true,observerBodySHA256:OBSERVER_HASHES};
    fs.writeFileSync(path.join(output,'unicode_character_owner_regeneration.json'),JSON.stringify(receipt,null,2)+'\n',{flag:'wx'});
    console.log(JSON.stringify(receipt));
  } catch(error) {
    fs.writeFileSync(path.join(output,'failed-regeneration.json'),JSON.stringify({conversions,name:error.name,message:error.message,stack:error.stack,completeRuntimeRepeats:runtimeRepeats,completeFreshObservations:{path:observationPath,records:observationCount,sha256:sha(fs.readFileSync(observationPath))},resourceProfile:{peakHeapUsed,peakRSS}},null,2)+'\n',{flag:'wx'});
    throw error;
  }
}
module.exports={encode,PILOT_OBSERVER_BODY,RANGE_OBSERVER_BODY,OBSERVER_HASHES,FILES};
if(require.main===module)main().catch(error=>{console.error(error);process.exitCode=1;});
