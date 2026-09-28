const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),crypto=require('node:crypto');
const assets=path.resolve(process.argv[2]), hashes={};const context=vm.createContext({console});context.globalThis=context;
for(const f of ['polyfills.js','mathjax.js','setup.js']){const b=fs.readFileSync(path.join(assets,f));hashes[f]=crypto.createHash('sha256').update(b).digest('hex');new vm.Script(b.toString(),{filename:f}).runInContext(context);}
const methods=new vm.Script(`(() => {
const util=MathJax._.input.tex.mathtools.MathtoolsUtil.MathtoolsUtil, factory=html.inputJax[0].parseOptions.nodeFactory;
const encode=v=>({type:typeof v,...(v===undefined?{}:{value:v})});const rows=[];
for(const kind of ['mtable','mrow'])for(const [label,value]of [['default',null],['undefined',undefined],['null',null],['empty',''],['false',false],['zero',0],['text-zero','0'],['spacing','1ex'],['whitespace',' ']])for(const [spreadLabel,spread]of [['undefined',undefined],['one-pt','1pt']]){
const n=factory.create('node',kind,[]);if(label!=='default')n.attributes.set('rowspacing',value);
const before=encode(n.attributes.get('rowspacing'));let error=null;try{util.spreadLines(n,spread)}catch(e){error=String(e.stack)}
rows.push({kind,label,spreadLabel,before,after:encode(n.attributes.get('rowspacing')),explicitKeys:Object.keys(n.attributes.getAllAttributes()),error});
}
return {source:String(util.spreadLines),rows};})()`).runInContext(context);
process.stdout.write(JSON.stringify({hashes,...methods},null,2)+'\n');
