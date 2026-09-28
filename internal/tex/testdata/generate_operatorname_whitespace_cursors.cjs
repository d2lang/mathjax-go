const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),crypto=require('node:crypto');
const assets=process.argv[2]; if(!assets)throw Error('expected pinned D2 assets');
const hashes={'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01','mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869','setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'};
const scripts=Object.entries(hashes).map(([n,h])=>{const b=fs.readFileSync(path.join(assets,n));if(crypto.createHash('sha256').update(b).digest('hex')!==h)throw Error(n);return new vm.Script(b.toString(),{filename:n})});
const cases=[['single-letter','{a} tail'],['single-multi-letter','{arg} tail'],['split-single','{a b} tail'],['split-multi','{arg max} tail'],['explicit-space','{arg\\,max} tail'],['empty','{} tail'],['starred','*{a b}_i^n'],['missing-argument','']].map(([name,source])=>({name,source}));
for(const space of [' ','\t','\ufeff','\x85','\u2028','\u2029'])for(const tail of ['x','\\limits_i x','\\nolimits_i x','\\alpha'])cases.push({name:'following-'+cases.length,source:'{lim}'+space+tail});
const results=[];let methods;
for(const request of cases){
 const tex='\\operatorname'+request.source;let observed,plain;
 for(const observe of [false,true]){
  const c=vm.createContext({console,tex,observed:[],argumentEvents:[]});c.globalThis=c;for(const s of scripts)s.runInContext(c);
  if(observe)vm.runInContext(`(()=>{const texmod=MathJax._.input.tex;const P=texmod.TexParser.default.prototype;const push=P.Push,argument=P.GetArgument;
   methods={HandleOperatorName:texmod.ams.AmsMethods.AmsMethods.HandleOperatorName.toString(),GetNext:P.GetNext.toString()};
   P.GetArgument=function(...args){try{return argument.apply(this,args)}finally{if(this.string===tex)argumentEvents.push({cursor:this.i,remaining:this.string.slice(this.i)});}};
   P.Push=function(node){if(this.string===tex&&node&&node.properties&&Object.prototype.hasOwnProperty.call(node.properties,'movesupsub'))observed.push({cursor:this.i,remaining:this.string.slice(this.i),kind:node.kind,star:node.properties.movesupsub});return push.call(this,node);};
  })()`,c);
  let output;try{output=vm.runInContext(`({svg:adaptor.innerHTML(html.convert(tex,{display:false,em:16,ex:8}))})`,c)}catch(e){output={error:e.stack};}
  if(!observe)plain=output;else{observed={output,events:c.observed,argumentEvents:c.argumentEvents};methods=c.methods;}
 }
 if(JSON.stringify(plain)!==JSON.stringify(observed.output))throw Error('observer changed original '+request.name);
 if(request.name!=='missing-argument'&&observed.events.length!==1)throw Error(JSON.stringify({request,events:observed.events}));
 const event=observed.events[0]||observed.argumentEvents[0];if(!event)throw Error('missing cursor '+request.name);
 results.push({...request,cursorCodeUnits:event.cursor-13,consumedBytes:Buffer.byteLength(request.source.slice(0,event.cursor-13)),remaining:event.remaining,events:observed.events,argumentEvents:observed.argumentEvents,original:plain});
}
fs.writeFileSync(path.join(__dirname,'operatorname_whitespace_cursors.json'),JSON.stringify({mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,methods,cases:results},null,2)+'\n');console.log(results.map(({name,cursorCodeUnits,consumedBytes,remaining})=>({name,cursorCodeUnits,consumedBytes,remaining})));
