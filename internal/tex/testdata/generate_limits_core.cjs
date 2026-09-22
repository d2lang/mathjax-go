// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
// Private Limits node/family contract; no source overlays.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_limits_core.cjs /path/to/pinned/d2latex [evidence-directory]"
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

const c=vm.createContext({console});for(const [f,b] of assets)vm.runInContext(b.toString(),c,{filename:f});
const result=vm.runInContext(`(()=>{const f=html.inputJax[0].mmlFactory;const kinds=['msub','msup','msubsup','munder','mover','munderover'];const rows=kinds.map(k=>{const x=f.create(k);return {kind:k,isKindFunction:x.isKind.toString(),isMsubsup:x.isKind('msubsup'),isMunderover:x.isKind('munderover')}});const selections=[-2,0,1,2,3,99,1.5,null,true,false,'',' 2 ','0x2','0X2','0b10','0B10','0o2','0O2','1_0','Inf','Infinity','+Infinity','-Infinity','\\uFEFF2','\\u00852','+2','1e400','-1e400','1e-400','0x','0b2','0o8','0x1p1','+0x2','-0b10','NaN','infinity','.2','2.','02','2e0','1e','1e+','2a','2 0','1_000','0x1_0','0b1_0','0o1_0', '0x'+ 'f'.repeat(300), '0b'+'1'.repeat(2000), '0o'+'7'.repeat(700), ...[9,10,11,12,13,32,160,5760,8192,8193,8194,8195,8196,8197,8198,8199,8200,8201,8202,8232,8233,8239,8287,12288,65279,133,6158,8203].map(c=>String.fromCodePoint(c)+'2'+String.fromCodePoint(c))];const selected=selections.map(s=>{const a=f.create('mi',{},[f.create('text').setText('x')]),b=f.create('mo',{},[f.create('text').setText('∑')]);const x=f.create('maction',{selection:s},[a,b]);try{return{selection:s,corePath:x.coreMO()===a?0:x.coreMO()===b?1:null,coreKind:x.coreMO()?.kind,selectedSource:Object.getOwnPropertyDescriptor(Object.getPrototypeOf(x),'selected')?.get.toString(),coreMOSource:x.coreMO.toString()}}catch(e){return{selection:s,error:String(e)}}});return{rows,selected};})()`,c);fs.writeFileSync(path.join(evidence,'limits_core_mathjax_3_2_2.json'),JSON.stringify({assetsSHA256:hashes,...result},null,2)+'\n');
