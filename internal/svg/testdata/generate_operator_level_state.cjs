// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Public current inherited operator script-level references; no renderer overlays.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_operator_level.cjs /path/to/pinned/d2latex [evidence-directory]"
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

const cases = [
  { name: "inherited-zero", inherited: 0, previous: 2 },
  { name: "inherited-two", inherited: 2, previous: 0 },
  { name: "explicit-ignored", inherited: 1, explicit: 4, previous: 0 },
  { name: "default-after-absent", default: 3, explicit: 5, previous: 0 },
  { name: "global-after-absent", global: 2, explicit: 5, previous: 0 },
  { name: "zero-default", explicit: 5, previous: 2 },
  { name: "no-previous-fresh", inherited: 2, previous: null },
  { name: "no-previous-after-call", inherited: 2, previous: 0, thenNull: true },
  { name: "none-class", inherited: 2, previous: 1, none: true },
];
const records = [];
for (const kind of ["mo", "TeXAtom"])
  for (const spec of cases) {
    const c = vm.createContext({
      console: { log() {}, warn() {}, error() {} },
    });
    for (const [f, b] of assets)
      vm.runInContext(b.toString(), c, { filename: f });
    c.request = { kind, spec };
    const result = vm.runInContext(
      `(()=>{
  const factory=html.inputJax[0].mmlFactory,s=request.spec;
  const node=request.kind==='mo'?factory.create('mo',{},[factory.create('text').setText('=')]):factory.create('TeXAtom',{},[factory.create('mi',{},[factory.create('text').setText('x')])]);
  node.setInheritedAttributes({},true,0,false);
  delete node.attributes.getAllInherited().scriptlevel;
  if(s.inherited!==undefined)node.attributes.setInherited('scriptlevel',s.inherited);
  if(s.explicit!==undefined)node.attributes.set('scriptlevel',s.explicit);
  if(s.default!==undefined)node.attributes.getAllDefaults().scriptlevel=s.default;
  if(s.global!==undefined){delete node.attributes.getAllDefaults().scriptlevel;node.attributes.getAllGlobals().scriptlevel=s.global}
  const previous=s.previous===null?null:factory.create('mi',{scriptlevel:s.previous},[factory.create('text').setText('y')]);
  if(previous){previous.setInheritedAttributes({},true,0,false);previous.setTeXclass(null)}
  if(s.none){node.texClass=-1;node.prevLevel=2;node.setProperty('texClass',-1)}
  const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.coreIndex||0},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});
  const state=n=>({kind:n.kind,texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(state)});
  const before=full(node),previousBefore=previous?full(previous):null;
  const observation=()=>({currentExplicit:node.attributes.getExplicit('scriptlevel')??null,currentInherited:node.attributes.getInherited('scriptlevel')??null,currentResolved:node.attributes.get('scriptlevel')??null,previousResolved:previous?previous.attributes.get('scriptlevel'):null,rawPreviousLevel:node.prevLevel??null});
  const initialObservation=observation();const returned=node.adjustTeXclass(previous);const after=state(node),previousAfter=previous?state(previous):null,afterObservation=observation();
  let afterNull=null;if(s.thenNull){node.adjustTeXclass(null);afterNull=state(node)}
  return{before,previous:previousBefore,after,previousAfter,returned:returned===node?'node':returned===previous?'previous':'unexpected',afterNull,initialObservation,afterObservation};
 })()`,
      c
    );
    records.push({ name: kind + "-" + spec.name, kind, spec, ...result });
  }
fs.writeFileSync(
  path.join(__dirname, "operator_level_state_mathjax_3_2_2.json"),
  JSON.stringify(
    {
      oracle:
        "Unmodified pinned MathJax3.2.2 registered MmlMo/TeXAtom adjustTeXclass before/after state; current inherited/default/global boundary, explicit ignored and previous-null retention",
      assetsSHA256: hashes,
      cases: records,
    },
    null,
    2
  ) + "\n"
);
console.log(records.length + " direct inherited-level policy controls");
