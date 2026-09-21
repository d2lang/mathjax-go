// SPDX-License-Identifier: Apache-2.0
// Pure registered MathML accent inheritance and getter/remap policy; no TeX dispatch overlays.
const fs = require("node:fs"),
  vm = require("node:vm"),
  path = require("node:path"),
  crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_accent_policy.cjs /path/to/pinned/d2latex [evidence-directory]"
  );
if (evidence) fs.mkdirSync(evidence, { recursive: true });
const hash = (b) => crypto.createHash("sha256").update(b).digest("hex");
const hashes = {
  "polyfills.js":
    "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01",
  "mathjax.js":
    "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869",
  "setup.js":
    "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881",
};
const assets = Object.entries(hashes).map(([f, h]) => {
  const b = fs.readFileSync(path.join(base, f));
  if (hash(b) !== h) throw Error("Unpinned " + f);
  return [f, b];
});
const t = (text) => ({ kind: "text", text });
const n = (kind, attributes = {}, children = [], inherited = {}) => ({
  kind,
  attributes,
  children,
  inherited,
});
const mo=(attrs={},marker)=>{const x=n("mo",attrs,[t("→")]);if(marker!==undefined)x.properties={mathaccent:marker};return x};
const mi=()=>n("mi",{},[t("x")]);
const cases=[];
for(const [parentName,parentValue]of [["unset",undefined],["null",null],["false",false],["true",true]])
 for(const [operatorName,operatorValue]of [["unset",undefined],["false",false],["true",true]]){
  const attrs=parentValue===undefined?{}:{accent:parentValue};const op=operatorValue===undefined?{}:{accent:operatorValue};op.stretchy=false;
  cases.push(["over-"+parentName+"-operator-"+operatorName,n("mover",attrs,[mi(),mo(op,true)])]);
 }
for(const [name,value]of [["unset",undefined],["null",null],["false",false],["true",true]]){
 cases.push(["under-"+name,n("munder",value===undefined?{}:{accentunder:value},[mi(),mo({accent:true,stretchy:false},true)])]);
}
cases.push(["both-mixed",n("munderover",{},[mi(),mo({accent:false,stretchy:false},true),mo({accent:true,stretchy:false},false)])]);
cases.push(["grouped-atom",n("mover",{},[mi(),n("TeXAtom",{},[mo({accent:false,stretchy:false},true)])])]);
cases.push(["grouped-row",n("mover",{},[mi(),n("mrow",{},[mo({accent:true,stretchy:false},true)])])]);
cases.push(["grouped-style",n("munder",{},[mi(),n("mstyle",{mathcolor:"red"},[mo({accent:true,stretchy:false},true)])])]);
cases.push(["ordinary-marker",mo({accent:true,stretchy:false},true)]);
cases.push(["ordinary-unmarked",mo({accent:true,stretchy:false},false)]);
cases.push(["ordinary-null",mo({accent:true,stretchy:false},null)]);
cases.push(["ordinary-absent",mo({accent:true,stretchy:false})]);
cases.push(["base-not-script",n("mover",{},[mo({accent:true,stretchy:false},true),mi()])]);
cases.push(["genuine-no-marker",n("mover",{},[mi(),mo({accent:true,stretchy:false})])]);
const records = [];
for (const [name, spec] of cases)
  for (const display of [false, true]) {
    const c = vm.createContext({
      console: { log() {}, warn() {}, error() {} },
    });
    for (const [f, b] of assets)
      vm.runInContext(b.toString(), c, { filename: f });
    c.request = { spec, display };
    const result = vm.runInContext(
      `(()=>{
  const factory=html.inputJax[0].mmlFactory;
  const overlays=[];
  function build(s){if(s.kind==='text')return factory.create('text').setText(s.text);const node=factory.create(s.kind,s.attributes,s.children.map(build));for(const[k,v]of Object.entries(s.properties||{}))node.setProperty(k,v);if(Object.keys(s.inherited).length)overlays.push([node,s.inherited]);return node}
  const root=factory.create('math',request.display?{display:'block'}:{},[build(request.spec)]);
  root.setInheritedAttributes({},request.display,0,false);
  for(const [node,attrs]of overlays)for(const [k,v]of Object.entries(attrs))node.attributes.setInherited(k,v);
  root.setTeXclass(null);
  const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));
  const project=node=>({kind:node.kind,text:node.kind==='text'?node.getText():null,
   attributes:node.attributes?{explicit:pairs(node.attributes.getAllAttributes()),inherited:pairs(node.attributes.getAllInherited()),defaults:pairs(node.attributes.getAllDefaults()),global:pairs(node.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},
   properties:pairs(node.getAllProperties()),
   flags:{token:node.isToken,embellished:node.isEmbellished,spacelike:node.isSpacelike,linebreakContainer:node.linebreakContainer,hasNewline:node.hasNewline,inferred:node.isInferred,notParent:node.notParent,arity:node.arity===Infinity?-2:node.arity,coreIndex:node.coreIndex||0},
   texClass:Number.isFinite(node.texClass)?node.texClass:-1,prevClass:Number.isFinite(node.prevClass)?node.prevClass:-1,prevLevel:Number.isFinite(node.prevLevel)?node.prevLevel:0,
   children:node.childNodes.map(project)});
  const tree=project(root), before=JSON.stringify(tree),original=html.outputJax.typeset;
  const observations=[];function observe(node,path=[]){if(node.kind!=='text'){const attrs=node.attributes,keys=['accent','accentunder','scriptlevel','displaystyle'];const values={};for(const key of keys)values[key]={explicit:attrs.getExplicit(key)===undefined?'__absent__':attrs.getExplicit(key),inherited:attrs.getInherited(key)===undefined?'__absent__':attrs.getInherited(key),resolved:attrs.get(key)===undefined?'__absent__':attrs.get(key)};observations.push({path,kind:node.kind,values,prime:node.getProperty('texprimestyle')===undefined?'__absent__':node.getProperty('texprimestyle'),isAccent:node.kind==='mo'?node.isAccent:null})}node.childNodes.forEach((c,i)=>observe(c,path.concat(i)))}observe(root);
  html.outputJax.typeset=function(math,doc){math.root=root;return original.call(this,math,doc)};
  const svg=adaptor.innerHTML(html.convert('x',{display:request.display,em:16,ex:8}));
  if(JSON.stringify(project(root))!==before)throw Error('Output mutated original MathML');
  return{tree,svg,observations};
 })()`,
      c
    );
    const id = name + "-" + (display ? "display" : "inline");
    if (evidence) {fs.writeFileSync(path.join(evidence, id + ".svg"), result.svg);fs.writeFileSync(path.join(evidence,id+".json"),JSON.stringify({tree:result.tree,observations:result.observations},null,2)+"\n");}
    records.push({
      name: id,
      display,
      spec,
      tree: result.tree,
      observations: result.observations,
      sha256: hash(result.svg),
    });
  }
fs.writeFileSync(
  path.join(__dirname, "accent_policy_mathjax_3_2_2.json"),
  JSON.stringify(
    {
      oracle:
        "Unmodified pinned D2 MathJax 3.2.2, direct registered MathML trees; no parser command dependency",
      assetsSHA256: hashes,
      cases: records,
    },
    null,
    2
  ) + "\n"
);
console.log(records.length + " pinned registered-MathML SVG cases");
