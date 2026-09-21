// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Fixed-size delimiter alias selection beyond the largest available glyph.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_wide_accents.cjs /path/to/pinned/d2latex [evidence-directory]"
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
    "widehat-one",
    "\\widehat{x}"
  ],
  [
    "widehat-two",
    "\\widehat{xy}"
  ],
  [
    "widehat-three",
    "\\widehat{xyz}"
  ],
  [
    "widehat-four",
    "\\widehat{abcd}"
  ],
  [
    "widehat-long",
    "\\widehat{abcdef}"
  ],
  [
    "widehat-very-long",
    "\\widehat{abcdefghijklmnop}"
  ],
  [
    "widehat-roman",
    "\\widehat{\\mathrm{abcdef}}"
  ],
  [
    "widehat-bold",
    "\\widehat{\\mathbf{abcdef}}"
  ],
  [
    "widehat-fraction",
    "\\widehat{\\frac{abcdef}{xy}}"
  ],
  [
    "widehat-scripts",
    "\\widehat{x_1+y_2}"
  ],
  [
    "widetilde-one",
    "\\widetilde{x}"
  ],
  [
    "widetilde-two",
    "\\widetilde{xy}"
  ],
  [
    "widetilde-three",
    "\\widetilde{xyz}"
  ],
  [
    "widetilde-four",
    "\\widetilde{abcd}"
  ],
  [
    "widetilde-long",
    "\\widetilde{abcdef}"
  ],
  [
    "widetilde-very-long",
    "\\widetilde{abcdefghijklmnop}"
  ],
  [
    "widetilde-roman",
    "\\widetilde{\\mathrm{abcdef}}"
  ],
  [
    "widetilde-bold",
    "\\widetilde{\\mathbf{abcdef}}"
  ],
  [
    "widetilde-fraction",
    "\\widetilde{\\frac{abcdef}{xy}}"
  ],
  [
    "widetilde-scripts",
    "\\widetilde{x_1+y_2}"
  ],
  [
    "ordinary-hat",
    "\\hat{abcdef}"
  ],
  [
    "ordinary-tilde",
    "\\tilde{abcdef}"
  ],
  [
    "overarrow",
    "\\overrightarrow{abcdef}"
  ],
  [
    "ordinary",
    "x+y+1"
  ],
  [
    "nested-script",
    "z^{\\widehat{abcdef}}"
  ],
  [
    "nested-wide",
    "\\widehat{\\widetilde{abcdef}}"
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
 records.push({name,tex,display,width:Math.ceil(Number(result.svg.match(/width="([\d.]+)ex"/)[1])*8),height:Math.ceil(Number(result.svg.match(/height="([\d.]+)ex"/)[1])*8),svgSHA256:hash(result.svg)});
}
fs.writeFileSync(path.join(__dirname,'wide_accent_mathjax_3_2_2.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' complete SVG cases');
