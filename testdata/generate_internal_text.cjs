// SPDX-License-Identifier: Apache-2.0
// Unmodified pinned MathJax HBox and shared-caller controls; no parser overlays.
const fs = require('node:fs'), vm = require('node:vm'), path = require('node:path'), crypto = require('node:crypto');
const base = process.argv[2], evidence = process.argv[3];
if (!base) throw Error('usage: node generate_internal_text.cjs /path/to/pinned/d2latex [evidence-directory]');
if (process.env.NODE_OPTIONS) throw Error('Expected normal Node');
if (evidence) fs.mkdirSync(evidence, {recursive:true});
const hash = x => crypto.createHash('sha256').update(x).digest('hex');
const hashes = {
  "polyfills.js": "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01",
  "mathjax.js": "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869",
  "setup.js": "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881"
};
const assets = Object.entries(hashes).map(([name,expected]) => {const b=fs.readFileSync(path.join(base,name));if(hash(b)!==expected)throw Error('Unpinned '+name);return[name,b.toString()];});
const inputs = [
  {
    "name": "dollar-inline",
    "tex": "\\text{$x$}",
    "display": false
  },
  {
    "name": "dollar-display",
    "tex": "\\text{$x$}",
    "display": true
  },
  {
    "name": "dollar-error-inline",
    "tex": "\\text{a $x\\relax$ b}",
    "display": false
  },
  {
    "name": "dollar-error-display",
    "tex": "\\text{a $x\\relax$ b}",
    "display": true
  },
  {
    "name": "plain-inline",
    "tex": "\\text{plain text}",
    "display": false
  },
  {
    "name": "plain-display",
    "tex": "\\text{plain text}",
    "display": true
  },
  {
    "name": "empty-inline",
    "tex": "\\text{}",
    "display": false
  },
  {
    "name": "empty-display",
    "tex": "\\text{}",
    "display": true
  },
  {
    "name": "mixed-inline",
    "tex": "\\text{a $x_i^n$ b $y$ c}",
    "display": false
  },
  {
    "name": "mixed-display",
    "tex": "\\text{a $x_i^n$ b $y$ c}",
    "display": true
  },
  {
    "name": "parenthesis-inline",
    "tex": "\\text{a \\(x+y\\) b}",
    "display": false
  },
  {
    "name": "parenthesis-display",
    "tex": "\\text{a \\(x+y\\) b}",
    "display": true
  },
  {
    "name": "nested-braces-inline",
    "tex": "\\text{a ${x_{i}}$ b}",
    "display": false
  },
  {
    "name": "nested-braces-display",
    "tex": "\\text{a ${x_{i}}$ b}",
    "display": true
  },
  {
    "name": "nested-text-inline",
    "tex": "\\text{a $x+\\text{b $y$ c}$ d}",
    "display": false
  },
  {
    "name": "nested-text-display",
    "tex": "\\text{a $x+\\text{b $y$ c}$ d}",
    "display": true
  },
  {
    "name": "escaped-inline",
    "tex": "\\text{\\$x\\$ \\{a\\} \\\\ b}",
    "display": false
  },
  {
    "name": "escaped-display",
    "tex": "\\text{\\$x\\$ \\{a\\} \\\\ b}",
    "display": true
  },
  {
    "name": "literal-command-inline",
    "tex": "\\text{\\relax \\alpha}",
    "display": false
  },
  {
    "name": "literal-command-display",
    "tex": "\\text{\\relax \\alpha}",
    "display": true
  },
  {
    "name": "tilde-inline",
    "tex": "\\text{a~b}",
    "display": false
  },
  {
    "name": "tilde-display",
    "tex": "\\text{a~b}",
    "display": true
  },
  {
    "name": "edge-spaces-inline",
    "tex": "\\text{  a  b  }",
    "display": false
  },
  {
    "name": "edge-spaces-display",
    "tex": "\\text{  a  b  }",
    "display": true
  },
  {
    "name": "all-spaces-inline",
    "tex": "\\text{   }",
    "display": false
  },
  {
    "name": "all-spaces-display",
    "tex": "\\text{   }",
    "display": true
  },
  {
    "name": "bom-edges-inline",
    "tex": "\\text{﻿a﻿}",
    "display": false
  },
  {
    "name": "bom-edges-display",
    "tex": "\\text{﻿a﻿}",
    "display": true
  },
  {
    "name": "nel-edges-inline",
    "tex": "\\text{a}",
    "display": false
  },
  {
    "name": "nel-edges-display",
    "tex": "\\text{a}",
    "display": true
  },
  {
    "name": "supplementary-inline",
    "tex": "\\text{a𝛼 $x$ 𝛽}",
    "display": false
  },
  {
    "name": "supplementary-display",
    "tex": "\\text{a𝛼 $x$ 𝛽}",
    "display": true
  },
  {
    "name": "unclosed-dollar-inline",
    "tex": "\\text{a $x}",
    "display": false
  },
  {
    "name": "unclosed-dollar-display",
    "tex": "\\text{a $x}",
    "display": true
  },
  {
    "name": "unclosed-parenthesis-inline",
    "tex": "\\text{a \\(x}",
    "display": false
  },
  {
    "name": "unclosed-parenthesis-display",
    "tex": "\\text{a \\(x}",
    "display": true
  },
  {
    "name": "textbf-inline",
    "tex": "\\textbf{a $x+1$ b}",
    "display": false
  },
  {
    "name": "textbf-display",
    "tex": "\\textbf{a $x+1$ b}",
    "display": true
  },
  {
    "name": "textit-inline",
    "tex": "\\textit{a $x$ b}",
    "display": false
  },
  {
    "name": "textit-display",
    "tex": "\\textit{a $x$ b}",
    "display": true
  },
  {
    "name": "textrm-inline",
    "tex": "\\mathbf{\\textrm{a $x$}}",
    "display": false
  },
  {
    "name": "textrm-display",
    "tex": "\\mathbf{\\textrm{a $x$}}",
    "display": true
  },
  {
    "name": "textnormal-inline",
    "tex": "\\mathbf{\\textnormal{a $x$}}",
    "display": false
  },
  {
    "name": "textnormal-display",
    "tex": "\\mathbf{\\textnormal{a $x$}}",
    "display": true
  },
  {
    "name": "textsf-inline",
    "tex": "\\textsf{a $x$}",
    "display": false
  },
  {
    "name": "textsf-display",
    "tex": "\\textsf{a $x$}",
    "display": true
  },
  {
    "name": "texttt-inline",
    "tex": "\\texttt{a $x$}",
    "display": false
  },
  {
    "name": "texttt-display",
    "tex": "\\texttt{a $x$}",
    "display": true
  },
  {
    "name": "outer-font-inline",
    "tex": "\\mathbf{\\text{a $x+\\mathit{yz}$} z}",
    "display": false
  },
  {
    "name": "outer-font-display",
    "tex": "\\mathbf{\\text{a $x+\\mathit{yz}$} z}",
    "display": true
  },
  {
    "name": "font-declaration-inline",
    "tex": "{\\bf\\text{a $x$} y}",
    "display": false
  },
  {
    "name": "font-declaration-display",
    "tex": "{\\bf\\text{a $x$} y}",
    "display": true
  },
  {
    "name": "vector-inline",
    "tex": "\\vb{\\text{a $x+\\vb{y}$} z}",
    "display": false
  },
  {
    "name": "vector-display",
    "tex": "\\vb{\\text{a $x+\\vb{y}$} z}",
    "display": true
  },
  {
    "name": "vector-font-inline",
    "tex": "\\mathrm{\\vb{\\text{a $x$} y}}",
    "display": false
  },
  {
    "name": "vector-font-display",
    "tex": "\\mathrm{\\vb{\\text{a $x$} y}}",
    "display": true
  },
  {
    "name": "hbox-inline",
    "tex": "\\hbox{a $\\sum_i^n$ b}",
    "display": false
  },
  {
    "name": "hbox-display",
    "tex": "\\hbox{a $\\sum_i^n$ b}",
    "display": true
  },
  {
    "name": "mbox-empty-inline",
    "tex": "\\mbox{}",
    "display": false
  },
  {
    "name": "mbox-empty-display",
    "tex": "\\mbox{}",
    "display": true
  },
  {
    "name": "script-context-inline",
    "tex": "x_{\\text{a $\\sum_i^n$}}",
    "display": false
  },
  {
    "name": "script-context-display",
    "tex": "x_{\\text{a $\\sum_i^n$}}",
    "display": true
  },
  {
    "name": "textstyle-inline",
    "tex": "\\displaystyle\\text{a $\\sum_i^n$}",
    "display": false
  },
  {
    "name": "textstyle-display",
    "tex": "\\displaystyle\\text{a $\\sum_i^n$}",
    "display": true
  },
  {
    "name": "primes-inline",
    "tex": "\\text{a $x\\prime+x\\limits$ b}",
    "display": false
  },
  {
    "name": "primes-display",
    "tex": "\\text{a $x\\prime+x\\limits$ b}",
    "display": true
  },
  {
    "name": "ref-inline",
    "tex": "\\text{a \\ref{missing} b}",
    "display": false
  },
  {
    "name": "ref-display",
    "tex": "\\text{a \\ref{missing} b}",
    "display": true
  },
  {
    "name": "eqref-inline",
    "tex": "\\label{a}\\text{b \\eqref {a} c}",
    "display": false
  },
  {
    "name": "eqref-display",
    "tex": "\\label{a}\\text{b \\eqref {a} c}",
    "display": true
  },
  {
    "name": "fbox-inline",
    "tex": "\\fbox{a $x$}",
    "display": false
  },
  {
    "name": "fbox-display",
    "tex": "\\fbox{a $x$}",
    "display": true
  },
  {
    "name": "colorbox-inline",
    "tex": "\\colorbox{red}{a $x$}",
    "display": false
  },
  {
    "name": "colorbox-display",
    "tex": "\\colorbox{red}{a $x$}",
    "display": true
  },
  {
    "name": "fcolorbox-inline",
    "tex": "\\fcolorbox{blue}{red}{x}",
    "display": false
  },
  {
    "name": "fcolorbox-display",
    "tex": "\\fcolorbox{blue}{red}{x}",
    "display": true
  },
  {
    "name": "boxed-inline",
    "tex": "\\boxed{x+y}",
    "display": false
  },
  {
    "name": "boxed-display",
    "tex": "\\boxed{x+y}",
    "display": true
  },
  {
    "name": "ams-tag-inline",
    "tex": "x\\tag{a~b}",
    "display": false
  },
  {
    "name": "ams-tag-display",
    "tex": "x\\tag{a~b}",
    "display": true
  },
  {
    "name": "ams-ref-inline",
    "tex": "\\label{a}\\ref{a}",
    "display": false
  },
  {
    "name": "ams-ref-display",
    "tex": "\\label{a}\\ref{a}",
    "display": true
  },
  {
    "name": "textsl-inline",
    "tex": "\\textsl{a $x$}",
    "display": false
  },
  {
    "name": "textsl-display",
    "tex": "\\textsl{a $x$}",
    "display": true
  },
  {
    "name": "textup-inline",
    "tex": "\\textup{a $x$}",
    "display": false
  },
  {
    "name": "textup-display",
    "tex": "\\textup{a $x$}",
    "display": true
  },
  {
    "name": "mathmbox-inline",
    "tex": "\\mathmbox{x_i}",
    "display": false
  },
  {
    "name": "mathmbox-display",
    "tex": "\\mathmbox{x_i}",
    "display": true
  },
  {
    "name": "ordinary-math-inline",
    "tex": "\\mathbf{x}+\\vb{y}+x_i^n",
    "display": false
  },
  {
    "name": "ordinary-math-display",
    "tex": "\\mathbf{x}+\\vb{y}+x_i^n",
    "display": true
  }
];
const cases = [];
for (const input of inputs) {
  const ctx=vm.createContext({console});for(const[name,b]of assets)vm.runInContext(b,ctx,{filename:name});ctx.request=input;
  const result=vm.runInContext(`(() => {
    const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?n.attributes.getAllAttributes():{},properties:n.getAllProperties(),children:n.childNodes.map(full)});
    let tree;const original=html.outputJax.typeset;
    html.outputJax.typeset=function(math,doc){tree=full(math.root);return Reflect.apply(original,this,[math,doc]);};
    const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));
    return{tree,svg};
  })()`,ctx);
  cases.push({...input,svgSHA256:hash(result.svg),tree:result.tree});
  if(evidence){fs.writeFileSync(path.join(evidence,input.name+'.svg'),result.svg);fs.writeFileSync(path.join(evidence,input.name+'.json'),JSON.stringify(result.tree,null,2)+'\n');}
}
process.stdout.write(JSON.stringify({mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',cases},null,2)+'\n');
