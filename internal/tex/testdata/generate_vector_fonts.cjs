// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Physics VectorBold font policy, with inherited accent/tree limits documented separately.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_vector_fonts.cjs /path/to/pinned/d2latex [evidence-directory]"
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
  [
    "plain",
    "\\vb{x}"
  ],
  [
    "star",
    "\\vb*{x}"
  ],
  [
    "outer-roman",
    "\\mathrm{\\vb{x}}"
  ],
  [
    "outer-bold",
    "\\mathbf{\\vb*{x}}"
  ],
  [
    "inner-roman",
    "\\vb{\\mathrm{x+\\alpha+\\Gamma}}"
  ],
  [
    "inner-bold",
    "\\vb*{\\mathbf{x+\\alpha+\\Gamma}}"
  ],
  [
    "inner-mathit",
    "\\vb{\\mathit{x}}"
  ],
  [
    "declaration-outer",
    "{\\rm \\vb{x}+y}"
  ],
  [
    "declaration-inner",
    "\\vb{x+{\\rm y+\\alpha}+z}"
  ],
  [
    "declaration-rest",
    "\\vb{x+\\rm y+\\alpha}"
  ],
  [
    "greek",
    "\\vb{\\alpha+α+\\phi+ϕ}"
  ],
  [
    "greek-star",
    "\\vb*{\\alpha+α+\\phi+ϕ}"
  ],
  [
    "caps",
    "\\vb{\\Gamma+Γ+A+Z}"
  ],
  [
    "digits",
    "\\vb{1+12+2.3}"
  ],
  [
    "mixed",
    "\\vb{x+y+1+\\sum+\\sin x}"
  ],
  [
    "accent-hat",
    "\\vb{\\hat{x}}"
  ],
  [
    "accent-dot",
    "\\vb*{\\dot{x}}"
  ],
  [
    "multi-outer",
    "\\mathrm{\\vb{abc}}"
  ],
  [
    "multi-inner",
    "\\vb{\\mathrm{abc}}"
  ],
  [
    "multi-fraction",
    "\\mathrm{\\vb{\\frac{ab}{cd}}}"
  ],
  [
    "authored",
    "\\vb{\\mmlToken{mi}{x}}"
  ],
  [
    "authored-kept",
    "\\vb*{\\mmlToken{mi}[mathvariant=normal]{x}}"
  ],
  [
    "authored-outer",
    "\\mathbf{\\vb{\\mmlToken{mi}{x}}}"
  ],
  [
    "nested",
    "\\vb{x\\vb*{y}z}"
  ],
  [
    "nested-group",
    "\\vb{x{\\vb*{y}}z}"
  ],
  [
    "nested-inner-font",
    "\\vb{x\\mathrm{\\vb*{y}+z}q}"
  ],
  [
    "arrow",
    "\\va{x}"
  ],
  [
    "arrow-star",
    "\\va*{\\alpha}"
  ],
  [
    "unit",
    "\\vu{x}"
  ],
  [
    "unit-star",
    "\\vu*{\\alpha}"
  ],
  [
    "outer-arrow",
    "\\mathrm{\\va{x}}"
  ],
  [
    "nested-arrow",
    "\\vb{x+\\va*{y}+z}"
  ],
  [
    "script",
    "\\vb{x_i^2}"
  ],
  [
    "script-font",
    "\\vb{x_{\\rm y}}"
  ],
  [
    "text",
    "\\vb{\\text{x}}"
  ],
  [
    "supplementary",
    "\\vb*{𝑥}"
  ],
  [
    "ordinary-control",
    "x+\\alpha+\\Gamma+1"
  ],
  [
    "after-vector",
    "\\mathrm{a+\\vb{x}+b}"
  ],
  [
    "after-inner",
    "\\vb{\\mathrm{x}+y}"
  ],
  [
    "long-alias",
    "\\vectorbold*{x}+\\vectorarrow{y}+\\vectorunit{z}"
  ],
  [
    "multi-bold-outer",
    "\\mathbf{\\vb{abc}}x"
  ],
  [
    "bold-arrow-control",
    "\\vb{\\vec{y}}"
  ],
  [
    "multi-alias",
    "\\mathrm{\\va{abc}}"
  ],
  [
    "text-inner-font",
    "\\vb{\\rm\\alpha+\\Gamma+\\text{x}}"
  ],
  [
    "quick-text",
    "\\vb{\\qq{x}}"
  ],
  [
    "clap-text",
    "\\vb{\\clap{x}}"
  ],
  [
    "operator-font",
    "\\vb{\\operatorname{x}+\\operatorname{\\alpha}}"
  ],
  [
    "outer-unselected",
    "\\mathit{\\vb{12+\\alpha+\\Gamma+\\sum}}"
  ]
];
const records=[];
for(const [label,tex] of inputs)for(const display of [false,true]){
 const name=label+(display?'-display':'-inline');
 const c=vm.createContext({console:{log(){},warn(){},error(){}}});
 for(const [f,b]of assets)vm.runInContext(b.toString(),c,{filename:f});c.request={tex,display};
 const result=vm.runInContext(`(()=>{let tree;const project=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?Object.fromEntries(n.attributes.getExplicitNames().sort().map(k=>[k,n.attributes.getExplicit(k)])):{},children:n.childNodes.map(project)});const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){tree=project(math.root);return original.call(this,math,doc)};const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));return{tree,svg};})()`,c);
 if(result.svg.includes('data-mml-node="merror"'))throw Error(name+' error');
 if(evidence)fs.writeFileSync(path.join(evidence,name+'.svg'),result.svg);
 records.push({name,tex,display,tree:result.tree,svgSHA256:hash(result.svg)});
}
fs.writeFileSync(path.join(__dirname,'vector_font_mathjax_3_2_2.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' exact AST and complete SVG cases');
