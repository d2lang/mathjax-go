// SPDX-License-Identifier: Apache-2.0
// Unmodified pinned MathJax FBox/ColorBox/FColorBox references; no parser overlays.
const fs = require('node:fs'), vm = require('node:vm'), path = require('node:path'), crypto = require('node:crypto');
const base = process.argv[2], evidence = process.argv[3];
if (!base) throw Error('usage: node generate_framed_internal_text.cjs /path/to/pinned/d2latex [evidence-directory]');
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
    "name": "fbox-plain-inline",
    "tex": "\\fbox{ a~b }",
    "display": false,
    "purpose": "literal text, boundary whitespace and tilde source behavior"
  },
  {
    "name": "fbox-plain-display",
    "tex": "\\fbox{ a~b }",
    "display": true,
    "purpose": "literal text, boundary whitespace and tilde source behavior"
  },
  {
    "name": "fbox-empty-inline",
    "tex": "\\fbox{}",
    "display": false,
    "purpose": "zero content without added text child"
  },
  {
    "name": "fbox-empty-display",
    "tex": "\\fbox{}",
    "display": true,
    "purpose": "zero content without added text child"
  },
  {
    "name": "fbox-dollar-inline",
    "tex": "\\fbox{a $x$}",
    "display": false,
    "purpose": "confirmed embedded dollar math; framed-color route new witness"
  },
  {
    "name": "fbox-dollar-display",
    "tex": "\\fbox{a $x$}",
    "display": true,
    "purpose": "confirmed embedded dollar math; framed-color route new witness"
  },
  {
    "name": "fbox-paren-inline",
    "tex": "\\fbox{a \\(x+1\\) z}",
    "display": false,
    "purpose": "parenthesis delimiters"
  },
  {
    "name": "fbox-paren-display",
    "tex": "\\fbox{a \\(x+1\\) z}",
    "display": true,
    "purpose": "parenthesis delimiters"
  },
  {
    "name": "fbox-escaped-inline",
    "tex": "\\fbox{a \\$x\\$ \\{b\\} \\\\}",
    "display": false,
    "purpose": "escaped delimiters and literal backslash"
  },
  {
    "name": "fbox-escaped-display",
    "tex": "\\fbox{a \\$x\\$ \\{b\\} \\\\}",
    "display": true,
    "purpose": "escaped delimiters and literal backslash"
  },
  {
    "name": "fbox-multiple-inline",
    "tex": "\\fbox{a $x$ b $y$ c}",
    "display": false,
    "purpose": "multiple fragments and exact explicit-row topology"
  },
  {
    "name": "fbox-multiple-display",
    "tex": "\\fbox{a $x$ b $y$ c}",
    "display": true,
    "purpose": "multiple fragments and exact explicit-row topology"
  },
  {
    "name": "fbox-unterminated-dollar-inline",
    "tex": "\\fbox{a $x}",
    "display": false,
    "purpose": "required MathNotTerminated error"
  },
  {
    "name": "fbox-unterminated-dollar-display",
    "tex": "\\fbox{a $x}",
    "display": true,
    "purpose": "required MathNotTerminated error"
  },
  {
    "name": "colorbox-plain-inline",
    "tex": "\\colorbox{red}{ a~b }",
    "display": false,
    "purpose": "literal text, boundary whitespace and tilde source behavior"
  },
  {
    "name": "colorbox-plain-display",
    "tex": "\\colorbox{red}{ a~b }",
    "display": true,
    "purpose": "literal text, boundary whitespace and tilde source behavior"
  },
  {
    "name": "colorbox-empty-inline",
    "tex": "\\colorbox{red}{}",
    "display": false,
    "purpose": "zero content without added text child"
  },
  {
    "name": "colorbox-empty-display",
    "tex": "\\colorbox{red}{}",
    "display": true,
    "purpose": "zero content without added text child"
  },
  {
    "name": "colorbox-dollar-inline",
    "tex": "\\colorbox{red}{a $x$}",
    "display": false,
    "purpose": "confirmed embedded dollar math; framed-color route new witness"
  },
  {
    "name": "colorbox-dollar-display",
    "tex": "\\colorbox{red}{a $x$}",
    "display": true,
    "purpose": "confirmed embedded dollar math; framed-color route new witness"
  },
  {
    "name": "colorbox-paren-inline",
    "tex": "\\colorbox{red}{a \\(x+1\\) z}",
    "display": false,
    "purpose": "parenthesis delimiters"
  },
  {
    "name": "colorbox-paren-display",
    "tex": "\\colorbox{red}{a \\(x+1\\) z}",
    "display": true,
    "purpose": "parenthesis delimiters"
  },
  {
    "name": "colorbox-escaped-inline",
    "tex": "\\colorbox{red}{a \\$x\\$ \\{b\\} \\\\}",
    "display": false,
    "purpose": "escaped delimiters and literal backslash"
  },
  {
    "name": "colorbox-escaped-display",
    "tex": "\\colorbox{red}{a \\$x\\$ \\{b\\} \\\\}",
    "display": true,
    "purpose": "escaped delimiters and literal backslash"
  },
  {
    "name": "colorbox-multiple-inline",
    "tex": "\\colorbox{red}{a $x$ b $y$ c}",
    "display": false,
    "purpose": "multiple fragments and exact explicit-row topology"
  },
  {
    "name": "colorbox-multiple-display",
    "tex": "\\colorbox{red}{a $x$ b $y$ c}",
    "display": true,
    "purpose": "multiple fragments and exact explicit-row topology"
  },
  {
    "name": "colorbox-unterminated-dollar-inline",
    "tex": "\\colorbox{red}{a $x}",
    "display": false,
    "purpose": "required MathNotTerminated error"
  },
  {
    "name": "colorbox-unterminated-dollar-display",
    "tex": "\\colorbox{red}{a $x}",
    "display": true,
    "purpose": "required MathNotTerminated error"
  },
  {
    "name": "fcolorbox-plain-inline",
    "tex": "\\fcolorbox{blue}{red}{ a~b }",
    "display": false,
    "purpose": "literal text, boundary whitespace and tilde source behavior"
  },
  {
    "name": "fcolorbox-plain-display",
    "tex": "\\fcolorbox{blue}{red}{ a~b }",
    "display": true,
    "purpose": "literal text, boundary whitespace and tilde source behavior"
  },
  {
    "name": "fcolorbox-empty-inline",
    "tex": "\\fcolorbox{blue}{red}{}",
    "display": false,
    "purpose": "zero content without added text child"
  },
  {
    "name": "fcolorbox-empty-display",
    "tex": "\\fcolorbox{blue}{red}{}",
    "display": true,
    "purpose": "zero content without added text child"
  },
  {
    "name": "fcolorbox-dollar-inline",
    "tex": "\\fcolorbox{blue}{red}{a $x$}",
    "display": false,
    "purpose": "confirmed embedded dollar math; framed-color route new witness"
  },
  {
    "name": "fcolorbox-dollar-display",
    "tex": "\\fcolorbox{blue}{red}{a $x$}",
    "display": true,
    "purpose": "confirmed embedded dollar math; framed-color route new witness"
  },
  {
    "name": "fcolorbox-paren-inline",
    "tex": "\\fcolorbox{blue}{red}{a \\(x+1\\) z}",
    "display": false,
    "purpose": "parenthesis delimiters"
  },
  {
    "name": "fcolorbox-paren-display",
    "tex": "\\fcolorbox{blue}{red}{a \\(x+1\\) z}",
    "display": true,
    "purpose": "parenthesis delimiters"
  },
  {
    "name": "fcolorbox-escaped-inline",
    "tex": "\\fcolorbox{blue}{red}{a \\$x\\$ \\{b\\} \\\\}",
    "display": false,
    "purpose": "escaped delimiters and literal backslash"
  },
  {
    "name": "fcolorbox-escaped-display",
    "tex": "\\fcolorbox{blue}{red}{a \\$x\\$ \\{b\\} \\\\}",
    "display": true,
    "purpose": "escaped delimiters and literal backslash"
  },
  {
    "name": "fcolorbox-multiple-inline",
    "tex": "\\fcolorbox{blue}{red}{a $x$ b $y$ c}",
    "display": false,
    "purpose": "multiple fragments and exact explicit-row topology"
  },
  {
    "name": "fcolorbox-multiple-display",
    "tex": "\\fcolorbox{blue}{red}{a $x$ b $y$ c}",
    "display": true,
    "purpose": "multiple fragments and exact explicit-row topology"
  },
  {
    "name": "fcolorbox-unterminated-dollar-inline",
    "tex": "\\fcolorbox{blue}{red}{a $x}",
    "display": false,
    "purpose": "required MathNotTerminated error"
  },
  {
    "name": "fcolorbox-unterminated-dollar-display",
    "tex": "\\fcolorbox{blue}{red}{a $x}",
    "display": true,
    "purpose": "required MathNotTerminated error"
  },
  {
    "name": "font-inline",
    "tex": "{\\bf\\fbox{a $x$}}",
    "display": false,
    "purpose": "inherited literal font versus fresh lexical math environment"
  },
  {
    "name": "font-display",
    "tex": "{\\bf\\fbox{a $x$}}",
    "display": true,
    "purpose": "inherited literal font versus fresh lexical math environment"
  },
  {
    "name": "script-inline",
    "tex": "x_{\\colorbox{red}{a $x$}}",
    "display": false,
    "purpose": "caller level remains unspecified; no forced level-zero mstyle"
  },
  {
    "name": "script-display",
    "tex": "x_{\\colorbox{red}{a $x$}}",
    "display": true,
    "purpose": "caller level remains unspecified; no forced level-zero mstyle"
  },
  {
    "name": "outer-color-inline",
    "tex": "\\color{green}\\fcolorbox{blue}{yellow}{a $x$}",
    "display": false,
    "purpose": "outer color inheritance plus box/frame colors"
  },
  {
    "name": "outer-color-display",
    "tex": "\\color{green}\\fcolorbox{blue}{yellow}{a $x$}",
    "display": true,
    "purpose": "outer color inheritance plus box/frame colors"
  },
  {
    "name": "defined-colors-inline",
    "tex": "\\definecolor{bg}{RGB}{255,0,0}\\definecolor{fr}{RGB}{0,0,255}\\fcolorbox{fr}{bg}{a $x$}",
    "display": false,
    "purpose": "existing shared named colors"
  },
  {
    "name": "defined-colors-display",
    "tex": "\\definecolor{bg}{RGB}{255,0,0}\\definecolor{fr}{RGB}{0,0,255}\\fcolorbox{fr}{bg}{a $x$}",
    "display": true,
    "purpose": "existing shared named colors"
  },
  {
    "name": "inner-defines-background-inline",
    "tex": "\\colorbox{late}{$\\definecolor{late}{RGB}{255,0,0}x$}",
    "display": false,
    "purpose": "internalMath before background lookup, configuration side effect"
  },
  {
    "name": "inner-defines-background-display",
    "tex": "\\colorbox{late}{$\\definecolor{late}{RGB}{255,0,0}x$}",
    "display": true,
    "purpose": "internalMath before background lookup, configuration side effect"
  },
  {
    "name": "inner-defines-frame-inline",
    "tex": "\\fcolorbox{late}{yellow}{$\\definecolor{late}{RGB}{0,0,255}x$}",
    "display": false,
    "purpose": "internalMath before frame lookup"
  },
  {
    "name": "inner-defines-frame-display",
    "tex": "\\fcolorbox{late}{yellow}{$\\definecolor{late}{RGB}{0,0,255}x$}",
    "display": true,
    "purpose": "internalMath before frame lookup"
  },
  {
    "name": "macro-inside-inline",
    "tex": "\\def\\foo{x+1}\\colorbox{red}{a $\\foo$}",
    "display": false,
    "purpose": "shared registered macro state"
  },
  {
    "name": "macro-inside-display",
    "tex": "\\def\\foo{x+1}\\colorbox{red}{a $\\foo$}",
    "display": true,
    "purpose": "shared registered macro state"
  },
  {
    "name": "nested-inline",
    "tex": "\\fbox{a $\\colorbox{red}{b $x$}$}",
    "display": false,
    "purpose": "nested text/mathematical scopes"
  },
  {
    "name": "nested-display",
    "tex": "\\fbox{a $\\colorbox{red}{b $x$}$}",
    "display": true,
    "purpose": "nested text/mathematical scopes"
  },
  {
    "name": "unterminated-paren-inline",
    "tex": "\\fcolorbox{blue}{red}{a \\(x}",
    "display": false,
    "purpose": "required incomplete parenthesis error"
  },
  {
    "name": "unterminated-paren-display",
    "tex": "\\fcolorbox{blue}{red}{a \\(x}",
    "display": true,
    "purpose": "required incomplete parenthesis error"
  },
  {
    "name": "inner-error-inline",
    "tex": "\\colorbox{red}{a $\\notADefinedCommand$}",
    "display": false,
    "purpose": "actual inner TeX error propagation"
  },
  {
    "name": "inner-error-display",
    "tex": "\\colorbox{red}{a $\\notADefinedCommand$}",
    "display": true,
    "purpose": "actual inner TeX error propagation"
  },
  {
    "name": "boxed-control-inline",
    "tex": "\\boxed{x+y}",
    "display": false,
    "purpose": "existing separate math-only boxed route unchanged"
  },
  {
    "name": "boxed-control-display",
    "tex": "\\boxed{x+y}",
    "display": true,
    "purpose": "existing separate math-only boxed route unchanged"
  }
];
const cases = [];
for (const input of inputs) {
  const ctx=vm.createContext({console});for(const[name,b]of assets)vm.runInContext(b,ctx,{filename:name});ctx.request=input;
  const result=vm.runInContext(`(() => {
    const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?n.attributes.getAllAttributes():{},properties:n.getAllProperties(),children:n.childNodes.map(full)});
    const pairs=o=>Object.entries(o).map(([name,value])=>({name,value}));
    const raw=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.defaults),global:pairs(n.attributes.global)}:{},properties:pairs(n.getAllProperties()),children:n.childNodes.map(raw)});
    let tree,rawTree;const original=html.outputJax.typeset;
    html.outputJax.typeset=function(math,doc){tree=JSON.parse(JSON.stringify(full(math.root)));rawTree=JSON.parse(JSON.stringify(raw(math.root)));return Reflect.apply(original,this,[math,doc]);};
    const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));
    return{tree,rawTree,svg};
  })()`,ctx);
  cases.push({...input,svgSHA256:hash(result.svg),tree:result.tree});
  if(evidence){fs.writeFileSync(path.join(evidence,input.name+'.svg'),result.svg);fs.writeFileSync(path.join(evidence,input.name+'.json'),JSON.stringify(result.tree,null,2)+'\n');fs.writeFileSync(path.join(evidence,input.name+'.full.json'),JSON.stringify(result.rawTree,null,2)+'\n');}
}
process.stdout.write(JSON.stringify({mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',cases},null,2)+'\n');
