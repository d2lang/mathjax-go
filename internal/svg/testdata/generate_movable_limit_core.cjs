// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Nested movable-limit renderer contract; no dispatcher or source overlays.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_nested_movable_limits.cjs /path/to/pinned/d2latex [evidence-directory]"
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

const t=text=>({kind:'text',text}),n=(kind,children=[],attributes={})=>({kind,children,attributes}),mi=(attrs={})=>n('mi',[t('x')],attrs),mo=()=>n('mo',[t('∑')],{movablelimits:true});
const inputs=[];for(const value of [true,false]){const a={movablelimits:value};for(const [name,spec] of [
['direct-mi',mi(a)],['direct-mn',n('mn',[t('1')],a)],['row',n('mrow',[mi(),mi()],a)],['one-mi-row',n('mrow',[mi(a)])],['script-mi',n('msub',[mi(a),mi()])],['script-row',n('msup',[n('mrow',[mi(),mi()],a),mi()])],['style-row',n('mstyle',[mi(),mi()],a)],['padded-row',n('mpadded',[n('mrow',[mi(),mi()],a)])],['TeXAtom-row',n('TeXAtom',[n('mrow',[mi(),mi()],a)])],['fraction-mi',n('mfrac',[mi(a),mi()])],['root-own',n('mroot',[mi(),mi()],a)],['sqrt-own',n('msqrt',[mi()],a)],['enclose-own',n('menclose',[mi()],a)],['maction-selected',n('maction',[mi(),mi(a)],{selection:2})],['row-core-index1',n('mrow',[n('mspace',[],{width:'1em'}),n('mo',[t('∑')],a)])]
])inputs.push([name+'-'+value,spec]);}inputs.push(['nested-mo',n('msubsup',[mo(),mi(),mi()])]);
const records=[];
for(const [label,spec]of inputs)for(const display of[false,true]){
const c=vm.createContext({console:{log(){},warn(){},error(){}}});for(const[f,b]of assets)vm.runInContext(b.toString(),c,{filename:f});c.request={spec,display};
const result=vm.runInContext(`(()=>{
const factory=html.inputJax[0].mmlFactory;const build=s=>s.kind==='text'?factory.create('text').setText(s.text):factory.create(s.kind,s.attributes||{},(s.children||[]).map(build));
const base=build(request.spec),root=factory.create('math',{display:request.display?'block':'inline'},[factory.create('munderover',{},[base,factory.create('mi',{},[factory.create('text').setText('b')]),factory.create('mi',{},[factory.create('text').setText('a')])])]);root.setInheritedAttributes({},request.display,0,false);root.setTeXclass(null);
const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.isEmbellished&&n.core()!==n?n.childNodes.indexOf(n.core()):(typeof n.coreIndex==='function'?n.coreIndex():n.coreIndex||0)},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});
const core=base.coreMO();let corePath=null;function seek(x,p=[]){if(x===core)corePath=p;x.childNodes.forEach((q,i)=>seek(q,p.concat(i)))}seek(base);
const fullTree=full(root);let wrapperValue=null;
const originalWrap=html.outputJax.factory.wrap;html.outputJax.factory.wrap=function(...args){const w=originalWrap.apply(this,args);if(w.node===root.childNodes[0].childNodes[0])wrapperValue=w.hasMovableLimits();return w};
const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){math.root=root;return original.call(this,math,doc)};
let svg=null,renderError=null;try{svg=adaptor.innerHTML(html.convert('x',{display:request.display,em:16,ex:8}))}catch(e){renderError=String(e)}return{fullTree,svg,renderError,corePath,coreKind:core.kind,coreAttribute:core.attributes.get('movablelimits')??null,expected:!request.display&&!!core.attributes.get('movablelimits'),wrapperValue};})()`,c);
const name=label+(display?'-display':'-inline');if(result.svg!==null)fs.writeFileSync(path.join(evidence,name+'.svg'),result.svg);records.push({name,display,...result});}
fs.writeFileSync(path.join(evidence,'boundaries.json'),JSON.stringify({assetsSHA256:hashes,cases:records},null,2));console.log(records.length+' boundary observations');
