// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// D074 mod argument/style references with exact surrounding controls.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_mod_command.cjs /path/to/pinned/d2latex [evidence-directory]"
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
    "plain",
    "\\mod{x}",
    false
  ],
  [
    "plain",
    "\\mod{x}",
    true
  ],
  [
    "unbraced",
    "\\mod x",
    false
  ],
  [
    "unbraced",
    "\\mod x",
    true
  ],
  [
    "missing",
    "\\mod",
    false
  ],
  [
    "missing",
    "\\mod",
    true
  ],
  [
    "empty",
    "\\mod{}",
    false
  ],
  [
    "empty",
    "\\mod{}",
    true
  ],
  [
    "group",
    "\\mod{{a+b}}",
    false
  ],
  [
    "group",
    "\\mod{{a+b}}",
    true
  ],
  [
    "control-argument",
    "\\mod\\frac{a}{b}",
    false
  ],
  [
    "control-argument",
    "\\mod\\frac{a}{b}",
    true
  ],
  [
    "row",
    "x\\mod n",
    false
  ],
  [
    "row",
    "x\\mod n",
    true
  ],
  [
    "relation",
    "a=b\\mod{n}",
    false
  ],
  [
    "relation",
    "a=b\\mod{n}",
    true
  ],
  [
    "binary",
    "a+\\mod{n}+b",
    false
  ],
  [
    "binary",
    "a+\\mod{n}+b",
    true
  ],
  [
    "displaystyle",
    "\\displaystyle\\mod{x}",
    false
  ],
  [
    "displaystyle",
    "\\displaystyle\\mod{x}",
    true
  ],
  [
    "textstyle",
    "\\textstyle\\mod{x}",
    false
  ],
  [
    "textstyle",
    "\\textstyle\\mod{x}",
    true
  ],
  [
    "scriptstyle",
    "\\scriptstyle\\mod{x}",
    false
  ],
  [
    "scriptstyle",
    "\\scriptstyle\\mod{x}",
    true
  ],
  [
    "scriptscriptstyle",
    "\\scriptscriptstyle\\mod{x}",
    false
  ],
  [
    "scriptscriptstyle",
    "\\scriptscriptstyle\\mod{x}",
    true
  ],
  [
    "superscript",
    "x^{\\mod{n}}",
    false
  ],
  [
    "superscript",
    "x^{\\mod{n}}",
    true
  ],
  [
    "subscript",
    "x_{\\mod{n}}",
    false
  ],
  [
    "subscript",
    "x_{\\mod{n}}",
    true
  ],
  [
    "nested-script",
    "x^{y_{\\mod{n}}}",
    false
  ],
  [
    "nested-script",
    "x^{y_{\\mod{n}}}",
    true
  ],
  [
    "nested-mod",
    "\\mod{a+\\mod{b}}",
    false
  ],
  [
    "nested-mod",
    "\\mod{a+\\mod{b}}",
    true
  ],
  [
    "tail",
    "\\mod xz",
    false
  ],
  [
    "tail",
    "\\mod xz",
    true
  ],
  [
    "declared-override",
    "\\DeclareMathOperator{\\mod}{pick}x\\mod y",
    false
  ],
  [
    "declared-override",
    "\\DeclareMathOperator{\\mod}{pick}x\\mod y",
    true
  ],
  [
    "bmod-control",
    "x\\bmod n",
    false
  ],
  [
    "bmod-control",
    "x\\bmod n",
    true
  ],
  [
    "parentheses-control",
    "(x+y)",
    false
  ],
  [
    "parentheses-control",
    "(x+y)",
    true
  ],
  [
    "fraction-control",
    "\\frac{x}{y}",
    false
  ],
  [
    "fraction-control",
    "\\frac{x}{y}",
    true
  ]
];
const explicit=n=>({kind:n.kind,text:n.text,attributes:n.attributes,children:n.children.map(explicit)});
const records=[];
for(const [label,tex,display] of inputs){
 const name=label+(display?'-display':'-inline');
 const c=vm.createContext({console:{log(){},warn(){},error(){}}});
 for(const [f,b]of assets)vm.runInContext(b.toString(),c,{filename:f});c.request={tex,display};
 const result=vm.runInContext(`(()=>{let tree, fullTree;const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.coreIndex||0},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});const project=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?Object.fromEntries(n.attributes.getExplicitNames().sort().map(k=>[k,n.attributes.getExplicit(k)])):{},properties:n.getAllProperties?n.getAllProperties():{},isAccent:n.kind==='mo'?n.isAccent:null,coreParentKind:n.kind==='mo'?n.coreParent().kind:null,coreParentParentKind:n.kind==='mo'&&n.coreParent().parent?n.coreParent().parent.kind:null,children:n.childNodes.map(project)});const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){tree=JSON.parse(JSON.stringify(project(math.root)));fullTree=full(math.root);return original.call(this,math,doc)};const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));return{tree,svg,fullTree};})()`,c);
 // Expected invalid-operator cases retain complete primary error SVG and AST.
 if(evidence){fs.writeFileSync(path.join(evidence,name+'.svg'),result.svg);fs.writeFileSync(path.join(evidence,name+'.json'),JSON.stringify(result.tree,null,2)+'\n');fs.writeFileSync(path.join(evidence,name+'.full.json'),JSON.stringify(result.fullTree,null,2)+'\n');}
 records.push({name,tex,display,width:Math.ceil(Number(result.svg.match(/width="([\d.]+)ex"/)[1])*8),height:Math.ceil(Number(result.svg.match(/height="([\d.]+)ex"/)[1])*8),tree:explicit(result.tree),propertiesTree:result.tree,svgSHA256:hash(result.svg)});
}
fs.writeFileSync(path.join(evidence||__dirname,'mod_command_mathjax_3_2_2.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' exact AST and complete SVG cases');
