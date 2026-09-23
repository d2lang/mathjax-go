const fs=require('fs'),vm=require('vm'),path=require('path'),crypto=require('crypto');
const D=__dirname,base=process.argv[2];
if(process.env.NODE_OPTIONS)throw Error('Normal runtime only; unexpected NODE_OPTIONS');
const sha=x=>crypto.createHash('sha256').update(x).digest('hex');
const hashes={'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01','mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869','setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'};
if(!base||!process.argv[3])throw Error('usage: node generate_infix_scope.cjs PINNED_ASSETS EVIDENCE_DIRECTORY');
const assets=Object.entries(hashes).map(([f,h])=>{const b=fs.readFileSync(path.join(base,f));if(sha(b)!==h)throw Error('Unpinned '+f);return[f,b.toString()]});
const body=fs.readFileSync(path.join(D,'infix_scope_primary_body.js'),'utf8');const results=[],references=[];const project=n=>({kind:n.kind,text:n.text,attributes:n.attributes.explicit,properties:n.properties,children:n.children.map(project)});
const traced=false,out=process.argv[3];fs.mkdirSync(out,{recursive:true});
for(const c of JSON.parse(fs.readFileSync(path.join(D,'infix_scope_inputs.json'))).cases){
 const ctx=vm.createContext({console});for(const[f,b]of assets)vm.runInContext(b,ctx,{filename:f});ctx.request=c;

 const r=vm.runInContext(body,ctx,{filename:'infix_scope_primary_body.js'});if(r.svg!==null)fs.writeFileSync(path.join(out,c.name+'.svg'),r.svg);
 fs.writeFileSync(path.join(out,c.name+'.json'),JSON.stringify(r.tree||null,null,2)+'\n');
 if(traced)fs.writeFileSync(path.join(out,c.name+'.trace.json'),JSON.stringify(r.trace,null,2)+'\n');
 results.push({...c,error:r.error,formattedError:r.formattedError,parserError:r.parserError,svgSHA256:r.svg===null?null:sha(r.svg),treeSHA256:sha(JSON.stringify(r.tree||null,null,2)+'\n')});
 const reference={...c,svgSHA256:sha(r.svg),tree:project(r.tree),error:r.formattedError};if(r.parserError&&!c.registration){if(!/^[\x00-\x7f]*$/.test(r.parserError.source))throw Error('Unspecified UTF16/byte cursor mapping');reference.cursor={source:r.parserError.source,byteCursor:r.parserError.cursorUTF16,remaining:r.parserError.remaining}}references.push(reference);
}
fs.writeFileSync(path.join(out,'results.json'),JSON.stringify({assetsSHA256:hashes,runtime:process.version,traced,cases:results},null,2)+'\n');console.log(results.length+' captured '+(traced?'actual primary event traces':'unmodified primary outputs'));

fs.writeFileSync(path.join(out,'infix_scope_mathjax_3_2_2.json'),JSON.stringify({oracle:'Untouched pinned MathJax3.2.2 with delegating pre-render/error observers',assetsSHA256:hashes,cases:references},null,2)+'\n');
