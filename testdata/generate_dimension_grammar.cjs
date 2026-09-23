const fs=require('fs'),vm=require('vm'),path=require('path'),crypto=require('crypto');
const D=process.argv[3]||__dirname,base=process.argv[2];if(!base)throw Error('usage: node generate_dimension_grammar.cjs PINNED_ASSETS [EVIDENCE]');fs.mkdirSync(D,{recursive:true});
if(process.env.NODE_OPTIONS)throw Error('Normal runtime only; unexpected NODE_OPTIONS');
const sha=x=>crypto.createHash('sha256').update(x).digest('hex');
const hashes={'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01','mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869','setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'};
const assets=Object.entries(hashes).map(([f,h])=>{const b=fs.readFileSync(path.join(base,f));if(sha(b)!==h)throw Error('Unpinned '+f);return[f,b.toString()]});
const body="(()=>{\n const r=request.registration;const methods=MathJax._.input.tex.base.BaseMethods.default;\n if(r){const {Macro}=MathJax._.input.tex.Symbol;html.inputJax[0].configuration.handlers.retrieve('ams-declare-ops').add(r.name,new Macro(r.name,methods.Macro,[r.body,r.arguments]));}\n const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));\n const full=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},properties:pairs(n.getAllProperties()),children:n.childNodes.map(full)});\n let tree;const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){tree=full(math.root);return original.call(this,math,doc)};\n let svg=null,error=null;try{svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));}catch(e){error={id:e.id||null,message:e.message||String(e),type:e.constructor?.name||null};}\n return{tree,svg,error,trace:globalThis.d070Trace||null};\n})()\n";
const project=n=>({kind:n.kind,text:n.text,attributes:Object.fromEntries(n.attributes.explicit.map(p=>[p.name,p.value])),properties:Object.fromEntries(n.properties.map(p=>[p.name,p.value])),children:n.children.map(project)});
const cases=JSON.parse(fs.readFileSync(path.join(__dirname,'dimension_grammar_inputs.json'))).cases,records=[];
for(const c of cases){
 const ctx=vm.createContext({console});for(const[f,b]of assets)vm.runInContext(b,ctx,{filename:f});ctx.request=c;
 const r=vm.runInContext(body,ctx,{filename:'unmodified-primary-observer.js'});if(r.error||r.svg===null||!r.tree)throw Error(JSON.stringify(r.error));
 if(process.argv[3]){fs.writeFileSync(path.join(D,c.name+'.svg'),r.svg);fs.writeFileSync(path.join(D,c.name+'.json'),JSON.stringify(r.tree,null,2)+'\n');}
 records.push({...c,svgSHA256:sha(r.svg),tree:project(r.tree)});
}
fs.writeFileSync(path.join(D,'dimension_grammar_mathjax_3_2_2.json'),JSON.stringify({mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',cases:records},null,2)+'\n');console.log(records.length+' unmodified primary complete SVG/own-tree observations');
