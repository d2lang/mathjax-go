const fs=require('node:fs'),vm=require('node:vm'),path=require('node:path'),crypto=require('node:crypto');
if(process.env.NODE_OPTIONS)throw Error('Normal Node required');
const A=process.argv[2],D=process.argv[3],h=b=>crypto.createHash('sha256').update(b).digest('hex');if(!A||!D)throw Error('usage: node generator pinned-assets output-dir');fs.mkdirSync(D,{recursive:true});
const hashes={"polyfills.js": "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01", "mathjax.js": "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869", "setup.js": "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881"};
const assets=Object.entries(hashes).map(([n,expected])=>{const b=fs.readFileSync(path.join(A,n));if(h(b)!==expected)throw Error(n);return[n,b.toString()]});
const rows=[];
for(const request of JSON.parse(fs.readFileSync(path.join(__dirname,'numeric_script_inputs.json')))){
 if(request.registration)continue;
 const context=vm.createContext({console});for(const[n,b]of assets)vm.runInContext(b,context,{filename:n});context.request=request;
 const row=vm.runInContext(`(()=>{
  const T=MathJax._.input.tex.TexParser.default,B=MathJax._.input.tex.base.BaseMethods.default,r=request.registration;
  if(r){const M=MathJax._.input.tex.Symbol.Macro;html.inputJax[0].configuration.handlers.retrieve('ams-declare-ops').add(r.name,new M(r.name,B.Macro,[r.body,r.arguments]));}
  const original=T.prototype.Parse,trace=[];
  T.prototype.Parse=function(...args){
   const rec={sourceBefore:this.string,cursorBeforeUTF16:this.i,envBefore:{...this.stack.env},macroCountBefore:this.macroCount};trace.push(rec);
   try{return Reflect.apply(original,this,args)}catch(e){rec.error={id:e.id||null,message:e.message};throw e}finally{rec.sourceAfter=this.string;rec.cursorAfterUTF16=this.i;rec.remaining=this.string.slice(this.i);rec.envAfter={...this.stack.env};rec.macroCountAfter=this.macroCount;}
  };
  const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));
  return {trace,svg};
 })()`,context);
 const expected=JSON.parse(fs.readFileSync(path.join(__dirname,'numeric_script_mathjax_3_2_2.json'))).cases.find(c=>c.name===request.name).svgSHA256;
 if(h(row.svg)!==expected)throw Error('instrumentation changed paint: '+request.name);
 rows.push({name:request.name,tex:request.tex,display:request.display,svgSHA256:expected,trace:row.trace});
}
fs.writeFileSync(path.join(D,'numeric_script_cursors.json'),JSON.stringify({mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',rows:rows.map(c=>({name:c.name,tex:c.tex,display:c.display,...c.trace[c.trace.length-1]}))},null,2)+'\n');console.log(rows.length+' delegated Parse cursor/source/error traces, SVG bytes unchanged');
