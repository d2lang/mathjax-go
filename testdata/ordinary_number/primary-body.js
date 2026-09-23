(()=>{
 const jax=html.inputJax[0], config=jax.parseOptions, handlers=jax.configuration.handlers, T=MathJax._.input.tex.TexParser.default;
 const methods=MathJax._.input.tex.ParseMethods.default, Other=MathJax._.input.tex.base.BaseConfiguration.Other;
 const digit=handlers.retrieve('digit'),fallback=Array.from(handlers.get('character')._fallback)[0];
 if(digit.parser!==methods.digit||fallback.item!==Other)throw Error('actual registered numeric route mismatch');
 const copy=x=>JSON.parse(JSON.stringify(x,(_,v)=>v===undefined?{__js_undefined:true}:v));
 const pairs=o=>Object.keys(o).map(name=>({name,value:copy(o[name])}));
 const full=n=>n?{kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),children:n.childNodes.map(full)}:null;
 const ids=new WeakMap();let serial=0;const id=o=>o&&typeof o==='object'?(ids.has(o)?ids.get(o):(ids.set(o,++serial),serial)):null;
 const error=e=>({id:e.id||null,message:e.message||String(e),type:e.constructor?.name||null});
 const events=[],created=[],pushes=[];let depth=0;
 const state=p=>copy({source:p.string,cursorUTF16:p.i,remaining:p.string.slice(p.i),environment:p.stack.env,environmentIdentity:id(p.stack.env),configurationIdentity:id(p.configuration),macroCount:p.macroCount});
 if(request.registration){const r=request.registration;const {Macro}=MathJax._.input.tex.Symbol;handlers.retrieve('ams-declare-ops').add(r.name,new Macro(r.name,MathJax._.input.tex.base.BaseMethods.default.Macro,[r.body,r.arguments]));}
 const factory=config.nodeFactory,create=factory.create,push=T.prototype.Push,parse=T.prototype.parse;let parserCompletion=null,parserError=null;
 factory.create=function(...args){const n=Reflect.apply(create,this,args);if(passive&&depth)created.push(copy({identity:id(n),requestedFactory:args[0],requestedKind:typeof args[1]==='string'?args[1]:null,children:n.childNodes?.map(id)||[],parent:id(n.parent),full:full(n)}));return n};
 T.prototype.Push=function(...args){if(passive&&depth)pushes.push(copy({nodes:args.filter(n=>n?.kind).map(n=>({identity:id(n),parent:id(n.parent),full:full(n)}))}));try{const r=Reflect.apply(push,this,args);if(args[0]?.kind==='stop')parserCompletion=state(this);return r}catch(e){parserError={...error(e),...state(this)};throw e}};
 T.prototype.parse=function(...args){try{return Reflect.apply(parse,this,args)}catch(e){parserError={...error(e),...state(this)};throw e}};
 const observe=(name,original)=>function(p,c){const event={handler:name,character:c,before:state(p),creationStart:created.length,pushStart:pushes.length};if(passive)events.push(event);depth++;try{return Reflect.apply(original,this,[p,c])}catch(e){event.error=error(e);throw e}finally{depth--;event.after=state(p);event.creationEnd=created.length;event.pushEnd=pushes.length}};
 if(passive){digit._parser=observe('digit',methods.digit);fallback.item=observe('Other',Other);}
 let tree=null,after=null,root=null,formattedError=null,thrown=null,svg=null;
 const output=html.outputJax.typeset,format=jax.formatError;
 html.outputJax.typeset=function(m,d){root=m.root;tree=copy(full(root));const r=Reflect.apply(output,this,[m,d]);after=copy(full(root));return r};
 jax.formatError=function(e){formattedError=error(e);return Reflect.apply(format,this,[e])};
 try{svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}))}catch(e){thrown=error(e);after=copy(full(root))}
 return copy({request,passive,svg,tree,after,formattedError,thrown,parserError,parserCompletion,events,created,pushes,methodSources:{digit:methods.digit.toString(),Other:Other.toString(),digits:String(config.options.digits)}});
})()
