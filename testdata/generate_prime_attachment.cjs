// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// ASCII pending-prime attachment plus separately retained Unicode dispatch controls.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_prime_attachment.cjs /path/to/pinned/d2latex [evidence-directory]"
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
const inputs = [["single", "x'", false], ["single", "x'", true], ["double", "x''", false], ["double", "x''", true], ["triple", "x'''", false], ["triple", "x'''", true], ["four", "x''''", false], ["four", "x''''", true], ["five", "x'''''", false], ["five", "x'''''", true], ["prime-sup", "x'^n", false], ["prime-sup", "x'^n", true], ["double-sup", "x''^n", false], ["double-sup", "x''^n", true], ["prime-sup-sub", "x'^n_i", false], ["prime-sup-sub", "x'^n_i", true], ["prime-sub-sup", "x'_i^n", false], ["prime-sub-sup", "x'_i^n", true], ["sub-prime-sup", "x_i'^n", false], ["sub-prime-sup", "x_i'^n", true], ["prime-sub", "x'_i", false], ["prime-sub", "x'_i", true], ["sub-prime", "x_i'", false], ["sub-prime", "x_i'", true], ["prime-group-sup", "x'^{n+1}", false], ["prime-group-sup", "x'^{n+1}", true], ["group-prime-sup", "{x'}^n", false], ["group-prime-sup", "{x'}^n", true], ["prime-empty-sup", "x'^{}", false], ["prime-empty-sup", "x'^{}", true], ["bare-prime-sup", "'^n", false], ["bare-prime-sup", "'^n", true], ["prime-spaced", "x' ^n", false], ["prime-spaced", "x' ^n", true], ["prime-comment", "x'%comment\n^n", false], ["prime-comment", "x'%comment\n^n", true], ["prime-ordinary-then-sup", "x'y^n", false], ["prime-ordinary-then-sup", "x'y^n", true], ["prime-double-sup-error", "x'^i^j", false], ["prime-double-sup-error", "x'^i^j", true], ["sup-prime-error", "x^n'", false], ["sup-prime-error", "x^n'", true], ["scripts-prime-error", "x_i^n'", false], ["scripts-prime-error", "x_i^n'", true], ["double-sub-error", "x'_i_j", false], ["double-sub-error", "x'_i_j", true], ["sup-prime-spaced-error", "x^n '", false], ["sup-prime-spaced-error", "x^n '", true], ["missing-sup-error", "x'^", false], ["missing-sup-error", "x'^", true], ["over-prime-sup", "\\overset{a}{x}'^n", false], ["over-prime-sup", "\\overset{a}{x}'^n", true], ["under-prime-sup", "\\underset{a}{x}'^n", false], ["under-prime-sup", "\\underset{a}{x}'^n", true], ["over-prime-sub", "\\overset{a}{x}'_i", false], ["over-prime-sub", "\\overset{a}{x}'_i", true], ["under-prime-sub", "\\underset{a}{x}'_i", false], ["under-prime-sub", "\\underset{a}{x}'_i", true], ["sum-prime-sup", "\\sum'^n", false], ["sum-prime-sup", "\\sum'^n", true], ["sum-sub-prime-sup", "\\sum_i'^n", false], ["sum-sub-prime-sup", "\\sum_i'^n", true], ["sum-prime-sub", "\\sum'_i", false], ["sum-prime-sub", "\\sum'_i", true], ["sum-prime-limit-error", "\\sum'\\limits", false], ["sum-prime-limit-error", "\\sum'\\limits", true], ["sum-consumed-prime-sub-limits", "\\sum'_i\\limits", false], ["sum-consumed-prime-sub-limits", "\\sum'_i\\limits", true], ["sum-consumed-prime-sup-limits", "\\sum'^n\\limits", false], ["sum-consumed-prime-sup-limits", "\\sum'^n\\limits", true], ["sum-group-prime-limits", "{\\sum'}\\limits", false], ["sum-group-prime-limits", "{\\sum'}\\limits", true], ["sum-limits-prime-sup", "\\sum\\limits'^n", false], ["sum-limits-prime-sup", "\\sum\\limits'^n", true], ["sum-nolimits-prime-sup", "\\sum\\nolimits'^n", false], ["sum-nolimits-prime-sup", "\\sum\\nolimits'^n", true], ["prime-function", "\\sin'^n x", false], ["prime-function", "\\sin'^n x", true], ["grouped-base-prime", "{x+y}'^n", false], ["grouped-base-prime", "{x+y}'^n", true], ["nested-subscript-prime", "z_{x'^n}", false], ["nested-subscript-prime", "z_{x'^n}", true], ["nested-superscript-prime", "z^{x'^n}", false], ["nested-superscript-prime", "z^{x'^n}", true], ["styled-prime", "\\mathbf{x}'^n", false], ["styled-prime", "\\mathbf{x}'^n", true], ["literal-prime-symbol", "x\\prime^n", false], ["literal-prime-symbol", "x\\prime^n", true], ["unicode-right-prime", "x’^n", false], ["unicode-right-prime", "x’^n", true], ["mixed-prime", "x'’^n", false], ["mixed-prime", "x'’^n", true], ["ordinary-scripts", "x_i^n", false], ["ordinary-scripts", "x_i^n", true], ["ordinary-double-sup-error", "x^i^j", false], ["ordinary-double-sup-error", "x^i^j", true]];
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
fs.writeFileSync(path.join(evidence||__dirname,'prime_attachment_mathjax_3_2_2.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' exact AST and complete SVG cases');
