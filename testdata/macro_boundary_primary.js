(()=>{
 const T=MathJax._.input.tex.TexParser.default, U=MathJax._.input.tex.ParseUtil.default, B=MathJax._.input.tex.base.BaseMethods.default;
 const N=MathJax._.input.tex.newcommand?.NewcommandMethods.default;
 const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));
 const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),children:n.childNodes.map(full)});
 const err=e=>({id:e.id||null,message:e.message||String(e),type:e.constructor?.name||null});
 const ids=new WeakMap();let next=0;const id=o=>{if(!ids.has(o))ids.set(o,++next);return ids.get(o)};
 const parsers=[],events=[]; let active=null;
 const state=p=>({parserID:id(p),source:p.string,cursorUTF16:p.i,remaining:p.string.slice(p.i),macroCount:p.macroCount,currentCS:p.currentCS,env:{...p.stack?.env},stack:p.stack?.stack?.map(x=>({id:id(x),kind:x.kind,nodeKinds:x.nodes?.map(n=>n.kind)}))||[],parserStack:p.configuration.parsers?.map(x=>id(x))||null});
 const summary=x=>x==null||typeof x!=='object'?x:Array.isArray(x)?x.map(summary):{id:id(x),kind:x.kind||null};
 function wrap(obj,key,parserArg){const original=obj[key];if(typeof original!=='function')throw Error('missing method '+key);obj[key]=function(...args){const p=parserArg?args[0]:this;const e={method:key,before:state(p),arguments:(parserArg?args.slice(1):args).map(summary)};events.push(e);try{const r=original.apply(this,args);e.returnValue=summary(r);return r}catch(x){e.error=err(x);throw x}finally{e.after=state(p)}};}
 if(passive){
  const parse=T.prototype.Parse;T.prototype.Parse=function(...args){if(!parsers.includes(this))parsers.push(this);return parse.apply(this,args)};
  for(const k of ['GetArgument','GetBrackets','GetNext','GetCS','Push','create'])wrap(T.prototype,k,false);
  for(const k of ['addArgs','substituteArgs','checkMaxMacros'])wrap(U,k,true);
  if(N)for(const k of ['MacroWithTemplate','BeginEnv'])wrap(N,k,true);
 }
 function register(c){const {Macro}=MathJax._.input.tex.Symbol;for(const r of c.registrations||[]){const original=r.handler==='MacroWithTemplate'?N.MacroWithTemplate:B.Macro;const fn=function(p,...args){if(c.initialCount!==undefined&&!active.seeded){p.macroCount=c.initialCount;active.seeded=true}const e={method:'registered '+(r.handler||'Macro'),before:state(p),arguments:args};if(passive)events.push(e);try{return original.call(this,p,...args)}catch(x){e.error=err(x);throw x}finally{e.after=state(p)}};const args=r.handler==='MacroWithTemplate'?[r.body,String(r.arguments),...(r.parameters||[])]:[r.body,r.arguments,...(Object.hasOwn(r,'optionalDefault')?[r.optionalDefault]:[])];html.inputJax[0].configuration.handlers.retrieve('ams-declare-ops').add(r.name,new Macro(r.name,fn,args));}}
 function convert(c){active={seeded:false};events.length=0;parsers.length=0;register(c);const jax=html.inputJax[0];Object.assign(jax.parseOptions.options,c.options||{});let tree=null,svg=null,error=null;const formatted=[];const originalFormat=jax.formatError;jax.formatError=function(e){formatted.push(err(e));return originalFormat.call(this,e)};const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){tree=JSON.parse(JSON.stringify(full(math.root)));return original.call(this,math,doc)};
  try{svg=adaptor.innerHTML(html.convert(c.tex,{display:c.display,em:16,ex:8}));}catch(e){error=err(e)}finally{html.outputJax.typeset=original;jax.formatError=originalFormat}
  return {name:c.name,svg,tree,error,formattedErrors:formatted,trace:passive?{events:JSON.parse(JSON.stringify(events)),finalParsers:parsers.map(state),documentID:id(html),options:{maxBuffer:jax.parseOptions.options.maxBuffer,maxMacros:jax.parseOptions.options.maxMacros}}:null};
 }
 if(request.type==='helper'){
  const p=new T('',{},html.inputJax[0].parseOptions);p.mml();if(request.maxBuffer!==undefined)p.options.maxBuffer=request.maxBuffer;
  events.length=0;const before=state(p);let result=null,error=null;
  try{result=request.body!==undefined?U.substituteArgs(p,request.args,request.body):U.addArgs(p,request.s1??request.s1Repeat.text.repeat(request.s1Repeat.count),request.s2)}catch(e){error=err(e)}
  return{result,error,before,after:state(p),options:{maxBuffer:p.options.maxBuffer},trace:passive?JSON.parse(JSON.stringify(events)):null};
 }
 return request.type==='sequence'?request.cases.map(convert):convert(request);
})()
