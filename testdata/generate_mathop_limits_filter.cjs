// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Explicit limits public contract, including primary rejection behavior; no product changes.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_explicit_limits.cjs /path/to/pinned/d2latex [evidence-directory]"
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
const inputs = [["mathop-default", "\\mathop{x}_i^n", false], ["mathop-default", "\\mathop{x}_i^n", true], ["mathop-ordinary", "\\mathop{x}", false], ["mathop-ordinary", "\\mathop{x}", true], ["mathop-group-default", "{\\mathop{x}_i^n}", false], ["mathop-group-default", "{\\mathop{x}_i^n}", true], ["nested-mathop-default", "\\overunderset{a}{b}{\\mathop{x}_i^n}", false], ["nested-mathop-default", "\\overunderset{a}{b}{\\mathop{x}_i^n}", true], ["group-mathop-under-limits", "{\\mathop{x}_i}\\limits^n", false], ["group-mathop-under-limits", "{\\mathop{x}_i}\\limits^n", true], ["group-mathop-under-no", "{\\mathop{x}_i}\\nolimits^n", false], ["group-mathop-under-no", "{\\mathop{x}_i}\\nolimits^n", true], ["group-mathop-both-limits", "{\\mathop{x}_i^n}\\limits", false], ["group-mathop-both-limits", "{\\mathop{x}_i^n}\\limits", true], ["group-mathop-both-no", "{\\mathop{x}_i^n}\\nolimits", false], ["group-mathop-both-no", "{\\mathop{x}_i^n}\\nolimits", true], ["annotation-group-mathop", "{\\overset{a}{\\mathop{x}_i}}\\limits^n", false], ["annotation-group-mathop", "{\\overset{a}{\\mathop{x}_i}}\\limits^n", true], ["annotation-group-mathop-no", "{\\underset{a}{\\mathop{x}^n}}\\nolimits_i", false], ["annotation-group-mathop-no", "{\\underset{a}{\\mathop{x}^n}}\\nolimits_i", true], ["mathop-double-switch", "\\mathop{x}_i\\limits\\nolimits^n", false], ["mathop-double-switch", "\\mathop{x}_i\\limits\\nolimits^n", true], ["mathop-reverse-switch", "\\mathop{x}^n\\nolimits\\limits_i", false], ["mathop-reverse-switch", "\\mathop{x}^n\\nolimits\\limits_i", true], ["mathop-style-inline", "\\textstyle\\mathop{x}_i^n", false], ["mathop-style-inline", "\\textstyle\\mathop{x}_i^n", true], ["mathop-style-display", "\\displaystyle\\mathop{x}_i^n", false], ["mathop-style-display", "\\displaystyle\\mathop{x}_i^n", true]];
const records=[];
for(const [label,tex,display] of inputs){
 const name=label+(display?'-display':'-inline');
 const c=vm.createContext({console:{log(){},warn(){},error(){}}});
 for(const [f,b]of assets)vm.runInContext(b.toString(),c,{filename:f});c.request={tex,display};
 const result=vm.runInContext(`(()=>{let tree, fullTree;const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.coreIndex||0},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});const project=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?Object.fromEntries(n.attributes.getExplicitNames().sort().map(k=>[k,n.attributes.getExplicit(k)])):{},properties:n.getAllProperties?n.getAllProperties():{},isAccent:n.kind==='mo'?n.isAccent:null,coreParentKind:n.kind==='mo'?n.coreParent().kind:null,coreParentParentKind:n.kind==='mo'&&n.coreParent().parent?n.coreParent().parent.kind:null,children:n.childNodes.map(project)});const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){tree=project(math.root);fullTree=full(math.root);return original.call(this,math,doc)};const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));return{tree,svg,fullTree};})()`,c);
 // Expected invalid-operator cases retain complete primary error SVG and AST.
 if(evidence){fs.writeFileSync(path.join(evidence,name+'.svg'),result.svg);fs.writeFileSync(path.join(evidence,name+'.json'),JSON.stringify(result.tree,null,2)+'\n');}
 records.push({name,tex,display,width:Math.ceil(Number(result.svg.match(/width="([\d.]+)ex"/)[1])*8),height:Math.ceil(Number(result.svg.match(/height="([\d.]+)ex"/)[1])*8),tree:result.tree,fullTree:result.fullTree,svgSHA256:hash(result.svg)});
}
fs.writeFileSync(path.join(__dirname,'mathop_limits_filter_mathjax_3_2_2.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' exact AST and complete SVG cases');
