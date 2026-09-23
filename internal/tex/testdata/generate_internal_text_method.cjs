// Read-only calls to the actual registered internalMath helper and TexParser.
const fs = require('node:fs'), vm = require('node:vm'), path = require('node:path'), crypto = require('node:crypto');
const base = process.argv[2];
if (!base) throw Error('usage: node generate_internal_text_method.cjs /path/to/pinned/d2latex');
if (process.env.NODE_OPTIONS) throw Error('Expected normal Node');
const hashes = {'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01','mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869','setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'};
const assets = Object.entries(hashes).map(([name, h]) => {const b=fs.readFileSync(path.join(base,name));if(crypto.createHash('sha256').update(b).digest('hex')!==h)throw Error(name);return [name,b.toString()];});
const cases = [
  {name:'empty',text:''},{name:'literal',text:'abc'},{name:'all-space',text:'   '},
  {name:'edges',text:'  a  b  '},{name:'bom',text:'\ufeffa\ufeff'},{name:'nel',text:'\u0085a\u0085'},
  {name:'tilde',text:'a~b'},{name:'escaped',text:'\\$x\\$ \\{a\\} \\\\ b'},
  {name:'literal-command',text:'\\alpha \\relax'},
  {name:'dollar-only',text:'$x$'},{name:'mixed',text:'a $x_i$ b $y$ c'},
  {name:'parenthesis',text:'a \\(x+y\\) b'},{name:'nested-braces',text:'a ${x_i}$ b'},
  {name:'empty-level-zero',text:'',level:0},{name:'mixed-level-zero',text:'a $x$ b',level:0},
  {name:'ambient-font',text:'a $x+\\mathit{yz}$ b',env:{font:'bold'}},
  {name:'explicit-normal',text:'a $x$ b',env:{font:'bold'},font:'normal'},
  {name:'explicit-italic',text:'a $x$ b',font:'italic'},
  {name:'missing-dollar',text:'a $x'},{name:'missing-parenthesis',text:'a \\(x'},
  {name:'completed-inner-error',text:'a $x\\relax$ b'},
  {name:'reference',text:'a \\ref{missing} b',font:'italic'},
  {name:'equation-reference',text:'a \\eqref {missing} b',env:{font:'bold'}},
  {name:'registered-inner',text:'a $\\dproof$ b',registration:{name:'dproof',body:'\\alpha',arguments:0}},
  {name:'registered-literal',text:'a \\dproof b',registration:{name:'dproof',body:'\\alpha',arguments:0}},
  {name:'outer-macro-limit',text:'$\\dproof$ $\\dproof$',outerMacroCount:10000,registration:{name:'dproof',body:'x',arguments:0}},
  {name:'inner-error-counter',text:'$\\dproof$',outerMacroCount:9999,registration:{name:'dproof',body:'\\relax',arguments:0}},
  {name:'supplementary',text:'a𝛼 $x$ 𝛽'},
];
const rows = [];
for (const input of cases) {
  const ctx=vm.createContext({console});for(const [name,source]of assets)vm.runInContext(source,ctx,{filename:name});ctx.request=input;
  rows.push(vm.runInContext(`(() => {
    const config=html.inputJax[0].parseOptions;
    const TexParser=MathJax._.input.tex.TexParser.default;
    const util=MathJax._.input.tex.ParseUtil.default;
    const methods=MathJax._.input.tex.base.BaseMethods.default;
    if(request.registration){const r=request.registration;const Macro=MathJax._.input.tex.Symbol.Macro;html.inputJax[0].configuration.handlers.retrieve('ams-declare-ops').add(r.name,new Macro(r.name,methods.Macro,[r.body,r.arguments]));}
    const parser=new TexParser('',{},config);
    // Empty parsing has already reduced Stop and restored the stack env.
    // Supply the live caller environment explicitly for this direct helper call.
    Object.assign(parser.stack.env,request.env||{});
    parser.macroCount=request.outerMacroCount||0;
    const originalCreate=config.nodeFactory.create, calls=[];
    config.nodeFactory.create=function(...args){const n=Reflect.apply(originalCreate,this,args);calls.push({type:args[0],kind:n?.kind||null,node:n});return n;};
    const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?n.attributes.getAllAttributes():{},properties:n.getAllProperties(),children:n.childNodes.map(full)});
    const envBefore={...parser.stack.env},counterBefore=parser.macroCount;
    let nodes=null,error=null;
    try{nodes=util.internalMath(parser,request.text,request.level,request.font);}catch(e){error={id:e.id||null,message:e.message||String(e)};}
    finally{config.nodeFactory.create=originalCreate;}
    const created=new Set(calls.map(c=>c.node));
    const result={input:request,error,nodes:nodes?.map(full)||null,creation:calls.map(({type,kind})=>({type,kind})),returnedNodesCreated:nodes?nodes.every(n=>created.has(n)):null,envBefore,envAfter:{...parser.stack.env},counterBefore,counterAfter:parser.macroCount,currentParserIsOuter:config.parser===parser};
    // Closing the outer parser is diagnostic teardown, after recording state.
    try{parser.mml();}catch(e){result.teardownError=e.message;}
    return result;
  })()`,ctx));
}
process.stdout.write(JSON.stringify({runtime:process.version,assetsSHA256:hashes,rows},null,2)+'\n');
