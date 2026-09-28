const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),readline=require('node:readline'),crypto=require('node:crypto');
const assets=path.resolve(process.argv[2]);
const expected=['7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01','cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869','a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'];
const scripts=['polyfills.js','mathjax.js','setup.js'].map((f,i)=>{const b=fs.readFileSync(path.join(assets,f));if(crypto.createHash('sha256').update(b).digest('hex')!==expected[i])throw Error('unverified '+f);return new vm.Script(b.toString(),{filename:f});});
(async()=>{for await(const line of readline.createInterface({input:process.stdin,crlfDelay:Infinity})){
if(!line)continue;const request=JSON.parse(line);const context=vm.createContext({console,request,observed:[]});context.globalThis=context;
for(const script of scripts)script.runInContext(context);
new vm.Script(`(() => {
 const jax=html.inputJax[0], ids=new WeakMap();let next=1;
 const id=n=>{if(!n)return null;if(!ids.has(n))ids.set(n,next++);return ids.get(n)};
 const item=(n,root)=>{let p=n,chain=[];for(let i=0;p&&i<64;i++,p=p.parent){chain.push({id:id(p),kind:p.kind});if(p===root)break;}return {id:id(n),kind:n.kind,children:n.childNodes.map(c=>c?{id:id(c),kind:c.kind,parent:id(c.parent)}:null),parents:chain,live:p===root,attributes:n.attributes.getAllAttributes(),properties:n.getAllProperties()};};
 const snapshot=(phase,options)=>{const lists={};for(const k of ['msubsup','munderover'])lists[k]=(options.nodeLists[k]||[]).map(n=>item(n,options.root));observed.push({phase,error:options.error,root:id(options.root),lists});};
 jax.postFilters.add(arg=>snapshot('before-cleanSubSup',arg.data),-6.1);
 jax.postFilters.add(arg=>snapshot('after-cleanSubSup',arg.data),-5.9);
 const util=MathJax._.input.tex.NodeUtil.default,old=util.copyAttributes;
 util.copyAttributes=function(src,dst){old.call(this,src,dst);observed.push({phase:'copyAttributes',oldID:id(src),newID:id(dst),oldKind:src.kind,newKind:dst.kind,sameAttributes:src.attributes===dst.attributes,sameProperties:src.properties===dst.properties,oldProperties:src.getAllProperties(),newProperties:dst.getAllProperties(),oldChildren:src.childNodes.map(id),newChildren:dst.childNodes.map(id)});};
})()`).runInContext(context);
let original=new vm.Script(`(() => {try{return {svg:adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}))};}catch(error){return {error:String(error.stack)};}})()`).runInContext(context);
process.stdout.write(JSON.stringify({original,observed:context.observed})+'\n');
}})();
