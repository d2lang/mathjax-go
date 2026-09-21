// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// D041 fixed-symbol/default-token font policy is a separate recorded discrepancy.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_nested_fonts.cjs /path/to/pinned/d2latex [evidence-directory]"
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
    "bold-roman",
    "\\mathbf{\\mathrm{x}}"
  ],
  [
    "roman-bold",
    "\\mathrm{\\mathbf{x}}"
  ],
  [
    "bold-roman-scope",
    "\\mathbf{a\\mathrm{x}b}+y"
  ],
  [
    "roman-bold-scope",
    "\\mathrm{a\\mathbf{x}b}+y"
  ],
  [
    "triple",
    "\\mathbf{a\\mathrm{b\\mathbf{x}c}d}"
  ],
  [
    "multiple",
    "\\mathbf{a\\mathrm{xy}b}"
  ],
  [
    "fraction",
    "\\mathbf{\\frac{a}{\\mathrm{x}}}"
  ],
  [
    "script",
    "\\mathbf{a_{\\mathrm{x}}^2}"
  ],
  [
    "radical",
    "\\mathrm{\\sqrt{\\mathbf{x}}}"
  ],
  [
    "decl-inner-roman",
    "\\mathbf{a{\\rm x}b}"
  ],
  [
    "decl-inner-bold",
    "\\mathrm{a{\\bf x}b}"
  ],
  [
    "decl-outer-bold",
    "{\\bf a\\mathrm{x}b}+y"
  ],
  [
    "decl-outer-roman",
    "{\\rm a\\mathbf{x}b}+y"
  ],
  [
    "decl-sequence",
    "{\\bf a\\rm x\\bf b}+y"
  ],
  [
    "decl-nested-group",
    "{\\bf a{\\rm x}b}+y"
  ],
  [
    "nested-operator",
    "\\mathbf{\\mathrm{+x}}"
  ],
  [
    "kept-token",
    "\\mathbf{\\mathrm{\\mmlToken{mi}[mathvariant=double-struck]{R}}}"
  ],
  [
    "kept-declaration",
    "{\\bf \\rm \\mmlToken{mi}[mathvariant=double-struck]{R}}"
  ],
  [
    "plain-control",
    "x+y"
  ],
  [
    "bold-control",
    "\\mathbf{x}"
  ],
  [
    "roman-control",
    "\\mathrm{x}"
  ],
  [
    "function-control",
    "\\sin x"
  ],
  [
    "bold-italic",
    "\\mathbf{\\mathit{x}}"
  ],
  [
    "italic-bold",
    "\\mathit{\\mathbf{x}}"
  ],
  [
    "sans-mono",
    "\\mathsf{\\mathtt{x}}"
  ],
  [
    "mono-sans",
    "\\mathtt{\\mathsf{x}}"
  ],
  [
    "word-italic",
    "\\mathbf{xy\\mathit{z}}"
  ],
  [
    "declaration-italic",
    "{\\bf a{\\it x}b}"
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
fs.writeFileSync(path.join(__dirname,'nested_fonts_mathjax_3_2_2.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' exact AST and complete SVG cases');
