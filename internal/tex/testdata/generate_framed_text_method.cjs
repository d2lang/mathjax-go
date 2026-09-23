// Read-only calls to the actual pinned framed-text handlers and TexParser.
const fs = require('node:fs'), vm = require('node:vm'), path = require('node:path'), crypto = require('node:crypto');
const base = process.argv[2];
if (!base) throw Error('usage: node generate_framed_text_method.cjs /path/to/pinned/d2latex');
if (process.env.NODE_OPTIONS) throw Error('Expected normal Node');
const hashes = {'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01','mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869','setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'};
const assets = Object.entries(hashes).map(([name, h]) => {const b=fs.readFileSync(path.join(base,name));if(crypto.createHash('sha256').update(b).digest('hex')!==h)throw Error(name);return [name,b.toString()];});

const cases = [
 {name:'fbox-empty',command:'fbox',source:'{}Z'},
 {name:'colorbox-mixed',command:'colorbox',source:'{red}{a $x$ b}Z'},
 {name:'fcolorbox-mixed',command:'fcolorbox',source:'{blue}{red}{a $x$ b}Z'},
 {name:'font-environment',command:'fbox',source:'{a $x$ b}Z',font:'bold'},
 {name:'shared-macro-success',command:'fbox',source:'{a $\\dproof$ b}Z',outerMacroCount:999,registration:{name:'dproof',body:'x',arguments:0}},
 {name:'shared-macro-error',command:'colorbox',source:'{late}{a $\\dproof$}Z',outerMacroCount:999,registration:{name:'dproof',body:'\\definecolor{late}{RGB}{255,0,0}\\notADefinedCommand',arguments:0}},
 {name:'late-color',command:'colorbox',source:'{late}{$\\definecolor{late}{RGB}{255,0,0}x$}Z'},
 {name:'late-frame',command:'fcolorbox',source:'{late}{yellow}{$\\definecolor{late}{RGB}{0,0,255}x$}Z'},
 {name:'fbox-missing',command:'fbox',source:''},
 {name:'colorbox-missing',command:'colorbox',source:'{red}'},
 {name:'fcolorbox-missing',command:'fcolorbox',source:'{blue}{red}'},
 {name:'incomplete-math',command:'fbox',source:'{a $x}Z'},
];
const rows=[];
for(const input of cases){
 const ctx=vm.createContext({console});for(const [name,source]of assets)vm.runInContext(source,ctx,{filename:name});ctx.request=input;
 rows.push(vm.runInContext(`(() => {
  const config=html.inputJax[0].parseOptions, TexParser=MathJax._.input.tex.TexParser.default;
  const baseMethods=MathJax._.input.tex.base.BaseMethods.default, colors=MathJax._.input.tex.color.ColorMethods.ColorMethods;
  const methods={fbox:baseMethods.FBox,colorbox:colors.ColorBox,fcolorbox:colors.FColorBox};
  if(request.registration){const r=request.registration,Macro=MathJax._.input.tex.Symbol.Macro;html.inputJax[0].configuration.handlers.retrieve('ams-declare-ops').add(r.name,new Macro(r.name,baseMethods.Macro,[r.body,r.arguments]));}
  const parser=new TexParser('',{},config);parser.string=request.source;parser.i=0;parser.currentCS=String.fromCharCode(92)+request.command;
  if(request.font)parser.stack.env.font=request.font;parser.macroCount=request.outerMacroCount||0;
  const env=parser.stack.env,envBefore={...env},colorModel=config.packageData.get('color').model;
  const create=config.nodeFactory.create,made=[],pushed=[];config.nodeFactory.create=function(...args){const n=Reflect.apply(create,this,args);made.push(n);return n;};
  // Observe only the handler's final Push instead of resuming an already closed outer stack.
  // All argument reads, inner parsing, factories, and color accesses stay untouched.
  parser.Push=n=>pushed.push(n);
  let error=null;try{methods[request.command](parser,parser.currentCS);}catch(e){error={id:e.id||null,message:e.message||String(e)};}finally{config.nodeFactory.create=create;}
  const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?n.attributes.getAllAttributes():{},properties:n.getAllProperties(),children:n.childNodes.map(full)});
  const seen=new Set();let unique=true,parents=true;const walk=n=>{if(seen.has(n))unique=false;seen.add(n);for(const c of n.childNodes){if(c.parent!==n)parents=false;walk(c);}};pushed.forEach(walk);
  const outerKind=request.command==='fbox'?'menclose':'mpadded';
  return JSON.parse(JSON.stringify({input:request,error,nodes:pushed.map(full),cursor:parser.i,remaining:parser.string.slice(parser.i),sourceUnchanged:parser.string===request.source,envBefore,envAfter:{...parser.stack.env},envIdentity:parser.stack.env===env,configIdentity:parser.configuration===config,colorModelIdentity:config.packageData.get('color').model===colorModel,counter:parser.macroCount,lateColor:colorModel.getColor('named','late'),pushCount:pushed.length,outerCreationCount:made.filter(n=>n.kind===outerKind).length,allPushedCreated:pushed.every(n=>made.includes(n)),rootParentsNull:pushed.every(n=>n.parent===null),unique,parents,method:methods[request.command].toString()}));
 })()`,ctx));
}
process.stdout.write(JSON.stringify({runtime:process.version,assetsSHA256:hashes,rows},null,2)+'\n');
