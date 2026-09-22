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
    "usage: node generate_prime_lifetime.cjs /path/to/pinned/d2latex [evidence-directory]"
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
const inputs = [["life-immediate", "x'^n", false], ["life-immediate", "x'^n", true], ["life-space", "x' ^n", false], ["life-space", "x' ^n", true], ["life-comment", "x'% hi\n^n", false], ["life-comment", "x'% hi\n^n", true], ["life-notag", "x'\\notag^n", false], ["life-nonumber", "x'\\nonumber^n", false], ["life-label", "x'\\label{a}^n", false], ["life-label", "x'\\label{a}^n", true], ["life-empty-macro", "\\newcommand{\\noop}{}x'\\noop^n", false], ["life-empty-macro", "\\newcommand{\\noop}{}x'\\noop^n", true], ["life-empty-group", "x'{}^n", false], ["life-empty-group", "x'{}^n", true], ["life-ordinary", "x'y^n", false], ["life-ordinary", "x'y^n", true], ["life-spacing", "x'\\,^n", false], ["life-spacing", "x'\\,^n", true], ["life-group-close", "{x'}^n", false], ["life-group-close", "{x'}^n", true], ["life-left-right", "\\left(x'\\right)^n", false], ["life-left-right", "\\left(x'\\right)^n", true], ["life-fraction", "x'\\over y", false], ["life-fraction", "x'\\over y", true], ["life-style", "x'\\displaystyle ^n", false], ["life-style", "x'\\displaystyle ^n", true], ["life-font", "x'\\bf ^n", false], ["life-font", "x'\\bf ^n", true], ["life-size", "x'\\large ^n", false], ["life-size", "x'\\large ^n", true], ["life-color", "x'\\color{red}^n", false], ["life-color", "x'\\color{red}^n", true], ["life-repeat-space", "x' '^n", false], ["life-repeat-space", "x' '^n", true], ["life-repeat-empty-macro", "\\newcommand{\\noop}{}x'\\noop'^n", false], ["life-repeat-empty-macro", "\\newcommand{\\noop}{}x'\\noop'^n", true], ["life-pending-limits", "\\sum'\\notag\\limits", false], ["life-pending-limits", "\\sum'\\notag\\limits", true], ["life-consumed-sub-limits", "\\sum'_i\\limits", false], ["life-consumed-sub-limits", "\\sum'_i\\limits", true], ["life-consumed-sup-limits", "\\sum'^n\\limits", false], ["life-consumed-sup-limits", "\\sum'^n\\limits", true], ["life-subprime", "x_i'^n", false], ["life-subprime", "x_i'^n", true], ["life-supprime", "x^i'", false], ["life-supprime", "x^i'", true], ["life-limits-prime", "\\sum\\limits'^n", false], ["life-limits-prime", "\\sum\\limits'^n", true], ["life-authored-over-prime", "\\overset{a}{x}'^n", false], ["life-authored-over-prime", "\\overset{a}{x}'^n", true], ["life-authored-under-prime", "\\underset{a}{x}'_n", false], ["life-authored-under-prime", "\\underset{a}{x}'_n", true], ["life-function-prime", "\\sin'^n x", false], ["life-function-prime", "\\sin'^n x", true], ["font-font-both", "x'\\bf^n_i", false], ["font-font-both", "x'\\bf^n_i", true], ["font-font-reverse", "x'\\bf_i^n", false], ["font-font-reverse", "x'\\bf_i^n", true], ["font-font-ordinary", "x'\\bf^n y_i", false], ["font-font-ordinary", "x'\\bf^n y_i", true], ["font-font-nested", "x'\\bf^{\\rm n}_i", false], ["font-font-nested", "x'\\bf^{\\rm n}_i", true], ["font-font-repeat", "x'\\bf\\rm^n_i", false], ["font-font-repeat", "x'\\bf\\rm^n_i", true], ["font-group-isolation", "{x'\\bf^n_i}y", false], ["font-group-isolation", "{x'\\bf^n_i}y", true]];
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
fs.writeFileSync(path.join(evidence||__dirname,'prime_lifetime_mathjax_3_2_2.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' exact AST and complete SVG cases');
