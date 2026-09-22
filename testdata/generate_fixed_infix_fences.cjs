// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Fixed infix fences, inherited styles, unchanged genfrac/binomial and error boundaries.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_fixed_infix_fences.cjs /path/to/pinned/d2latex [evidence-directory]"
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
    "choose",
    "x\\choose y",
    false
  ],
  [
    "choose",
    "x\\choose y",
    true
  ],
  [
    "brace",
    "x\\brace y",
    false
  ],
  [
    "brace",
    "x\\brace y",
    true
  ],
  [
    "brack",
    "x\\brack y",
    false
  ],
  [
    "brack",
    "x\\brack y",
    true
  ],
  [
    "tall",
    "\\frac{a}{b}\\choose\\sqrt{x+y}",
    false
  ],
  [
    "tall",
    "\\frac{a}{b}\\choose\\sqrt{x+y}",
    true
  ],
  [
    "displaystyle",
    "\\displaystyle{x\\choose y}",
    false
  ],
  [
    "displaystyle",
    "\\displaystyle{x\\choose y}",
    true
  ],
  [
    "textstyle",
    "\\textstyle{x\\choose y}",
    false
  ],
  [
    "textstyle",
    "\\textstyle{x\\choose y}",
    true
  ],
  [
    "scriptstyle",
    "\\scriptstyle{x\\choose y}",
    false
  ],
  [
    "scriptstyle",
    "\\scriptstyle{x\\choose y}",
    true
  ],
  [
    "scriptscriptstyle",
    "\\scriptscriptstyle{x\\choose y}",
    false
  ],
  [
    "scriptscriptstyle",
    "\\scriptscriptstyle{x\\choose y}",
    true
  ],
  [
    "superscript",
    "z^{x\\choose y}",
    false
  ],
  [
    "superscript",
    "z^{x\\choose y}",
    true
  ],
  [
    "subscript",
    "z_{x\\choose y}",
    false
  ],
  [
    "subscript",
    "z_{x\\choose y}",
    true
  ],
  [
    "nested-script",
    "z_{w_{x\\choose y}}",
    false
  ],
  [
    "nested-script",
    "z_{w_{x\\choose y}}",
    true
  ],
  [
    "surrounding-spacing",
    "a+{x\\choose y}=b",
    false
  ],
  [
    "surrounding-spacing",
    "a+{x\\choose y}=b",
    true
  ],
  [
    "operator-limit",
    "\\sum_{x\\choose y}",
    false
  ],
  [
    "operator-limit",
    "\\sum_{x\\choose y}",
    true
  ],
  [
    "nested-infix",
    "{x\\choose y}\\choose z",
    false
  ],
  [
    "nested-infix",
    "{x\\choose y}\\choose z",
    true
  ],
  [
    "outer-left-right",
    "\\left[{x\\choose y}\\right]",
    false
  ],
  [
    "outer-left-right",
    "\\left[{x\\choose y}\\right]",
    true
  ],
  [
    "over-control",
    "x\\over y",
    false
  ],
  [
    "over-control",
    "x\\over y",
    true
  ],
  [
    "atop-control",
    "x\\atop y",
    false
  ],
  [
    "atop-control",
    "x\\atop y",
    true
  ],
  [
    "above-control",
    "x\\above 1pt y",
    false
  ],
  [
    "above-control",
    "x\\above 1pt y",
    true
  ],
  [
    "genfrac-zero-control",
    "\\genfrac{(}{)}{0}{}{x}{y}",
    false
  ],
  [
    "genfrac-zero-control",
    "\\genfrac{(}{)}{0}{}{x}{y}",
    true
  ],
  [
    "genfrac-rule-control",
    "\\genfrac{[}{]}{1pt}{1}{x}{y}",
    false
  ],
  [
    "genfrac-rule-control",
    "\\genfrac{[}{]}{1pt}{1}{x}{y}",
    true
  ],
  [
    "genfrac-empty-control",
    "\\genfrac{}{)}{0}{2}{x}{y}",
    false
  ],
  [
    "genfrac-empty-control",
    "\\genfrac{}{)}{0}{2}{x}{y}",
    true
  ],
  [
    "binom-boundary",
    "\\binom{x}{y}",
    false
  ],
  [
    "binom-boundary",
    "\\binom{x}{y}",
    true
  ],
  [
    "dbinom-boundary",
    "\\dbinom{x}{y}",
    false
  ],
  [
    "dbinom-boundary",
    "\\dbinom{x}{y}",
    true
  ],
  [
    "tbinom-boundary",
    "\\tbinom{x}{y}",
    false
  ],
  [
    "tbinom-boundary",
    "\\tbinom{x}{y}",
    true
  ],
  [
    "ambiguous-boundary",
    "x\\choose y\\choose z",
    false
  ],
  [
    "ambiguous-boundary",
    "x\\choose y\\choose z",
    true
  ],
  [
    "bad-style-control",
    "\\genfrac{(}{)}{0}{4}{x}{y}",
    false
  ],
  [
    "bad-style-control",
    "\\genfrac{(}{)}{0}{4}{x}{y}",
    true
  ]
];
const explicit=n=>({kind:n.kind,text:n.text,attributes:n.attributes,children:n.children.map(explicit)});
const records=[];
for(const [label,tex,display] of inputs){
 const name=label+(display?'-display':'-inline');
 const c=vm.createContext({console:{log(){},warn(){},error(){}}});
 for(const [f,b]of assets)vm.runInContext(b.toString(),c,{filename:f});c.request={tex,display};
 const result=vm.runInContext(`(()=>{let tree, fullTree;const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.coreIndex||0},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});const project=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?Object.fromEntries(n.attributes.getExplicitNames().sort().map(k=>[k,n.attributes.getExplicit(k)])):{},properties:n.getAllProperties?n.getAllProperties():{},isAccent:n.kind==='mo'?n.isAccent:null,coreParentKind:n.kind==='mo'?n.coreParent().kind:null,coreParentParentKind:n.kind==='mo'&&n.coreParent().parent?n.coreParent().parent.kind:null,children:n.childNodes.map(project)});const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){tree=project(math.root);fullTree=full(math.root);return original.call(this,math,doc)};const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));return{tree,svg,fullTree};})()`,c);
 // Separate error observations retain the complete unmodified primary SVG and AST.
 if(evidence){fs.writeFileSync(path.join(evidence,name+'.svg'),result.svg);fs.writeFileSync(path.join(evidence,name+'.json'),JSON.stringify(result.tree,null,2)+'\n');fs.writeFileSync(path.join(evidence,name+'.full.json'),JSON.stringify(result.fullTree,null,2)+'\n');}
 records.push({name,tex,display,width:Math.ceil(Number(result.svg.match(/width="([\d.]+)ex"/)[1])*8),height:Math.ceil(Number(result.svg.match(/height="([\d.]+)ex"/)[1])*8),tree:explicit(result.tree),propertiesTree:result.tree,svgSHA256:hash(result.svg)});
}
const fixture={oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records.slice(0,48)};
fs.writeFileSync(path.join(evidence||__dirname,'fixed_infix_fences_mathjax_3_2_2.json'),JSON.stringify(fixture,null,2)+'\n');
if(evidence)fs.writeFileSync(path.join(evidence,'error-probes.json'),JSON.stringify({...fixture,cases:records.slice(48)},null,2)+'\n');
console.log('48 exact-reference cases plus 4 separate error observations');
