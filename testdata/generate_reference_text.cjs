// SPDX-License-Identifier: Apache-2.0
// Actual pinned HandleRef and public compiler; observers delegate all methods.
// Run in a scratch copy: node generate_reference_text.cjs PINNED_ASSETS OUTPUT.
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),crypto=require('node:crypto');
const base=process.argv[2],out=process.argv[3],D=path.join(__dirname,'reference_text');
if(!base||!out||process.env.NODE_OPTIONS)throw Error('Expected pinned assets, scratch output, normal Node');
const sha=x=>crypto.createHash('sha256').update(x).digest('hex');
const pins={'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01','mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869','setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'};
const assets=Object.entries(pins).map(([n,h])=>{const b=fs.readFileSync(path.join(base,n));if(sha(b)!==h)throw Error('Unpinned '+n);return[n,b.toString()]});
const pairs=x=>Object.fromEntries(x.map(p=>[p.name,p.value]));
const project=n=>({kind:n.kind,text:n.text,attributes:pairs(n.attributes.explicit),properties:pairs(n.properties),children:n.children.map(project)});
const write=(n,x)=>fs.writeFileSync(path.join(out,n),JSON.stringify(x,null,2)+'\n');
fs.mkdirSync(out,{recursive:true});
const run=(input,body)=>{const ctx=vm.createContext({console});for(const[n,s]of assets)vm.runInContext(s,ctx,{filename:n});ctx.request=input;return vm.runInContext(body,ctx)};
const pub=JSON.parse(fs.readFileSync(path.join(D,'public-inputs.json'))).cases;
const priv=JSON.parse(fs.readFileSync(path.join(D,'private-inputs.json'))).cases;
if(pub.length!==29||priv.length!==17)throw Error('Changed bounded source contract');
const publicBody=fs.readFileSync(path.join(D,'public-primary-body.js'),'utf8'),privateBody=fs.readFileSync(path.join(D,'private-primary-body.js'),'utf8');
const publicRows=pub.map(c=>{const r=run(c,publicBody);write('public-'+c.name+'.json',r);fs.writeFileSync(path.join(out,c.name+'.svg'),r.svg);return{...c,svgSHA256:sha(r.svg),tree:project(r.tree)}});
const privateRows=priv.map(c=>{const r=run(c,privateBody);write('private-'+c.name+'.json',r);if(r.observations.length!==1)throw Error('Changed method count');const o=r.observations[0];if(JSON.stringify(o.before)!==JSON.stringify(o.after))throw Error('Changed retained tag state');return{input:c,nodes:o.pushed.map(x=>project(x.full)),error:o.error,outer:o.outer,parents:o.parents,before:o.before,after:o.after,made:o.made}});
const mathjaxGitCommit='ad8f5c21cb810236551da8c6512ba733e67357ee';
write('reference_text_mathjax_3_2_2.json',{mathjaxGitCommit,cases:publicRows});write('reference_text_method_mathjax_3_2_2.json',{mathjaxGitCommit,cases:privateRows});
write('manifest.json',{assetsSHA256:pins,runtime:process.version,publicConversions:29,actualMethodConversions:17,wholePublicPendingOutsideThisFixture:19});
