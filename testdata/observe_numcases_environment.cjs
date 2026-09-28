// SPDX-License-Identifier: Apache-2.0
// Run through generate_numcases_dynamic_helper.cjs after original asset validation.
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),readline=require('node:readline');
const assets=path.resolve(process.argv[2]);
const scripts=['polyfills.js','mathjax.js','setup.js'].map(n=>new vm.Script(fs.readFileSync(path.join(assets,n),'utf8'),{filename:n}));
(async()=>{for await(const line of readline.createInterface({input:process.stdin,crlfDelay:Infinity})){
if(!line)continue;const request=JSON.parse(line);const context=vm.createContext({console,request,observed:[]});context.globalThis=context;
for(const s of scripts)s.runInContext(context);
new vm.Script(`(() => {
 const Parser=MathJax._.input.tex.TexParser.default;
 const original=Parser.prototype.Parse;
 Parser.prototype.Parse=function(){
  if(this.string.endsWith('\\\\empheqlbrace\\\\,')) {
   const env={};for(const [key,value]of Object.entries(this.stack.env))env[key]=value instanceof RegExp ? String(value) : value;
   observed.push({source:this.string,env,macroCount:this.macroCount,top:this.stack.Top().kind});
  }
  return original.call(this);
 };
})()`).runInContext(context);
let output;try{output=new vm.Script(`({svg:adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}))})`).runInContext(context);}catch(e){output={error:e.stack};}
process.stdout.write(JSON.stringify({original:output,observed:context.observed})+'\n');
}})();
