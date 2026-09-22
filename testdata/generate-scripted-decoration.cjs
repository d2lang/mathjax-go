// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Scripted ordinary-decoration references. This generator never changes the oracle.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate-scripted-decoration.cjs /path/to/pinned/d2latex [evidence-directory]"
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
const inputs = {
  "cases": [
    {
      "name": "witness-inline",
      "tex": "\\overline{\\sum_i^n}^a",
      "display": false
    },
    {
      "name": "witness-display",
      "tex": "\\overline{\\sum_i^n}^a",
      "display": true
    },
    {
      "name": "sum-sub-inline",
      "tex": "\\overline{\\sum_i}",
      "display": false
    },
    {
      "name": "sum-sub-display",
      "tex": "\\overline{\\sum_i}",
      "display": true
    },
    {
      "name": "sum-sup-inline",
      "tex": "\\overline{\\sum^n}",
      "display": false
    },
    {
      "name": "sum-sup-display",
      "tex": "\\overline{\\sum^n}",
      "display": true
    },
    {
      "name": "product-both-inline",
      "tex": "\\overline{\\prod_i^n}^a",
      "display": false
    },
    {
      "name": "product-both-display",
      "tex": "\\overline{\\prod_i^n}^a",
      "display": true
    },
    {
      "name": "integral-both-inline",
      "tex": "\\overline{\\int_i^n}^a",
      "display": false
    },
    {
      "name": "integral-both-display",
      "tex": "\\overline{\\int_i^n}^a",
      "display": true
    },
    {
      "name": "mathop-both-inline",
      "tex": "\\overline{\\mathop{x}_i^n}^a",
      "display": false
    },
    {
      "name": "mathop-both-display",
      "tex": "\\overline{\\mathop{x}_i^n}^a",
      "display": true
    },
    {
      "name": "right-both-inline",
      "tex": "\\overrightarrow{\\sum_i^n}^a",
      "display": false
    },
    {
      "name": "right-both-display",
      "tex": "\\overrightarrow{\\sum_i^n}^a",
      "display": true
    },
    {
      "name": "left-both-inline",
      "tex": "\\overleftarrow{\\sum_i^n}^a",
      "display": false
    },
    {
      "name": "left-both-display",
      "tex": "\\overleftarrow{\\sum_i^n}^a",
      "display": true
    },
    {
      "name": "under-both-inline",
      "tex": "\\underline{\\sum_i^n}",
      "display": false
    },
    {
      "name": "under-both-display",
      "tex": "\\underline{\\sum_i^n}",
      "display": true
    },
    {
      "name": "underright-both-inline",
      "tex": "\\underrightarrow{\\sum_i^n}",
      "display": false
    },
    {
      "name": "underright-both-display",
      "tex": "\\underrightarrow{\\sum_i^n}",
      "display": true
    },
    {
      "name": "underleft-both-inline",
      "tex": "\\underleftarrow{\\sum_i^n}",
      "display": false
    },
    {
      "name": "underleft-both-display",
      "tex": "\\underleftarrow{\\sum_i^n}",
      "display": true
    },
    {
      "name": "nested-over-inline",
      "tex": "\\overline{\\overline{\\sum_i^n}}^a",
      "display": false
    },
    {
      "name": "nested-over-display",
      "tex": "\\overline{\\overline{\\sum_i^n}}^a",
      "display": true
    },
    {
      "name": "grouped-script-inline",
      "tex": "\\overline{{\\sum_i^n}}^a",
      "display": false
    },
    {
      "name": "grouped-script-display",
      "tex": "\\overline{{\\sum_i^n}}^a",
      "display": true
    },
    {
      "name": "styled-script-inline",
      "tex": "\\overline{\\displaystyle\\sum_i^n}^a",
      "display": false
    },
    {
      "name": "styled-script-display",
      "tex": "\\overline{\\displaystyle\\sum_i^n}^a",
      "display": true
    },
    {
      "name": "explicit-limits-inline",
      "tex": "\\overline{\\sum\\limits_i^n}^a",
      "display": false
    },
    {
      "name": "explicit-limits-display",
      "tex": "\\overline{\\sum\\limits_i^n}^a",
      "display": true
    },
    {
      "name": "explicit-nolimits-inline",
      "tex": "\\overline{\\sum\\nolimits_i^n}^a",
      "display": false
    },
    {
      "name": "explicit-nolimits-display",
      "tex": "\\overline{\\sum\\nolimits_i^n}^a",
      "display": true
    },
    {
      "name": "ordinary-inline",
      "tex": "\\overline{x}^a",
      "display": false
    },
    {
      "name": "ordinary-display",
      "tex": "\\overline{x}^a",
      "display": true
    },
    {
      "name": "ordinary-side-inline",
      "tex": "\\overline{x_i^n}^a",
      "display": false
    },
    {
      "name": "ordinary-side-display",
      "tex": "\\overline{x_i^n}^a",
      "display": true
    },
    {
      "name": "row-inline",
      "tex": "\\overline{x+\\sum_i^n}^a",
      "display": false
    },
    {
      "name": "row-display",
      "tex": "\\overline{x+\\sum_i^n}^a",
      "display": true
    },
    {
      "name": "brace-inline",
      "tex": "\\overbrace{\\sum_i^n}^a",
      "display": false
    },
    {
      "name": "brace-display",
      "tex": "\\overbrace{\\sum_i^n}^a",
      "display": true
    },
    {
      "name": "underbrace-inline",
      "tex": "\\underbrace{\\sum_i^n}_a",
      "display": false
    },
    {
      "name": "underbrace-display",
      "tex": "\\underbrace{\\sum_i^n}_a",
      "display": true
    },
    {
      "name": "bar-inline",
      "tex": "\\bar{\\sum_i^n}^a",
      "display": false
    },
    {
      "name": "bar-display",
      "tex": "\\bar{\\sum_i^n}^a",
      "display": true
    },
    {
      "name": "permission-error-inline",
      "tex": "\\underline{x}_a",
      "display": false
    },
    {
      "name": "permission-error-display",
      "tex": "\\underline{x}_a",
      "display": true
    },
    {
      "name": "scripted-permission-error-inline",
      "tex": "\\underline{\\sum_i^n}_a",
      "display": false
    },
    {
      "name": "scripted-permission-error-display",
      "tex": "\\underline{\\sum_i^n}_a",
      "display": true
    }
  ]
}.cases;
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
fs.writeFileSync(path.join(evidence||__dirname,'scripted_decoration_mathjax_3_2_2.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' exact AST and complete SVG cases');
