// SPDX-License-Identifier: Apache-2.0
// Observe original GetCS only; never invokes Go.
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),crypto=require('node:crypto');
const assets=process.argv[2];
if (!assets) throw Error('expected pinned D2 asset directory');
const hashes={'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01','mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869','setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'};
const c=vm.createContext({console});c.globalThis=c;
for(const[n,h]of Object.entries(hashes)){const b=fs.readFileSync(path.join(assets,n));if(crypto.createHash('sha256').update(b).digest('hex')!==h)throw Error(n);new vm.Script(b.toString(),{filename:n}).runInContext(c);}
const chars=[['LF','\n'],['CR','\r'],['LS','\u2028'],['PS','\u2029'],['tab','\t'],['vt','\v'],['ff','\f'],['NEL','\x85'],['NBSP','\xa0'],['BOM','\ufeff'],['space',' '],['astral','🙂'],['nonletter','@']];
c.requests=[];for(const[label,ch]of chars)for(const tail of['','x','{x}','\n',' '])c.requests.push({source:ch+tail,label});
for(const source of['','alpha x','alpha  x','alpha\nx','alpha\u2028x','Alpha{x}','a1','𝛂x','éx'])c.requests.push({source,label:'name-or-EOF'});
const d=vm.runInContext(`(()=>{const p=MathJax._.input.tex.TexParser.default.prototype;return {method:p.GetCS.toString(),cases:requests.map(r=>{const x=Object.create(p);x.string=r.source;x.i=0;const value=p.GetCS.call(x);return {...r,value,cursorCodeUnits:x.i,consumedBytes:unescape(encodeURIComponent(x.string.slice(0,x.i))).length,remaining:x.string.slice(x.i)}})}})()`,c);
fs.writeFileSync(path.join(__dirname,'getcs_line_endings_mathjax_3_2_2.json'),JSON.stringify({mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,...d},null,2)+'\n');console.log(d.cases.length);
