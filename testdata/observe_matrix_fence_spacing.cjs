// SPDX-License-Identifier: Apache-2.0
const fs=require('node:fs'),vm=require('node:vm'),crypto=require('node:crypto');
const path=require('node:path');
const assets=process.argv[2];if(!assets)throw Error('usage: node --jitless testdata/observe_matrix_fence_spacing.cjs PINNED_ASSETS');
const hashes={'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01','mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869','setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'};
const scripts=Object.entries(hashes).map(([name,hash])=>{const b=fs.readFileSync(assets+'/'+name);if(crypto.createHash('sha256').update(b).digest('hex')!==hash)throw Error('hash');return new vm.Script(b.toString(),{filename:name});});
function fresh(){const c=vm.createContext({console});c.globalThis=c;for(const s of scripts)s.runInContext(c);return c;}
const c=fresh(),tex=c.MathJax._.input.tex,pu=tex.ParseUtil.default||tex.ParseUtil,items=tex.base.BaseItems;

const info={mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',freshRuntimePerObservation:true,hashes,unsupportedOriginalCommand:{name:'boldsymbol',registered:!!c.html.inputJax[0].parseOptions.handlers.get('macro').lookup('boldsymbol')},source:{Matrix:tex.base.BaseMethods.default.Matrix.toString(),fenced:pu.fenced.toString(),ArrayToMml:items.ArrayItem.prototype.toMml.toString()},commands:{},observations:[]};
for(const name of ['matrix','array','pmatrix','cases','eqalign','displaylines','eqalignno','leqalignno']){const m=c.html.inputJax[0].parseOptions.handlers.get('macro').lookup(name);info.commands[name]={args:m.args,BaseMatrix:m.func===tex.base.BaseMethods.default.Matrix};}
for(const [name,t]of Object.entries({pmatrix:String.raw`\text{Matrix: }\pmatrix{a&b\cr c&d}`,cases:String.raw`x\cases{x&if x>0\cr 0&otherwise}y`,environment:String.raw`\text{Matrix: }\begin{pmatrix}a&b\\c&d\end{pmatrix}`})){
 const r=fresh(),p=r.MathJax._.input.tex.ParseUtil.default||r.MathJax._.input.tex.ParseUtil;const old=p.fenced,rows=[];
 p.fenced=function(...args){const n=old.apply(this,args);rows.push(n);return n;};
 const output=r.html.convert(t,{em:16,ex:8,display:true});
 info.observations.push({name,tex:t,svg:r.adaptor.innerHTML(output),rows:rows.map(n=>({kind:n.kind,properties:n.getAllProperties(),texClass:n.texClass,prevClass:n.prevClass,children:n.childNodes.map(ch=>({kind:ch.kind,texClass:ch.texClass,prevClass:ch.prevClass}))}))});
}
fs.writeFileSync(path.join(__dirname,'matrix_fence_spacing_observations.json'),JSON.stringify(info,null,2)+'\n');
