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
const inputs = [
  ["style-change", "\\scriptstyle x\\textstyle +y", false],
  ["style-change", "\\scriptstyle x\\textstyle +y", true],
  ["script-script-change", "\\scriptscriptstyle x\\textstyle +y", false],
  ["script-script-change", "\\scriptscriptstyle x\\textstyle +y", true],
  ["grouped-prev", "{\\scriptstyle x}+y", false],
  ["grouped-prev", "{\\scriptstyle x}+y", true],
  ["grouped-bin", "x+{\\scriptstyle y}+z", false],
  ["grouped-bin", "x+{\\scriptstyle y}+z", true],
  ["style-op", "\\scriptstyle x\\textstyle \\sum_i^n+y", false],
  ["style-op", "\\scriptstyle x\\textstyle \\sum_i^n+y", true],
  ["style-rel", "\\scriptstyle x\\textstyle =y", false],
  ["style-rel", "\\scriptstyle x\\textstyle =y", true],
  ["current-small", "x\\scriptstyle +y", false],
  ["current-small", "x\\scriptstyle +y", true],
  ["plain", "x+y", false],
  ["plain", "x+y", true],
];
const records = [];
for (const [label, tex, display] of inputs) {
  const name = label + (display ? "-display" : "-inline");
  const c = vm.createContext({ console: { log() {}, warn() {}, error() {} } });
  for (const [f, b] of assets)
    vm.runInContext(b.toString(), c, { filename: f });
  c.request = { tex, display };
  const result = vm.runInContext(
    `(()=>{let tree, fullTree, afterTree;const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.coreIndex||0},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});const project=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?Object.fromEntries(n.attributes.getExplicitNames().sort().map(k=>[k,n.attributes.getExplicit(k)])):{},properties:n.getAllProperties?n.getAllProperties():{},isAccent:n.kind==='mo'?n.isAccent:null,coreParentKind:n.kind==='mo'?n.coreParent().kind:null,coreParentParentKind:n.kind==='mo'&&n.coreParent().parent?n.coreParent().parent.kind:null,children:n.childNodes.map(project)});const state=n=>({kind:n.kind,texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(state)});const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){tree=project(math.root);fullTree=full(math.root);const out=original.call(this,math,doc);afterTree=state(math.root);return out};const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));return{tree,svg,fullTree,afterTree};})()`,
    c
  );
  if (result.svg.includes('data-mml-node="merror"'))
    throw Error(name + " error");
  if (evidence) {
    fs.writeFileSync(path.join(evidence, name + ".svg"), result.svg);
    fs.writeFileSync(
      path.join(evidence, name + ".json"),
      JSON.stringify(result.tree, null, 2) + "\n"
    );
  }
  records.push({
    name,
    tex,
    display,
    width: Math.ceil(Number(result.svg.match(/width="([\d.]+)ex"/)[1]) * 8),
    height: Math.ceil(Number(result.svg.match(/height="([\d.]+)ex"/)[1]) * 8),
    fullTree: result.fullTree,
    afterTree: result.afterTree,
    svgSHA256: hash(result.svg),
  });
}
fs.writeFileSync(
  path.join(__dirname, "operator_level_mathjax_3_2_2.json"),
  JSON.stringify(
    {
      oracle:
        "Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer",
      mathjaxGitCommit: "ad8f5c21cb810236551da8c6512ba733e67357ee",
      assetsSHA256: hashes,
      cases: records,
    },
    null,
    2
  ) + "\n"
);
console.log(records.length + " exact AST and complete SVG cases");
