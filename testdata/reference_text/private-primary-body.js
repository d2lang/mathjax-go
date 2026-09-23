(()=>{
 const config=html.inputJax[0].parseOptions,tags=config.tags;
 const methods=MathJax._.input.tex.base.BaseMethods.default,util=MathJax._.input.tex.ParseUtil.default;
 const {Macro}=MathJax._.input.tex.Symbol,{Label}=MathJax._.input.tex.Tags;
 const copy=x=>JSON.parse(JSON.stringify(x,(_,v)=>v===undefined?{__js_undefined:true}:v));
 const pairs=o=>Object.keys(o).map(name=>({name,value:copy(o[name])}));
 const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),children:n.childNodes.map(full)});
 const ids=new WeakMap();let serial=0;const id=o=>o&&typeof o==='object'?(ids.has(o)?ids.get(o):(ids.set(o,++serial),serial)):null;
 const labels=o=>Object.keys(o).map(name=>({name,identity:id(o[name]),tag:o[name].tag,id:o[name].id}));
 const state=()=>copy({tags:id(tags),configuration:id(config),factory:id(config.nodeFactory),currentTag:{identity:id(tags.currentTag),...tags.currentTag},labelsIdentity:id(tags.labels),labels:labels(tags.labels),allLabelsIdentity:id(tags.allLabels),allLabels:labels(tags.allLabels),stackIdentity:id(tags.stack),stack:tags.stack.map(id),historyIdentity:id(tags.history),history:tags.history.map(id),redo:tags.redo,refUpdate:tags.refUpdate,counter:tags.counter,allCounter:tags.allCounter});
 if(request.operation==='sequence'){
  const stages=[];for(const tex of request.sources){let tree=null,formattedError=null,thrown=null;const output=html.outputJax.typeset,format=html.inputJax[0].formatError;html.outputJax.typeset=function(m,d){tree=copy(full(m.root));return Reflect.apply(output,this,[m,d])};html.inputJax[0].formatError=function(e){formattedError={id:e.id||null,message:e.message};return Reflect.apply(format,this,[e])};const before=state();let svg=null;try{svg=adaptor.innerHTML(html.convert(tex,{display:request.display,em:16,ex:8}))}catch(e){thrown={id:e.id||null,message:e.message||String(e)}}finally{html.outputJax.typeset=output;html.inputJax[0].formatError=format}stages.push({tex,before,after:state(),tree,svg,formattedError,thrown});}return{input:request,stages};
 }
 if(request.registration){const r=request.registration;html.inputJax[0].configuration.handlers.retrieve('ams-declare-ops').add(r.name,new Macro(r.name,methods.Macro,[r.body,r.arguments]));}
 if(request.baseURL!==undefined)config.options.baseURL=request.baseURL;
 const observations=[];let tree=null,formattedError=null,thrown=null;
 const observe=function(parser){
  if(!request.unknown)tags.labels.t=new Label(request.tag??'T',request.id??'mjx-eqn:t');
  if(request.secondary)tags.labels.u=new Label(request.secondary.tag,request.secondary.id);
  if(request.unrelated)tags.labels.other=new Label(request.unrelated.tag,request.unrelated.id);
  if(request.allLabel)tags.allLabels.t=new Label(request.allLabel.tag,request.allLabel.id);
  if(request.refUpdate!==undefined)tags.refUpdate=request.refUpdate;
  if(request.font)parser.stack.env.font=request.font;
  parser.macroCount=request.macroCount??999;
  const selected=tags.allLabels.t||tags.labels.t,current=tags.currentTag,environment=parser.stack.env,labelsMap=tags.labels,factory=config.nodeFactory,color=config.packageData.get('color').model;
  const before=state(),envBefore=copy(environment),create=factory.create,push=parser.Push,oldHook=config.options.internalMath;
  const made=[],created=[],pushed=[],callback=[];
  factory.create=function(...args){const n=Reflect.apply(create,this,args);created.push(n);made.push({identity:id(n),requestKind:args[0],requestedType:typeof args[1]==='string'?args[1]:null,kind:n.kind,children:n.childNodes?.map(id)||[]});return n};
  parser.Push=function(...args){for(const n of args)if(n&&n.kind)pushed.push({identity:id(n),kind:n.kind,children:n.childNodes?.map(id)||[],full:copy(full(n))});return Reflect.apply(push,this,args)};
  if(request.callback)config.options.internalMath=function(p,text,level,font){const n=factory.create('token','mtext',{},'CALLBACK');callback.push({parserSame:p===parser,text,level:copy(level),font:copy(font),node:id(n)});if(request.callback==='mutate-selected-id')selected.id='mutated id';else tags.labels.t=new Label('REPLACED','replacement id');return[n]};
  let error=null;try{methods.HandleRef(parser,'\\'+(request.command||'ref'),!!request.eqref)}catch(e){error={id:e.id||null,message:e.message||String(e)};throw e}finally{
   factory.create=create;parser.Push=push;config.options.internalMath=oldHook;
   observations.push(copy({before,after:state(),error,made,createdGraph:created.map(n=>({identity:id(n),parent:id(n.parent),children:n.childNodes?.map(id)||[],full:n.kind?copy(full(n)):null})),parents:created.every(n=>!n.childNodes||n.childNodes.every(c=>c.parent===n)),pushed,callback,outer:{source:parser.string,cursorUTF16:parser.i,remaining:parser.string.slice(parser.i),macroCount:parser.macroCount,envBefore,envAfter:parser.stack.env,environmentIdentity:parser.stack.env===environment,configurationIdentity:parser.configuration===config,currentIdentity:tags.currentTag===current,labelsIdentity:tags.labels===labelsMap,factoryIdentity:config.nodeFactory===factory,colorIdentity:config.packageData.get('color').model===color,lateColor:color.getColor('named','late'),selectedReferenceIdentity:id(selected),selectedReferenceAfter:selected?{tag:selected.tag,id:selected.id}:null}}));
  }
 };
 html.inputJax[0].configuration.handlers.retrieve('ams-declare-ops').add('DRefObserve',new Macro('DRefObserve',observe,[]));
 const output=html.outputJax.typeset,format=html.inputJax[0].formatError;html.outputJax.typeset=function(m,d){tree=copy(full(m.root));return Reflect.apply(output,this,[m,d])};html.inputJax[0].formatError=function(e){formattedError={id:e.id||null,message:e.message};return Reflect.apply(format,this,[e])};
 let svg=null;try{svg=adaptor.innerHTML(html.convert(request.source||'\\DRefObserve{t}TAIL',{display:true,em:16,ex:8}))}catch(e){thrown={id:e.id||null,message:e.message||String(e)}}
 return copy({input:request,observations,tree,svg,formattedError,thrown,methodSources:{HandleRef:methods.HandleRef.toString(),internalMath:util.internalMath.toString()}});
})()
