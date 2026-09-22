// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Accent character-base/script placement and exact inherited underline-error controls.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_accent_script_base.cjs /path/to/pinned/d2latex [evidence-directory]"
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
const inputs = [["accent-script-0", "\\vec{y}^n", false], ["accent-script-0", "\\vec{y}^n", true], ["accent-script-1", "\\bf\\vec{y}^n", false], ["accent-script-1", "\\bf\\vec{y}^n", true], ["accent-script-2", "\\vec{x}^n", false], ["accent-script-2", "\\vec{x}^n", true], ["accent-script-3", "\\hat{x}^n", false], ["accent-script-3", "\\hat{x}^n", true], ["accent-script-4", "\\bar{x}^n", false], ["accent-script-4", "\\bar{x}^n", true], ["accent-script-5", "\\overset{a}{x}^n", false], ["accent-script-5", "\\overset{a}{x}^n", true], ["accent-script-6", "\\vec{y}_i", false], ["accent-script-6", "\\vec{y}_i", true], ["accent-script-7", "\\vec{y}", false], ["accent-script-7", "\\vec{y}", true], ["accent-script-8", "\\bf\\vec{y}", false], ["accent-script-8", "\\bf\\vec{y}", true], ["accent-script-9", "y^n", false], ["accent-script-9", "y^n", true], ["nested-hat-vec", "\\hat{\\vec{x}}^n", false], ["nested-hat-vec", "\\hat{\\vec{x}}^n", true], ["nested-vec-hat", "\\vec{\\hat{x}}_i", false], ["nested-vec-hat", "\\vec{\\hat{x}}_i", true], ["both-scripts", "\\vec{x}_i^n", false], ["both-scripts", "\\vec{x}_i^n", true], ["multichar", "\\hat{xy}^n", false], ["multichar", "\\hat{xy}^n", true], ["group", "\\vec{{x}}^n", false], ["group", "\\vec{{x}}^n", true], ["scaled-base", "\\vec{\\scriptstyle x}^n", false], ["scaled-base", "\\vec{\\scriptstyle x}^n", true], ["scaled-accent", "{\\large\\vec{x}}^n", false], ["scaled-accent", "{\\large\\vec{x}}^n", true], ["nested-line", "\\overline{\\vec{x}}^n", false], ["nested-line", "\\overline{\\vec{x}}^n", true], ["nested-bar", "\\hat{\\bar{x}}^n", false], ["nested-bar", "\\hat{\\bar{x}}^n", true], ["under-stack", "\\underset{a}{x}^n", false], ["under-stack", "\\underset{a}{x}^n", true], ["fixed-delimiter", "\\hat{\\bigr)}^n", false], ["fixed-delimiter", "\\hat{\\bigr)}^n", true], ["large-op", "\\hat{\\sum}^n", false], ["large-op", "\\hat{\\sum}^n", true], ["bar-both", "\\bar{x}_i^n", false], ["bar-both", "\\bar{x}_i^n", true], ["bar-multichar", "\\bar{xy}^n", false], ["bar-multichar", "\\bar{xy}^n", true], ["ordinary-both", "x_i^n", false], ["ordinary-both", "x_i^n", true], ["nested-under", "\\underline{\\hat{x}}_i", false], ["nested-under", "\\underline{\\hat{x}}_i", true]];
const explicit=n=>({kind:n.kind,text:n.text,attributes:n.attributes,children:n.children.map(explicit)});
const records=[];
for(const [label,tex,display] of inputs){
 const name=label+(display?'-display':'-inline');
 const c=vm.createContext({console:{log(){},warn(){},error(){}}});
 for(const [f,b]of assets)vm.runInContext(b.toString(),c,{filename:f});c.request={tex,display};
 const result=vm.runInContext(`(()=>{let tree, fullTree;const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.coreIndex||0},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});const project=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?Object.fromEntries(n.attributes.getExplicitNames().sort().map(k=>[k,n.attributes.getExplicit(k)])):{},properties:n.getAllProperties?n.getAllProperties():{},isAccent:n.kind==='mo'?n.isAccent:null,coreParentKind:n.kind==='mo'?n.coreParent().kind:null,coreParentParentKind:n.kind==='mo'&&n.coreParent().parent?n.coreParent().parent.kind:null,children:n.childNodes.map(project)});const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){tree=project(math.root);fullTree=full(math.root);return original.call(this,math,doc)};const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));return{tree,svg,fullTree};})()`,c);
 // Expected invalid-operator cases retain complete primary error SVG and AST.
 if(evidence){fs.writeFileSync(path.join(evidence,name+'.svg'),result.svg);fs.writeFileSync(path.join(evidence,name+'.json'),JSON.stringify(result.tree,null,2)+'\n');}
 records.push({name,tex,display,width:Math.ceil(Number(result.svg.match(/width="([\d.]+)ex"/)[1])*8),height:Math.ceil(Number(result.svg.match(/height="([\d.]+)ex"/)[1])*8),tree:explicit(result.tree),propertiesTree:result.tree,svgSHA256:hash(result.svg)});
}
fs.writeFileSync(path.join(evidence||__dirname,'accent_script_base_mathjax_3_2_2.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' exact AST and complete SVG cases');
