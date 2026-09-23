// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Registered pod/pmod forwarding, argument, error and isolated-state controls.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_pod_macros.cjs /path/to/pinned/d2latex [evidence-directory]"
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
    "pod-direct",
    "\\pod{x}",
    false,
    {
      "name": "pod",
      "body": "#1+#1",
      "arguments": 1
    }
  ],
  [
    "pod-direct",
    "\\pod{x}",
    true,
    {
      "name": "pod",
      "body": "#1+#1",
      "arguments": 1
    }
  ],
  [
    "pod-through-pmod",
    "\\pmod{x}",
    false,
    {
      "name": "pod",
      "body": "#1+#1",
      "arguments": 1
    }
  ],
  [
    "pod-through-pmod",
    "\\pmod{x}",
    true,
    {
      "name": "pod",
      "body": "#1+#1",
      "arguments": 1
    }
  ],
  [
    "pmod-direct",
    "\\pmod{x}",
    false,
    {
      "name": "pmod",
      "body": "#1+#1",
      "arguments": 1
    }
  ],
  [
    "pmod-direct",
    "\\pmod{x}",
    true,
    {
      "name": "pmod",
      "body": "#1+#1",
      "arguments": 1
    }
  ],
  [
    "pmod-does-not-change-pod",
    "\\pod{x}",
    false,
    {
      "name": "pmod",
      "body": "#1+#1",
      "arguments": 1
    }
  ],
  [
    "pmod-does-not-change-pod",
    "\\pod{x}",
    true,
    {
      "name": "pmod",
      "body": "#1+#1",
      "arguments": 1
    }
  ],
  [
    "nested-forward",
    "\\probe{x}",
    false,
    {
      "name": "probe",
      "body": "\\pmod{#1}",
      "arguments": 1
    }
  ],
  [
    "nested-forward",
    "\\probe{x}",
    true,
    {
      "name": "probe",
      "body": "\\pmod{#1}",
      "arguments": 1
    }
  ],
  [
    "pod-zero-arguments",
    "\\pmod{x}",
    false,
    {
      "name": "pod",
      "body": "q",
      "arguments": 0
    }
  ],
  [
    "pod-zero-arguments",
    "\\pmod{x}",
    true,
    {
      "name": "pod",
      "body": "q",
      "arguments": 0
    }
  ],
  [
    "pmod-zero-arguments",
    "\\pmod",
    false,
    {
      "name": "pmod",
      "body": "q",
      "arguments": 0
    }
  ],
  [
    "pmod-zero-arguments",
    "\\pmod",
    true,
    {
      "name": "pmod",
      "body": "q",
      "arguments": 0
    }
  ],
  [
    "pod-two-arguments",
    "\\pod{x}{y}",
    false,
    {
      "name": "pod",
      "body": "#1+#2",
      "arguments": 2
    }
  ],
  [
    "pod-two-arguments",
    "\\pod{x}{y}",
    true,
    {
      "name": "pod",
      "body": "#1+#2",
      "arguments": 2
    }
  ],
  [
    "two-argument-missing",
    "\\pod{x}",
    false,
    {
      "name": "pod",
      "body": "#1+#2",
      "arguments": 2
    }
  ],
  [
    "two-argument-missing",
    "\\pod{x}",
    true,
    {
      "name": "pod",
      "body": "#1+#2",
      "arguments": 2
    }
  ],
  [
    "recursive-limit",
    "\\pod{x}",
    false,
    {
      "name": "pod",
      "body": "\\pod{#1}",
      "arguments": 1
    }
  ],
  [
    "recursive-limit",
    "\\pod{x}",
    true,
    {
      "name": "pod",
      "body": "\\pod{#1}",
      "arguments": 1
    }
  ]
];
const explicit=n=>({kind:n.kind,text:n.text,attributes:n.attributes,children:n.children.map(explicit)});
const records=[];
for(const [label,tex,display,registration] of inputs){
 const name=label+(display?'-display':'-inline');
 const c=vm.createContext({console:{log(){},warn(){},error(){}}});
 for(const [f,b]of assets)vm.runInContext(b.toString(),c,{filename:f});c.request={tex,display,registration};
 const result=vm.runInContext(`(()=>{const r=request.registration;const {Macro}=MathJax._.input.tex.Symbol;const methods=MathJax._.input.tex.base.BaseMethods.default;html.inputJax[0].configuration.handlers.retrieve('ams-declare-ops').add(r.name,new Macro(r.name,methods.Macro,[r.body,r.arguments]));let tree, fullTree;const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.coreIndex||0},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});const project=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?Object.fromEntries(n.attributes.getExplicitNames().sort().map(k=>[k,n.attributes.getExplicit(k)])):{},properties:n.getAllProperties?n.getAllProperties():{},isAccent:n.kind==='mo'?n.isAccent:null,coreParentKind:n.kind==='mo'?n.coreParent().kind:null,coreParentParentKind:n.kind==='mo'&&n.coreParent().parent?n.coreParent().parent.kind:null,children:n.childNodes.map(project)});const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){tree=JSON.parse(JSON.stringify(project(math.root)));fullTree=full(math.root);return original.call(this,math,doc)};const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));return{tree,svg,fullTree};})()`,c);
 // Expected invalid-operator cases retain complete primary error SVG and AST.
 if(evidence){fs.writeFileSync(path.join(evidence,name+'.svg'),result.svg);fs.writeFileSync(path.join(evidence,name+'.json'),JSON.stringify(result.tree,null,2)+'\n');fs.writeFileSync(path.join(evidence,name+'.full.json'),JSON.stringify(result.fullTree,null,2)+'\n');}
 records.push({name,tex,display,registration,width:Math.ceil(Number(result.svg.match(/width="([\d.]+)ex"/)[1])*8),height:Math.ceil(Number(result.svg.match(/height="([\d.]+)ex"/)[1])*8),tree:explicit(result.tree),propertiesTree:result.tree,svgSHA256:hash(result.svg)});
}
fs.writeFileSync(path.join(evidence||__dirname,'pod_macros_mathjax_3_2_2.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' exact AST and complete SVG cases');
