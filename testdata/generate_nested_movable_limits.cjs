// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Nested movable-limit renderer contract; no dispatcher or source overlays.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_nested_movable_limits.cjs /path/to/pinned/d2latex [evidence-directory]"
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
const inputs = [["sum-both", "\\overunderset{a}{b}{\\sum_i^n}"], ["sum-under", "\\overunderset{a}{b}{\\sum_i}"], ["sum-over", "\\overunderset{a}{b}{\\sum^n}"], ["sum-grouped", "\\overunderset{a}{b}{{\\sum_i^n}}"], ["sum-double-grouped", "\\overunderset{a}{b}{{{\\sum_i^n}}}"], ["product-both", "\\overunderset{a}{b}{\\prod_i^n}"], ["product-grouped", "\\overunderset{a}{b}{{\\prod_i^n}}"], ["sum-row-context", "x+\\overunderset{a}{b}{\\sum_i^n}+y"], ["sum-font-wrapper", "\\overunderset{a}{b}{\\mathbf{\\sum_i^n}}"], ["sum-nested-base", "\\overunderset{c}{d}{\\overunderset{a}{b}{\\sum_i^n}}"], ["mathop-both", "\\overunderset{a}{b}{\\mathop{x}_i^n}"], ["integral-both", "\\overunderset{a}{b}{\\int_i^n}"], ["ordinary-script", "\\overunderset{a}{b}{x_i^n}"], ["fraction-base", "\\overunderset{a}{b}{\\frac{x}{y}}"], ["ordinary-base", "\\overunderset{a}{b}{x}"], ["bare-sum", "\\overunderset{a}{b}{\\sum}"], ["grouped-bare-sum", "\\overunderset{a}{b}{{\\sum}}"], ["ordinary-sum-control", "\\sum_i^n"], ["ordinary-product-control", "\\prod_i^n"], ["ordinary-integral-control", "\\int_i^n"]];
const records=[];
for(const [label,tex] of inputs)for(const display of [false,true]){
 const name=label+(display?'-display':'-inline');
 const c=vm.createContext({console:{log(){},warn(){},error(){}}});
 for(const [f,b]of assets)vm.runInContext(b.toString(),c,{filename:f});c.request={tex,display};
 const result=vm.runInContext(`(()=>{let tree, fullTree;const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.isEmbellished&&n.core()!==n?n.childNodes.indexOf(n.core()):(typeof n.coreIndex==='function'?n.coreIndex():n.coreIndex||0)},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});const project=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?Object.fromEntries(n.attributes.getExplicitNames().sort().map(k=>[k,n.attributes.getExplicit(k)])):{},properties:n.getAllProperties?n.getAllProperties():{},isAccent:n.kind==='mo'?n.isAccent:null,coreParentKind:n.kind==='mo'?n.coreParent().kind:null,coreParentParentKind:n.kind==='mo'&&n.coreParent().parent?n.coreParent().parent.kind:null,children:n.childNodes.map(project)});const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){tree=project(math.root);fullTree=full(math.root);return original.call(this,math,doc)};const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));return{tree,svg,fullTree};})()`,c);
 if(result.svg.includes('data-mml-node="merror"'))throw Error(name+' error');
 if(evidence){fs.writeFileSync(path.join(evidence,name+'.svg'),result.svg);fs.writeFileSync(path.join(evidence,name+'.json'),JSON.stringify(result.tree,null,2)+'\n');}
 records.push({name,tex,display,width:Math.ceil(Number(result.svg.match(/width="([\d.]+)ex"/)[1])*8),height:Math.ceil(Number(result.svg.match(/height="([\d.]+)ex"/)[1])*8),tree:result.tree,fullTree:result.fullTree,svgSHA256:hash(result.svg)});
}
fs.writeFileSync(path.join(evidence,'discovery.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' exact AST and complete SVG cases');
