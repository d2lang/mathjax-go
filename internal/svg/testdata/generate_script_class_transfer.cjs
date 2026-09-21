// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Immediate script-base TeX-class transfer; no parser or renderer overlays.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_script_class.cjs /path/to/pinned/d2latex [evidence-directory]"
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

const t = (text) => ({ kind: "text", text });
const n = (kind, children = [], attributes = {}) => ({
  kind,
  children,
  attributes,
});
const mi = () => n("mi", [t("a")]),
  op = () => n("mo", [t("∫")]);
const inner = () => n("msub", [op(), mi()]);
const inputs = [
  ["nested", n("msup", [inner(), mi()])],
  ["row-base", n("msup", [n("mrow", [inner()]), mi()])],
  [
    "style-base",
    n("msup", [n("mstyle", [inner()], { mathcolor: "red" }), mi()]),
  ],
  ["direct-mo", n("msup", [op(), mi()])],
  ["direct-mi", n("msup", [mi(), mi()])],
  ["nonembellished", n("msup", [n("mrow", [mi(), mi()]), mi()])],
];
const records = [];
for (const [label, spec] of inputs)
  for (const level of [0, 2]) {
    const c = vm.createContext({
      console: { log() {}, warn() {}, error() {} },
    });
    for (const [f, b] of assets)
      vm.runInContext(b.toString(), c, { filename: f });
    c.request = { spec, level };
    const result = vm.runInContext(
      `(()=>{
 const factory=html.inputJax[0].mmlFactory;
 const build=s=>s.kind==='text'?factory.create('text').setText(s.text):factory.create(s.kind,s.attributes,s.children.map(build));
 const root=build(request.spec),previous=factory.create('mo',{scriptlevel:request.level},[factory.create('text').setText('=')]);
 root.setInheritedAttributes({},true,request.level,false);previous.setInheritedAttributes({},true,request.level,false);previous.setTeXclass(null);
 const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));
 const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.coreIndex||0},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});
 const state=n=>({kind:n.kind,texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(state)});
 const before=full(root),prevBefore=full(previous),returned=root.setTeXclass(previous);
 let returnedPath=null;function seek(n,p=[]){if(n===returned)returnedPath=p;n.childNodes.forEach((x,i)=>seek(x,p.concat(i)))}seek(root);
 if(returnedPath===null)throw Error('returned node outside root');
 return {before,previous:prevBefore,after:state(root),previousAfter:state(previous),returnedPath};
 })()`,
      c
    );
    records.push({ name: label + "-level" + level, ...result });
  }
fs.writeFileSync(
  path.join(__dirname, "script_class_transfer_mathjax_3_2_2.json"),
  JSON.stringify(
    {
      oracle:
        "Unmodified pinned MathJax 3.2.2 registered-MML setTeXclass, observed before/after and return identity",
      assetsSHA256: hashes,
      cases: records,
    },
    null,
    2
  ) + "\n"
);
console.log(records.length + " direct class/level transfer controls");
