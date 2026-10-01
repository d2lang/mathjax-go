// SPDX-License-Identifier: Apache-2.0
// Observe active source handlers without changing the frozen runtime.
const fs=require('node:fs'),vm=require('node:vm'),crypto=require('node:crypto');
const dir=process.argv[2]; if(!dir)throw Error('usage: node --jitless generate_getnext_bindings.cjs PINNED_ASSETS');
const hashes={'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01','mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869','setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'};
const context=vm.createContext({console});context.globalThis=context;
for(const [name,hash] of Object.entries(hashes)){const b=fs.readFileSync(dir+'/'+name);if(crypto.createHash('sha256').update(b).digest('hex')!==hash)throw Error('hash');new vm.Script(b.toString(),{filename:name}).runInContext(context);}
const tex=context.MathJax._.input.tex, maps=context.html.inputJax[0].parseOptions.handlers.get('macro');
const owners={Base:tex.base.BaseMethods.default,Braket:tex.braket.BraketMethods.default,Physics:tex.physics.PhysicsMethods.default};
const names=['matrix','pmatrix','cases','eqalign','eqalignno','leqalignno','displaylines','Braket','Set','set','dd','differential','variation','var','dv','derivative','pdv','pderivative','partialderivative','fdv','fderivative','functionalderivative','det','tr','trace','Tr','Trace','exp','Pr','erf','sin','ln','dmat','admat','diagonalmatrix','antidiagonalmatrix'];
const bindings={hashes,methods:{},commands:{}};
for(const name of ['GetNext','GetStar','nextIsSpace']){const d=Object.getOwnPropertyDescriptor(tex.TexParser.default.prototype,name);bindings.methods[name]=String(d?.value||d?.get||'absent');}
for(const name of names){const m=maps.lookup(name);let binding=[];if(m)for(const [owner,methods]of Object.entries(owners))for(const [method,value]of Object.entries(methods))if(value===m.func)binding.push(owner+'.'+method);bindings.commands[name]=m?{symbol:m.symbol,args:m.args,binding,source:m.func.toString()}:null;}
for(const [name,m]of Object.entries(bindings.commands))if(!m||!m.binding.length)throw Error('missing active original '+name);
// Authored state on the unmodified original TexParser prototype; every
// executed GetNext/nextIsSpace/getCodePoint method is the pinned original.
bindings.whitespaceObservations=[];
for(const point of [9,10,11,12,13,32,160,5760,8192,8193,8194,8195,8196,8197,8198,8199,8200,8201,8202,8232,8233,8239,8287,12288,65279,133,6158,8203]){
  const parser=Object.create(tex.TexParser.default.prototype);
  parser.string=String.fromCodePoint(point)+'x';parser.i=0;
  const isSpace=parser.nextIsSpace(),next=parser.GetNext(),cursor=parser.i;
  bindings.whitespaceObservations.push({point,isSpace,next,cursor});
}
console.log(JSON.stringify(bindings));
