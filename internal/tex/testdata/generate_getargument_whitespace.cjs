// SPDX-License-Identifier: Apache-2.0
// Observe the pinned original GetArgument/GetNext methods; this never runs Go.
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),crypto=require('node:crypto');
const assets=process.argv[2],c=vm.createContext({console});
const hashes={'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01','mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869','setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'};
for(const [n,h]of Object.entries(hashes)){const b=fs.readFileSync(path.join(assets,n));if(crypto.createHash('sha256').update(b).digest('hex')!==h)throw Error(n);vm.runInContext(b.toString(),c,{filename:n});}
const points=[9,10,11,12,13,32,0x85,0xa0,0x1680,...Array.from({length:11},(_,i)=>0x2000+i),0x2028,0x2029,0x202f,0x205f,0x3000,0xfeff,0x200b,0x180e,0x2060];
c.requests=[];for(const p of points)for(const tail of ['{ab}z','🙂z','\\alpha z','}z',''])for(const noneOK of [false,true])c.requests.push({source:String.fromCodePoint(p)+tail,noneOK,codePoint:p});
for(const source of ['\ufeff\x85{a}','\x85\ufeff{a}','%comment\n{a}','{a\ufeffb}','{a\x85b}','\\','\\\u2028','\\\r','\\\n'])for(const noneOK of [false,true])c.requests.push({source,noneOK});
const d=vm.runInContext(`(()=>{const p=MathJax._.input.tex.TexParser.default.prototype;return {methods:{GetArgument:p.GetArgument.toString(),GetNext:p.GetNext.toString(),nextIsSpace:p.nextIsSpace.toString()},cases:requests.map(r=>{const x=Object.create(p);x.string=r.source;x.i=0;x.currentCS='\\\\sample';let value,error;try{value=p.GetArgument.call(x,'\\\\different',r.noneOK)}catch(e){error={id:e.id,message:e.message}}return {...r,value:value===undefined?null:value,error:error||null,cursorCodeUnits:x.i,cursorBytes:unescape(encodeURIComponent(x.string.slice(0,x.i))).length,remaining:x.string.slice(x.i)}})}})()`,c);
const escapedLines = new Set(['\\\n', '\\\r', '\\\u2028']);
d.rawControls = d.cases.filter(r => escapedLines.has(r.source));
d.cases = d.cases.filter(r => !escapedLines.has(r.source));
fs.writeFileSync(path.join(__dirname,'getargument_whitespace_mathjax_3_2_2.json'),JSON.stringify({mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,...d},null,2)+'\n');console.log(d.cases.length);
