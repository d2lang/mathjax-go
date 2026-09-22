// SPDX-License-Identifier: Apache-2.0
// Calls the registered Prime handler and actual TexParser lookahead methods.
const fs=require('node:fs'),vm=require('node:vm'),crypto=require('node:crypto'),path=require('node:path');
const base=process.argv[2], output=process.argv[3]||path.join(__dirname,'unicode_prime_collector_mathjax_3_2_2.json');
if(!base)throw Error('usage: node generate_unicode_prime_collector.cjs /pinned/d2latex [output.json]');
const hashes={'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01','mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869','setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'};
const c=vm.createContext({console});for(const[name,h]of Object.entries(hashes)){const b=fs.readFileSync(path.join(base,name));if(crypto.createHash('sha256').update(b).digest('hex')!==h)throw Error('Unpinned '+name);vm.runInContext(b.toString(),c,{filename:name})}
const data=vm.runInContext(`(()=>{
const methods=MathJax._.input.tex.base.BaseMethods.default,proto=MathJax._.input.tex.TexParser.default.prototype,nf=html.inputJax[0].parseOptions.nodeFactory,f=html.inputJax[0].mmlFactory;
const strings=["'",'’',"'’", "’'", '’’’', "'’'’", "’'’'’",'’α','’𝑥',"’ \\t'’z","’\\u00a0'’z","’% comment\\n'",'’\\\\limits', '’‘', '’′', '’–', "'\\uFEFF'", "'\\u0085'", "’\\uFEFF’", "’\\u0085’"];
const rows=[];for(const source of strings)for(const reject of [false,true]){
const core=f.create('mi',{},[f.create('text').setText('x')]);const base=reject?f.create('msubsup',{},[core,f.create('mi',{},[f.create('text').setText('i')]),f.create('mi',{},[f.create('text').setText('n')])]):core;
const before=base.toString();let pair;const p={string:source,i:1,GetNext:proto.GetNext,nextIsSpace:proto.nextIsSpace,getCodePoint:proto.getCodePoint,stack:{Prev:()=>base},create:(...a)=>nf.create(...a),itemFactory:{create:(name,...nodes)=>({name,nodes})},Push:item=>pair=item};
let error=null;try{methods.Prime(p,source[0])}catch(e){error=e.message}
rows.push({source,reject,error,cursorUTF16:p.i,remaining:source.slice(p.i),baseUnchanged:base.toString()===before,baseIdentity:pair?pair.nodes[0]===base:null,token:pair?pair.nodes[1].getText():null,attributes:pair?pair.nodes[1].attributes.getAllAttributes():null,properties:pair?pair.nodes[1].getAllProperties():null});}
return{primeSource:methods.Prime.toString(),getNextSource:proto.GetNext.toString(),nextIsSpaceSource:proto.nextIsSpace.toString(),getCodePointSource:proto.getCodePoint.toString(),rows};})()`,c);
fs.writeFileSync(output,JSON.stringify({assetsSHA256:hashes,...data},null,2)+'\n');console.log(data.rows.length+' registered collector records');
