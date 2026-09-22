// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Explicit limits public contract, including primary rejection behavior; no product changes.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_limits_filter.cjs /path/to/pinned/d2latex [evidence-directory]"
  );
if (evidence) fs.mkdirSync(evidence, { recursive: true });
const hash = (x) => crypto.createHash("sha256").update(x).digest("hex");
const hashes = {
  "polyfills.js":
    "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01",
  "mathjax.js":
    "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869",
  "setup.js":
    "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881",
};
const assets = Object.entries(hashes).map(([f, sha]) => {
  const b = fs.readFileSync(path.join(base, f));
  if (hash(b) !== sha) throw Error("Unpinned " + f);
  return [f, b];
});
const c=vm.createContext({console});for(const[f,b]of assets)vm.runInContext(b.toString(),c,{filename:f});

const outputs=[];
for (const kinds of [['munderover','munderover'],['munder','munder'],['mover','mover'],['munderover','munder'],['munder','mover'],['mover','munderover']]) for(const outerFirst of [false,true]) {
c.spec={kinds,outerFirst};
const result=vm.runInContext(`(()=>{
const options=html.inputJax[0].parseOptions; options.clear();
const nf=options.nodeFactory, originalCreate=nf.create.bind(nf), made=[];
const create=(kind,children=[],def={})=>originalCreate('node',kind,children,def);
const token=t=>originalCreate('token','mi',{},t);
const mark=k=>k==='munderover'?[token('i'),token('n')]:[token('i')];
const base=token('x'); base.setProperty('movablelimits',true);
const build=(kind,base)=>create(kind,[base,...mark(kind)],{displaystyle:false,movablelimits:true,custom:'retained'});
let inner,outer;
if(spec.outerFirst){outer=create(spec.kinds[0],[],{displaystyle:false,movablelimits:true,custom:'retained'});inner=build(spec.kinds[1],base);outer.setChildren([inner,...mark(spec.kinds[0])]);}
else {inner=build(spec.kinds[1],base);outer=build(spec.kinds[0],inner);}
for(const n of [inner,outer]) {n.setProperty('movesupsub',true);n.setProperty('texprimestyle',true);n.setProperty('unknownCopyProbe','kept');}
const detached=build('mover',token('q')); const root=create('math',[outer]);options.root=root;
const records=[['outer',outer],['inner',inner],['base',base],['detached',detached]],ids=new Map(records.map(([id,n])=>[n,id]));let next=0;
function id(n){if(!n)return null;if(!ids.has(n))ids.set(n,'new'+(++next));return ids.get(n)}
const before=records.map(([name,n])=>({name,kind:n.kind,children:n.childNodes.map(id),parent:id(n.parent)}));const trace=[];
nf.create=function(...args){const n=originalCreate(...args);if(args[0]==='node'){trace.push({kind:args[1],children:(args[2]||[]).map(id),id:id(n)});made.push(n)}return n};
MathJax._.input.tex.FilterUtil.default.moveLimits({data:options});nf.create=originalCreate;
const tree=n=>({kind:n.kind,id:id(n),attributes:n.attributes?.getAllAttributes()||{},properties:n.getAllProperties(),children:n.childNodes.map(tree)});
const after=records.map(([name,n])=>{const replacement=made.find(m=>m.attributes===n.attributes);return{name,kind:n.kind,parent:id(n.parent),children:n.childNodes.map(id),replacement:replacement?{kind:replacement.kind,id:id(replacement),attributesShared:true,propertyObjectShared:replacement.properties===n.properties,properties:replacement.getAllProperties(),children:replacement.childNodes.map(id),childParents:replacement.childNodes.map(c=>id(c.parent))}:null}});
return{spec,before,trace,after,tree:tree(options.root),lists:Object.fromEntries(['munderover','munder','mover'].map(k=>[k,options.getList(k).map(id)]))};
})()`,c);outputs.push(result);
}
const predicates=[];
for(const spec of [
{name:'inline-own-true',own:true}, {name:'display-true',own:true,display:true},
{name:'display-inherited-true',own:true,displayInherited:true}, {name:'display-explicit-false-wins',own:true,display:false,displayInherited:true},
{name:'missing-own'}, {name:'false-own',own:false}, {name:'null-own',own:null}, {name:'zero-own',own:0}, {name:'empty-own',own:''}, {name:'string-own',own:'false'},
{name:'base-attribute-only',baseAttribute:true}, {name:'explicit-core-true',own:true,explicit:true}, {name:'explicit-core-false',own:true,explicit:false},
{name:'inherited-core-true',own:true,inherited:true}, {name:'default-core-true',own:true,coreKind:'mo'},
{name:'wrapped-core-true',own:true,wrapped:true,explicit:true}, {name:'wrapped-core-false',own:true,wrapped:true,explicit:false},
{name:'core-own-only',wrapped:true,coreOwn:true}, {name:'detached',own:true,detached:true}, {name:'display-truthy-string',own:true,display:'false'}
]){
c.spec=spec;
predicates.push(vm.runInContext(`(()=>{
const o=html.inputJax[0].parseOptions;o.clear();const nf=o.nodeFactory;
const n=(k,ch=[],a={})=>nf.create('node',k,ch,a), tok=(k,t)=>nf.create('token',k,{},t);
const core=tok(spec.coreKind||'mi',spec.coreKind==='mo'?'∑':'x');
const base=spec.wrapped?n('mstyle',[core]):core;
if(Object.hasOwn(spec,'own'))base.setProperty('movablelimits',spec.own);
if(spec.coreOwn)core.setProperty('movablelimits',true);
if(spec.baseAttribute)base.attributes.set('movablelimits',true);
if(Object.hasOwn(spec,'explicit'))core.attributes.set('movablelimits',spec.explicit);
if(Object.hasOwn(spec,'inherited'))core.attributes.setInherited('movablelimits',spec.inherited);
const mark=tok('mi','a'),old=n('mover',[base,mark],{custom:'kept'});
if(Object.hasOwn(spec,'display'))old.attributes.set('displaystyle',spec.display);
if(Object.hasOwn(spec,'displayInherited'))old.attributes.setInherited('displaystyle',spec.displayInherited);
const root=n('math',[spec.detached?tok('mi','z'):old]);o.root=root;
MathJax._.input.tex.FilterUtil.default.moveLimits({data:o});
const now=spec.detached?old:root.childNodes[0].childNodes[0],changed=now!==old;
return {spec,changed,kind:now.kind,retained:o.getList('mover').includes(old),childrenSame:now.childNodes[0]===base&&now.childNodes[1]===mark,attributesShared:now.attributes===old.attributes,oldDetached:old.parent===null,childParent:base.parent===now,propertiesShared:now.properties===old.properties};
})()`,c));
}
fs.writeFileSync(path.join(evidence,'limits_filter_mathjax_3_2_2.json'),JSON.stringify({mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,outputs,predicates},null,2)+'\n');console.log(outputs.length,predicates.length);
