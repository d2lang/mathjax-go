// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Relax dispatch errors, macro/font/script boundaries and unchanged text/no-op controls.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_relax_command.cjs /path/to/pinned/d2latex [evidence-directory]"
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
    "bare",
    "\\relax",
    false
  ],
  [
    "bare",
    "\\relax",
    true
  ],
  [
    "plain",
    "x\\relax",
    false
  ],
  [
    "plain",
    "x\\relax",
    true
  ],
  [
    "following-sup",
    "x\\relax^n",
    false
  ],
  [
    "following-sup",
    "x\\relax^n",
    true
  ],
  [
    "following-sub",
    "x\\relax_i",
    false
  ],
  [
    "following-sub",
    "x\\relax_i",
    true
  ],
  [
    "unbraced-sup",
    "x^\\relax",
    false
  ],
  [
    "unbraced-sup",
    "x^\\relax",
    true
  ],
  [
    "unbraced-sub",
    "x_\\relax",
    false
  ],
  [
    "unbraced-sub",
    "x_\\relax",
    true
  ],
  [
    "group-sup",
    "x^{\\relax}",
    false
  ],
  [
    "group-sup",
    "x^{\\relax}",
    true
  ],
  [
    "group-sub",
    "x_{\\relax}",
    false
  ],
  [
    "group-sub",
    "x_{\\relax}",
    true
  ],
  [
    "fraction",
    "\\frac{\\relax}{y}",
    false
  ],
  [
    "fraction",
    "\\frac{\\relax}{y}",
    true
  ],
  [
    "sqrt",
    "\\sqrt{\\relax}",
    false
  ],
  [
    "sqrt",
    "\\sqrt{\\relax}",
    true
  ],
  [
    "font-wrapper",
    "\\mathbf{x\\relax}",
    false
  ],
  [
    "font-wrapper",
    "\\mathbf{x\\relax}",
    true
  ],
  [
    "font-declaration",
    "{\\bf x\\relax}",
    false
  ],
  [
    "font-declaration",
    "{\\bf x\\relax}",
    true
  ],
  [
    "ascii-prime",
    "x'\\relax^n",
    false
  ],
  [
    "ascii-prime",
    "x'\\relax^n",
    true
  ],
  [
    "unicode-prime",
    "x’\\relax^n",
    false
  ],
  [
    "unicode-prime",
    "x’\\relax^n",
    true
  ],
  [
    "overset",
    "\\overset{\\relax}{x}",
    false
  ],
  [
    "overset",
    "\\overset{\\relax}{x}",
    true
  ],
  [
    "stackrel-macro",
    "\\stackrel{\\relax}{=}",
    false
  ],
  [
    "stackrel-macro",
    "\\stackrel{\\relax}{=}",
    true
  ],
  [
    "pmod-macro",
    "\\pmod{\\relax}",
    false
  ],
  [
    "pmod-macro",
    "\\pmod{\\relax}",
    true
  ],
  [
    "control-name-boundary",
    "\\relaxation",
    false
  ],
  [
    "control-name-boundary",
    "\\relaxation",
    true
  ],
  [
    "empty-group",
    "\\relax{}",
    false
  ],
  [
    "empty-group",
    "\\relax{}",
    true
  ],
  [
    "text-literal",
    "\\text{\\relax}",
    false
  ],
  [
    "text-literal",
    "\\text{\\relax}",
    true
  ],
  [
    "mbox-literal",
    "\\mbox{\\relax}",
    false
  ],
  [
    "mbox-literal",
    "\\mbox{\\relax}",
    true
  ],
  [
    "mtext-literal",
    "\\mmlToken{mtext}{\\relax}",
    false
  ],
  [
    "mtext-literal",
    "\\mmlToken{mtext}{\\relax}",
    true
  ],
  [
    "text-embedded-math",
    "\\text{a $x\\relax$ b}",
    false
  ],
  [
    "text-embedded-math",
    "\\text{a $x\\relax$ b}",
    true
  ],
  [
    "comment",
    "% \\relax\nx",
    false
  ],
  [
    "comment",
    "% \\relax\nx",
    true
  ],
  [
    "ordinary",
    "x^n",
    false
  ],
  [
    "ordinary",
    "x^n",
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
  ],
  [
    "font-control",
    "\\mathbf{x}",
    false
  ],
  [
    "font-control",
    "\\mathbf{x}",
    true
  ],
  [
    "font-scope-control",
    "{\\bf x}",
    false
  ],
  [
    "font-scope-control",
    "{\\bf x}",
    true
  ],
  [
    "limits",
    "\\sum\\limits_i^n",
    false
  ],
  [
    "limits",
    "\\sum\\limits_i^n",
    true
  ],
  [
    "nolimits",
    "\\sum\\nolimits_i^n",
    false
  ],
  [
    "nolimits",
    "\\sum\\nolimits_i^n",
    true
  ],
  [
    "displaystyle",
    "{\\displaystyle x^n}",
    false
  ],
  [
    "displaystyle",
    "{\\displaystyle x^n}",
    true
  ],
  [
    "textstyle",
    "{\\textstyle x^n}",
    false
  ],
  [
    "textstyle",
    "{\\textstyle x^n}",
    true
  ],
  [
    "scriptstyle",
    "{\\scriptstyle x^n}",
    false
  ],
  [
    "scriptstyle",
    "{\\scriptstyle x^n}",
    true
  ],
  [
    "scriptscriptstyle",
    "{\\scriptscriptstyle x^n}",
    false
  ],
  [
    "scriptscriptstyle",
    "{\\scriptscriptstyle x^n}",
    true
  ],
  [
    "notag-control",
    "x\\notag",
    false
  ],
  [
    "nonumber-control",
    "x\\nonumber",
    false
  ]
];
const explicit=n=>({kind:n.kind,text:n.text,attributes:n.attributes,children:n.children.map(explicit)});
const records=[];
for(const [label,tex,display] of inputs){
 const name=label+(display?'-display':'-inline');
 const c=vm.createContext({console:{log(){},warn(){},error(){}}});
 for(const [f,b]of assets)vm.runInContext(b.toString(),c,{filename:f});c.request={tex,display};
 const result=vm.runInContext(`(()=>{let tree, fullTree;const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.coreIndex||0},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});const project=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?Object.fromEntries(n.attributes.getExplicitNames().sort().map(k=>[k,n.attributes.getExplicit(k)])):{},properties:n.getAllProperties?n.getAllProperties():{},isAccent:n.kind==='mo'?n.isAccent:null,coreParentKind:n.kind==='mo'?n.coreParent().kind:null,coreParentParentKind:n.kind==='mo'&&n.coreParent().parent?n.coreParent().parent.kind:null,children:n.childNodes.map(project)});const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){tree=project(math.root);fullTree=full(math.root);return original.call(this,math,doc)};const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));return{tree,svg,fullTree};})()`,c);
 // Expected invalid-operator cases retain complete primary error SVG and AST.
 if(evidence){fs.writeFileSync(path.join(evidence,name+'.svg'),result.svg);fs.writeFileSync(path.join(evidence,name+'.json'),JSON.stringify(result.tree,null,2)+'\n');fs.writeFileSync(path.join(evidence,name+'.full.json'),JSON.stringify(result.fullTree,null,2)+'\n');}
 records.push({name,tex,display,width:Math.ceil(Number(result.svg.match(/width="([\d.]+)ex"/)[1])*8),height:Math.ceil(Number(result.svg.match(/height="([\d.]+)ex"/)[1])*8),tree:explicit(result.tree),propertiesTree:result.tree,svgSHA256:hash(result.svg)});
}
fs.writeFileSync(path.join(evidence||__dirname,'relax_command_mathjax_3_2_2.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' exact AST and complete SVG cases');
