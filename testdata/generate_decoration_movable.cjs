// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// D072 ordinary-decoration direct-base normalization references.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate-decoration-permission.cjs /path/to/pinned/d2latex [evidence-directory]"
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
  {
    "name": "underline-sub-inline",
    "tex": "\\underline{x}_i",
    "display": false
  },
  {
    "name": "underline-sub-display",
    "tex": "\\underline{x}_i",
    "display": true
  },
  {
    "name": "underright-sub-inline",
    "tex": "\\underrightarrow{x}_i",
    "display": false
  },
  {
    "name": "underright-sub-display",
    "tex": "\\underrightarrow{x}_i",
    "display": true
  },
  {
    "name": "underleft-sub-inline",
    "tex": "\\underleftarrow{x}_i",
    "display": false
  },
  {
    "name": "underleft-sub-display",
    "tex": "\\underleftarrow{x}_i",
    "display": true
  },
  {
    "name": "underline-both-inline",
    "tex": "\\underline{x}_i^n",
    "display": false
  },
  {
    "name": "underline-both-display",
    "tex": "\\underline{x}_i^n",
    "display": true
  },
  {
    "name": "underarrow-both-inline",
    "tex": "\\underrightarrow{x}_i^n",
    "display": false
  },
  {
    "name": "underarrow-both-display",
    "tex": "\\underrightarrow{x}_i^n",
    "display": true
  },
  {
    "name": "underline-grouped-inline",
    "tex": "{\\underline{x}}_i",
    "display": false
  },
  {
    "name": "underline-grouped-display",
    "tex": "{\\underline{x}}_i",
    "display": true
  },
  {
    "name": "underline-nested-inline",
    "tex": "\\underline{\\hat{x}}_i",
    "display": false
  },
  {
    "name": "underline-nested-display",
    "tex": "\\underline{\\hat{x}}_i",
    "display": true
  },
  {
    "name": "underline-multichar-inline",
    "tex": "\\underline{xy}_i",
    "display": false
  },
  {
    "name": "underline-multichar-display",
    "tex": "\\underline{xy}_i",
    "display": true
  },
  {
    "name": "underline-bare-inline",
    "tex": "\\underline{x}",
    "display": false
  },
  {
    "name": "underline-bare-display",
    "tex": "\\underline{x}",
    "display": true
  },
  {
    "name": "underarrow-bare-inline",
    "tex": "\\underrightarrow{x}",
    "display": false
  },
  {
    "name": "underarrow-bare-display",
    "tex": "\\underrightarrow{x}",
    "display": true
  },
  {
    "name": "overline-sup-inline",
    "tex": "\\overline{x}^n",
    "display": false
  },
  {
    "name": "overline-sup-display",
    "tex": "\\overline{x}^n",
    "display": true
  },
  {
    "name": "overright-sup-inline",
    "tex": "\\overrightarrow{x}^n",
    "display": false
  },
  {
    "name": "overright-sup-display",
    "tex": "\\overrightarrow{x}^n",
    "display": true
  },
  {
    "name": "overleft-sup-inline",
    "tex": "\\overleftarrow{x}^n",
    "display": false
  },
  {
    "name": "overleft-sup-display",
    "tex": "\\overleftarrow{x}^n",
    "display": true
  },
  {
    "name": "bar-sub-inline",
    "tex": "\\bar{x}_i",
    "display": false
  },
  {
    "name": "bar-sub-display",
    "tex": "\\bar{x}_i",
    "display": true
  },
  {
    "name": "underbrace-label-inline",
    "tex": "\\underbrace{x}_i",
    "display": false
  },
  {
    "name": "underbrace-label-display",
    "tex": "\\underbrace{x}_i",
    "display": true
  },
  {
    "name": "overbrace-label-inline",
    "tex": "\\overbrace{x}^n",
    "display": false
  },
  {
    "name": "overbrace-label-display",
    "tex": "\\overbrace{x}^n",
    "display": true
  },
  {
    "name": "underset-duplicate-inline",
    "tex": "\\underset{a}{x}_i",
    "display": false
  },
  {
    "name": "underset-duplicate-display",
    "tex": "\\underset{a}{x}_i",
    "display": true
  },
  {
    "name": "underline-duplicate-inline",
    "tex": "\\underline{x}_i_j",
    "display": false
  },
  {
    "name": "underline-duplicate-display",
    "tex": "\\underline{x}_i_j",
    "display": true
  },
  {
    "name": "underline-nested-script-inline",
    "tex": "\\underline{x}_{i_j}",
    "display": false
  },
  {
    "name": "underline-nested-script-display",
    "tex": "\\underline{x}_{i_j}",
    "display": true
  },
  {
    "name": "ordinary-duplicate-inline",
    "tex": "x_i_j",
    "display": false
  },
  {
    "name": "ordinary-duplicate-display",
    "tex": "x_i_j",
    "display": true
  },
  {
    "name": "underline-sum-sub-inline",
    "tex": "\\underline{\\sum}_i",
    "display": false
  },
  {
    "name": "underline-sum-sub-display",
    "tex": "\\underline{\\sum}_i",
    "display": true
  },
  {
    "name": "underright-sum-sub-inline",
    "tex": "\\underrightarrow{\\sum}_i",
    "display": false
  },
  {
    "name": "underright-sum-sub-display",
    "tex": "\\underrightarrow{\\sum}_i",
    "display": true
  },
  {
    "name": "overline-sum-sup-inline",
    "tex": "\\overline{\\sum}^n",
    "display": false
  },
  {
    "name": "overline-sum-sup-display",
    "tex": "\\overline{\\sum}^n",
    "display": true
  },
  {
    "name": "overright-sum-sup-inline",
    "tex": "\\overrightarrow{\\sum}^n",
    "display": false
  },
  {
    "name": "overright-sum-sup-display",
    "tex": "\\overrightarrow{\\sum}^n",
    "display": true
  },
  {
    "name": "overline-product-inline",
    "tex": "\\overline{\\prod}^n",
    "display": false
  },
  {
    "name": "overline-product-display",
    "tex": "\\overline{\\prod}^n",
    "display": true
  },
  {
    "name": "overline-integral-inline",
    "tex": "\\overline{\\int}^n",
    "display": false
  },
  {
    "name": "overline-integral-display",
    "tex": "\\overline{\\int}^n",
    "display": true
  },
  {
    "name": "overline-mathop-inline",
    "tex": "\\overline{\\mathop{x}}^n",
    "display": false
  },
  {
    "name": "overline-mathop-display",
    "tex": "\\overline{\\mathop{x}}^n",
    "display": true
  },
  {
    "name": "overline-group-sum-inline",
    "tex": "\\overline{{\\sum}}^n",
    "display": false
  },
  {
    "name": "overline-group-sum-display",
    "tex": "\\overline{{\\sum}}^n",
    "display": true
  },
  {
    "name": "overline-scripted-sum-inline",
    "tex": "\\overline{\\sum_i^n}^a",
    "display": false
  },
  {
    "name": "overline-scripted-sum-display",
    "tex": "\\overline{\\sum_i^n}^a",
    "display": true
  },
  {
    "name": "underarrow-scripted-sum-inline",
    "tex": "\\underrightarrow{\\sum_i^n}_a",
    "display": false
  },
  {
    "name": "underarrow-scripted-sum-display",
    "tex": "\\underrightarrow{\\sum_i^n}_a",
    "display": true
  }
];
const explicit=n=>({kind:n.kind,text:n.text,attributes:n.attributes,children:n.children.map(explicit)});
const records=[];
for(const {name,tex,display} of inputs){
 const c=vm.createContext({console:{log(){},warn(){},error(){}}});
 for(const [f,b]of assets)vm.runInContext(b.toString(),c,{filename:f});c.request={tex,display};
 const result=vm.runInContext(`(()=>{let tree, fullTree;const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.coreIndex||0},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});const project=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?Object.fromEntries(n.attributes.getExplicitNames().sort().map(k=>[k,n.attributes.getExplicit(k)])):{},properties:n.getAllProperties?n.getAllProperties():{},isAccent:n.kind==='mo'?n.isAccent:null,coreParentKind:n.kind==='mo'?n.coreParent().kind:null,coreParentParentKind:n.kind==='mo'&&n.coreParent().parent?n.coreParent().parent.kind:null,children:n.childNodes.map(project)});const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){tree=project(math.root);fullTree=full(math.root);return original.call(this,math,doc)};const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));return{tree,svg,fullTree};})()`,c);
 // Preserve primary complete error SVG and observed tree alongside valid cases.
 if(evidence){fs.writeFileSync(path.join(evidence,name+'.svg'),result.svg);fs.writeFileSync(path.join(evidence,name+'.json'),JSON.stringify(result.tree,null,2)+'\n');}
 records.push({name,tex,display,width:Math.ceil(Number(result.svg.match(/width="([\d.]+)ex"/)[1])*8),height:Math.ceil(Number(result.svg.match(/height="([\d.]+)ex"/)[1])*8),tree:explicit(result.tree),propertiesTree:result.tree,svgSHA256:hash(result.svg)});
}
fs.writeFileSync(path.join(evidence||__dirname,'decoration_movable_mathjax_3_2_2.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' exact AST and complete SVG cases');
