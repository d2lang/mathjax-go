// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Registered MathML constructor observations; the original renderer and adaptor remain unmodified.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_accent_script_constructor.cjs /path/to/pinned/d2latex [evidence-directory]"
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

const text=s=>({kind:'text',text:s});
const n=(kind,children=[],attributes={},properties={})=>({kind,children,attributes,properties});
const tok=(kind,s,attributes={},properties={})=>n(kind,[text(s)],attributes,properties);
const x=()=>tok('mi','x');
const mark=(value=true,kind='mo')=>tok(kind,'^',{stretchy:false},{mathaccent:value});
const acc=(base=x(),script=mark(),kind='mover',attrs={})=>n(kind,[base,script],attrs);
const inputs=[];
for(const [label,value]of [['true',true],['false',false],['zero',0],['one',1],['empty',''],['false-string','false'],['zero-string','0'],['null',null]])inputs.push([label,acc(x(),mark(value))]);
inputs.push(['missing',acc(x(),tok('mo','^',{stretchy:false}))]);
inputs.push(['under',acc(x(),mark(), 'munder')]);
inputs.push(['three-under-true',n('munderover',[x(),mark(true),mark(false)])]);
inputs.push(['three-over-only',n('munderover',[x(),mark(false),mark(true)])]);
inputs.push(['digit',acc(tok('mn','1'))]);
inputs.push(['operator',acc(tok('mo','+',{stretchy:false}))]);
inputs.push(['large-operator',acc(tok('mo','∑',{largeop:true,movablelimits:false,stretchy:false}))]);
inputs.push(['fixed-operator',acc(tok('mo',')',{stretchy:true,minsize:'1.2em',maxsize:'1.2em'}))]);
inputs.push(['multi-character',acc(tok('mi','xy'))]);
inputs.push(['relative-scale',acc(tok('mi','x',{mathsize:'2em'}))]);
inputs.push(['transparent-style',acc(n('mstyle',[x()]))]);
inputs.push(['transparent-atom',acc(n('TeXAtom',[x()]))]);
inputs.push(['vcenter-atom',acc(n('TeXAtom',[x()],{}, {texClass:8}))]);
inputs.push(['non-mo-script',acc(x(),mark(true,'mi'))]);
inputs.push(['script-row-own',acc(x(),n('mrow',[x(),tok('mi','y')],{}, {mathaccent:true}))]);
inputs.push(['script-selected',acc(x(),n('maction',[mark(false),mark(true)],{selection:2}))]);
inputs.push(['outer-false-first',acc(acc(x(),mark(), 'mover',{accent:true}),mark(), 'mover',{accent:false})]);
inputs.push(['outer-under-false-first',acc(acc(x(),mark(), 'munder',{accentunder:true}),mark(), 'munder',{accentunder:false})]);
inputs.push(['wrapped-mathaccent',n('mpadded',[acc()])]);
inputs.push(['ordinary-overset',acc(x(),tok('mi','a'), 'mover',{accent:false})]);

for (const [label,value] of [['true',true],['false',false],['null',null],['empty',''],['whitespace',' '],['hex','0x2'],['padded',' 2 '],['number',2]]) {
 for (const first of [false,true]) inputs.push(['selection-'+label+'-'+(first?'first':'second'),acc(x(),n('maction',[mark(first),mark(!first)],{selection:value}))]);
}

const records=[];
for(const[name,spec]of inputs)for(const display of[false,true]){
 const c=vm.createContext({console:{log(){},warn(){},error(){}}});for(const[f,b]of assets)vm.runInContext(b.toString(),c,{filename:f});c.request={spec,display};
 const result=vm.runInContext(`(()=>{
 const factory=html.inputJax[0].mmlFactory;
 const build=s=>{const q=s.kind==='text'?factory.create('text').setText(s.text):factory.create(s.kind,s.attributes||{},(s.children||[]).map(build));for(const[k,v]of Object.entries(s.properties||{}))q.setProperty(k,v);return q};
 const base=build(request.spec),outer=factory.create('msup',{},[base,factory.create('mi',{},[factory.create('text').setText('n')])]),root=factory.create('math',{display:request.display?'block':'inline'},[outer]);
 root.setInheritedAttributes({},request.display,0,false);root.setTeXclass(null);
 const paths=new Map();const locate=(q,p=[])=>{paths.set(q,p);q.childNodes.forEach((c,i)=>locate(c,p.concat(i)))};locate(root);
 const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.isEmbellished&&n.core()!==n?n.childNodes.indexOf(n.core()):(typeof n.coreIndex==='function'?n.coreIndex():n.coreIndex||0)},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});
 const fullTree=full(root),observations=[];
 const originalWrap=html.outputJax.factory.wrap;html.outputJax.factory.wrap=function(...args){const w=originalWrap.apply(this,args);if(w.node.isKind('msubsup')||w.node.isKind('munderover'))observations.push({path:paths.get(w.node),kind:w.node.kind,isMathAccent:!!w.isMathAccent,baseCorePath:paths.get(w.baseCore.node),baseIsChar:!!w.baseIsChar,accentOverRaw:w.baseHasAccentOver,accentUnderRaw:w.baseHasAccentUnder,accentOver:!!w.baseHasAccentOver,accentUnder:!!w.baseHasAccentUnder,scriptCorePath:paths.get(w.scriptChild.coreMO().node),baseSize:w.baseCore.size??null,baseRelativeScale:w.baseCore.bbox.rscale});return w};
 const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){math.root=root;return original.call(this,math,doc)};
 let svg=null,renderError=null;try{svg=adaptor.innerHTML(html.convert('x',{display:request.display,em:16,ex:8}))}catch(e){renderError=String(e)}return{fullTree,observations,svg,renderError};})()`,c);
 const label=name+(display?'-display':'-inline');records.push({name:label,display,spec,...result});
 if(result.svg!==null)fs.writeFileSync(path.join(evidence,label+'.svg'),result.svg);
}
fs.writeFileSync(path.join(evidence,'constructor.json'),JSON.stringify({source:'Unmodified pinned CommonScriptbase constructor, registered MML inputs, observer only',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' registered constructor observations');
