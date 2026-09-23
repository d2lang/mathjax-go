// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Identifier automatic operator policy: public normal/font/script/error controls.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_identifier_operator.cjs /path/to/pinned/d2latex [evidence-directory]"
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
    "token-row",
    "x\\mmlToken{mi}{mod}y",
    false
  ],
  [
    "token-row",
    "x\\mmlToken{mi}{mod}y",
    true
  ],
  [
    "token-alone",
    "\\mmlToken{mi}{mod}",
    false
  ],
  [
    "token-alone",
    "\\mmlToken{mi}{mod}",
    true
  ],
  [
    "token-bin",
    "\\mmlToken{mi}{mod}+x",
    false
  ],
  [
    "token-bin",
    "\\mmlToken{mi}{mod}+x",
    true
  ],
  [
    "token-rel",
    "\\mmlToken{mi}{mod}=x",
    false
  ],
  [
    "token-rel",
    "\\mmlToken{mi}{mod}=x",
    true
  ],
  [
    "token-uppercase",
    "x\\mmlToken{mi}{Ab12}y",
    false
  ],
  [
    "token-uppercase",
    "x\\mmlToken{mi}{Ab12}y",
    true
  ],
  [
    "token-single",
    "x\\mmlToken{mi}{a}y",
    false
  ],
  [
    "token-single",
    "x\\mmlToken{mi}{a}y",
    true
  ],
  [
    "token-digit-first",
    "x\\mmlToken{mi}{1a}y",
    false
  ],
  [
    "token-digit-first",
    "x\\mmlToken{mi}{1a}y",
    true
  ],
  [
    "token-punctuation",
    "x\\mmlToken{mi}{a.b}y",
    false
  ],
  [
    "token-punctuation",
    "x\\mmlToken{mi}{a.b}y",
    true
  ],
  [
    "token-greek",
    "x\\mmlToken{mi}{αβ}y",
    false
  ],
  [
    "token-greek",
    "x\\mmlToken{mi}{αβ}y",
    true
  ],
  [
    "token-kelvin",
    "x\\mmlToken{mi}{Ka}y",
    false
  ],
  [
    "token-kelvin",
    "x\\mmlToken{mi}{Ka}y",
    true
  ],
  [
    "token-long-s",
    "x\\mmlToken{mi}{ſa}y",
    false
  ],
  [
    "token-long-s",
    "x\\mmlToken{mi}{ſa}y",
    true
  ],
  [
    "token-bold",
    "x\\mmlToken{mi}[mathvariant=\"bold\"]{mod}y",
    false
  ],
  [
    "token-bold",
    "x\\mmlToken{mi}[mathvariant=\"bold\"]{mod}y",
    true
  ],
  [
    "token-italic",
    "x\\mmlToken{mi}[mathvariant=\"italic\"]{mod}y",
    false
  ],
  [
    "token-italic",
    "x\\mmlToken{mi}[mathvariant=\"italic\"]{mod}y",
    true
  ],
  [
    "token-normal",
    "x\\mmlToken{mi}[mathvariant=\"normal\"]{mod}y",
    false
  ],
  [
    "token-normal",
    "x\\mmlToken{mi}[mathvariant=\"normal\"]{mod}y",
    true
  ],
  [
    "bold-scope",
    "\\mathbf{x\\mmlToken{mi}{mod}y}",
    false
  ],
  [
    "bold-scope",
    "\\mathbf{x\\mmlToken{mi}{mod}y}",
    true
  ],
  [
    "normal-kept",
    "\\mathit{x\\mmlToken{mi}[mathvariant=\"normal\"]{mod}y}",
    false
  ],
  [
    "normal-kept",
    "\\mathit{x\\mmlToken{mi}[mathvariant=\"normal\"]{mod}y}",
    true
  ],
  [
    "ordinary-roman",
    "x\\mathrm{mod}y",
    false
  ],
  [
    "ordinary-roman",
    "x\\mathrm{mod}y",
    true
  ],
  [
    "ordinary-letters",
    "x m o d y",
    false
  ],
  [
    "ordinary-letters",
    "x m o d y",
    true
  ],
  [
    "named-operator",
    "x\\operatorname{mod}y",
    false
  ],
  [
    "named-operator",
    "x\\operatorname{mod}y",
    true
  ],
  [
    "fixed-sin",
    "x\\sin y",
    false
  ],
  [
    "fixed-sin",
    "x\\sin y",
    true
  ],
  [
    "grouped",
    "x{\\mmlToken{mi}{mod}}y",
    false
  ],
  [
    "grouped",
    "x{\\mmlToken{mi}{mod}}y",
    true
  ],
  [
    "script",
    "x^{a\\mmlToken{mi}{mod}b}",
    false
  ],
  [
    "script",
    "x^{a\\mmlToken{mi}{mod}b}",
    true
  ],
  [
    "scripted-token",
    "x\\mmlToken{mi}{mod}_i^n y",
    false
  ],
  [
    "scripted-token",
    "x\\mmlToken{mi}{mod}_i^n y",
    true
  ],
  [
    "fraction",
    "\\frac{x\\mmlToken{mi}{mod}y}{z}",
    false
  ],
  [
    "fraction",
    "\\frac{x\\mmlToken{mi}{mod}y}{z}",
    true
  ],
  [
    "small-style",
    "\\scriptstyle x\\mmlToken{mi}{mod}y",
    false
  ],
  [
    "small-style",
    "\\scriptstyle x\\mmlToken{mi}{mod}y",
    true
  ],
  [
    "nested",
    "\\sqrt{x\\mmlToken{mi}{mod}y}",
    false
  ],
  [
    "nested",
    "\\sqrt{x\\mmlToken{mi}{mod}y}",
    true
  ],
  [
    "color",
    "x\\mmlToken{mi}[mathcolor=\"red\"]{mod}y",
    false
  ],
  [
    "color",
    "x\\mmlToken{mi}[mathcolor=\"red\"]{mod}y",
    true
  ],
  [
    "missing-content",
    "\\mmlToken{mi}",
    false
  ],
  [
    "missing-content",
    "\\mmlToken{mi}",
    true
  ],
  [
    "bad-attribute",
    "\\mmlToken{mi}[bad=\"x\"]{mod}",
    false
  ],
  [
    "bad-attribute",
    "\\mmlToken{mi}[bad=\"x\"]{mod}",
    true
  ],
  [
    "duplicate-script",
    "\\mmlToken{mi}{mod}^a^b",
    false
  ],
  [
    "duplicate-script",
    "\\mmlToken{mi}{mod}^a^b",
    true
  ]
];
const explicit=n=>({kind:n.kind,text:n.text,attributes:n.attributes,children:n.children.map(explicit)});
const records=[];
for(const [label,tex,display] of inputs){
 const name=label+(display?'-display':'-inline');
 const c=vm.createContext({console:{log(){},warn(){},error(){}}});
 for(const [f,b]of assets)vm.runInContext(b.toString(),c,{filename:f});c.request={tex,display};
 const result=vm.runInContext(`(()=>{let tree, fullTree;const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),flags:{token:n.isToken,embellished:n.isEmbellished,spacelike:n.isSpacelike,linebreakContainer:n.linebreakContainer,hasNewline:n.hasNewline,inferred:n.isInferred,notParent:n.notParent,arity:n.arity===Infinity?-2:n.arity,coreIndex:n.coreIndex||0},texClass:Number.isFinite(n.texClass)?n.texClass:-1,prevClass:Number.isFinite(n.prevClass)?n.prevClass:-1,prevLevel:Number.isFinite(n.prevLevel)?n.prevLevel:0,children:n.childNodes.map(full)});const project=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?Object.fromEntries(n.attributes.getExplicitNames().sort().map(k=>[k,n.attributes.getExplicit(k)])):{},properties:n.getAllProperties?n.getAllProperties():{},isAccent:n.kind==='mo'?n.isAccent:null,coreParentKind:n.kind==='mo'?n.coreParent().kind:null,coreParentParentKind:n.kind==='mo'&&n.coreParent().parent?n.coreParent().parent.kind:null,children:n.childNodes.map(project)});const identifiers=n=>{const rows=[];const walk=(q,p)=>{if(q.kind==='mi')rows.push({path:p,text:q.getText(),texClass:q.texClass,properties:q.getAllProperties(),variant:q.attributes.get('mathvariant')});q.childNodes.forEach((x,i)=>walk(x,p+'/'+i))};walk(n,'root');return rows};let identifierState;const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){tree=project(math.root);fullTree=full(math.root);const out=original.call(this,math,doc);identifierState=identifiers(math.root);return out};const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));return{tree,svg,fullTree,identifierState};})()`,c);
 // Expected invalid-operator cases retain complete primary error SVG and AST.
 if(evidence){fs.writeFileSync(path.join(evidence,name+'.svg'),result.svg);fs.writeFileSync(path.join(evidence,name+'.json'),JSON.stringify(result.tree,null,2)+'\n');fs.writeFileSync(path.join(evidence,name+'.full.json'),JSON.stringify(result.fullTree,null,2)+'\n');}
 records.push({name,tex,display,width:Math.ceil(Number(result.svg.match(/width="([\d.]+)ex"/)[1])*8),height:Math.ceil(Number(result.svg.match(/height="([\d.]+)ex"/)[1])*8),identifiers:result.identifierState,svgSHA256:hash(result.svg)});
}
fs.writeFileSync(path.join(__dirname,'identifier_operator_mathjax_3_2_2.json'),JSON.stringify({oracle:'Unmodified pinned D2 MathJax3.2.2; fresh VM per case/mode; AST observed before unchanged SVG renderer',mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');console.log(records.length+' exact AST and complete SVG cases');
