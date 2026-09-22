// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Explicit annotation script-attachment public contract, including primary rejection behavior; no product changes.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_annotation_attachment.cjs /path/to/pinned/d2latex [evidence-directory]"
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
const inputs = [["over-sub-sup", "\\overset{a}{x}_i^n", false], ["over-sub-sup", "\\overset{a}{x}_i^n", true], ["over-sup-sub", "\\overset{a}{x}^n_i", false], ["over-sup-sub", "\\overset{a}{x}^n_i", true], ["under-sub-sup", "\\underset{a}{x}_i^n", false], ["under-sub-sup", "\\underset{a}{x}_i^n", true], ["under-sup-sub", "\\underset{a}{x}^n_i", false], ["under-sup-sub", "\\underset{a}{x}^n_i", true], ["over-sub", "\\overset{a}{x}_i", false], ["over-sub", "\\overset{a}{x}_i", true], ["over-sup", "\\overset{a}{x}^n", false], ["over-sup", "\\overset{a}{x}^n", true], ["under-sub", "\\underset{a}{x}_i", false], ["under-sub", "\\underset{a}{x}_i", true], ["under-sup", "\\underset{a}{x}^n", false], ["under-sup", "\\underset{a}{x}^n", true], ["over-sum-sub-sup", "\\overset{a}{\\sum}_i^n", false], ["over-sum-sub-sup", "\\overset{a}{\\sum}_i^n", true], ["under-sum-sub-sup", "\\underset{a}{\\sum}_i^n", false], ["under-sum-sub-sup", "\\underset{a}{\\sum}_i^n", true], ["both-sub-sup", "\\overunderset{a}{b}{x}_i^n", false], ["both-sub-sup", "\\overunderset{a}{b}{x}_i^n", true], ["group-both-sub-sup", "{\\overunderset{a}{b}{x}}_i^n", false], ["group-both-sub-sup", "{\\overunderset{a}{b}{x}}_i^n", true], ["over-double-sub", "\\overset{a}{x}_i_j", false], ["over-double-sub", "\\overset{a}{x}_i_j", true], ["under-double-sup", "\\underset{a}{x}^i^j", false], ["under-double-sup", "\\underset{a}{x}^i^j", true], ["over-prime-sub", "\\overset{a}{x}'_i", false], ["over-prime-sub", "\\overset{a}{x}'_i", true], ["under-prime-sup", "\\underset{a}{x}'^n", false], ["under-prime-sup", "\\underset{a}{x}'^n", true], ["overbrace-sub-sup", "\\overbrace{x+y}_i^n", false], ["overbrace-sub-sup", "\\overbrace{x+y}_i^n", true], ["underbrace-sub-sup", "\\underbrace{x+y}_i^n", false], ["underbrace-sub-sup", "\\underbrace{x+y}_i^n", true], ["plain-scripts", "x_i^n", false], ["plain-scripts", "x_i^n", true], ["plain-double-sup", "x^i^j", false], ["plain-double-sup", "x^i^j", true], ["plain-prime-sup", "x'^n", false], ["plain-prime-sup", "x'^n", true], ["plain-prime-sub-sup", "x_i'^n", false], ["plain-prime-sub-sup", "x_i'^n", true], ["plain-double-prime", "x''", false], ["plain-double-prime", "x''", true], ["plain-prime-sup-sub", "x'^n_i", false], ["plain-prime-sup-sub", "x'^n_i", true]];
const explicit=n=>({kind:n.kind,text:n.text,attributes:n.attributes,children:n.children.map(explicit)});
const records=[];
for(const [label,tex,display] of inputs){
 const name=label+(display?'-display':'-inline');
 const c=vm.createContext({console:{log(){},warn(){},error(){}}});
 for(const [f,b]of assets)vm.runInContext(b.toString(),c,{filename:f});c.request={tex,display};
 const result=vm.runInContext(`(()=>{let tree, fullTree;const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.coreIndex||0},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});const project=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?Object.fromEntries(n.attributes.getExplicitNames().sort().map(k=>[k,n.attributes.getExplicit(k)])):{},properties:n.getAllProperties?n.getAllProperties():{},isAccent:n.kind==='mo'?n.isAccent:null,coreParentKind:n.kind==='mo'?n.coreParent().kind:null,coreParentParentKind:n.kind==='mo'&&n.coreParent().parent?n.coreParent().parent.kind:null,children:n.childNodes.map(project)});const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){tree=project(math.root);fullTree=full(math.root);return original.call(this,math,doc)};const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));return{tree,svg,fullTree};})()`,c);
 // Expected invalid-operator cases retain complete primary error SVG and AST.
 if(evidence){fs.writeFileSync(path.join(evidence,name+'.svg'),result.svg);fs.writeFileSync(path.join(evidence,name+'.json'),JSON.stringify(result.tree,null,2)+'\n');}
 records.push({name,tex,display,width:Math.ceil(Number(result.svg.match(/width="([\d.]+)ex"/)[1])*8),height:Math.ceil(Number(result.svg.match(/height="([\d.]+)ex"/)[1])*8),tree:explicit(result.tree),svgSHA256:hash(result.svg)});
}
fs.writeFileSync(path.join(evidence||__dirname,'annotation_attachment_mathjax_3_2_2.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' exact AST and complete SVG cases');
