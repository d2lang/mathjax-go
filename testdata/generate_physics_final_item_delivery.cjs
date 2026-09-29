// SPDX-License-Identifier: Apache-2.0
// node --jitless testdata/generate_physics_final_item_delivery.cjs PINNED_ASSETS
// Only the hash-verified original creates output references; no Go process.
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto'),vm=require('node:vm'),zlib=require('node:zlib');
const file=path.join(__dirname,'physics_final_item_delivery_mathjax_3_2_2.json.gz');
const data=JSON.parse(zlib.gunzipSync(fs.readFileSync(file))),assets=process.argv[2];
if(!assets||data.mathjaxGitCommit!=='ad8f5c21cb810236551da8c6512ba733e67357ee'||data.cases.length!==272)throw Error('unbound original corpus');
const scripts=Object.entries(data.originalAssets).map(([name,hash])=>{
 const bytes=fs.readFileSync(path.join(assets,name));
 if(crypto.createHash('sha256').update(bytes).digest('hex')!==hash)throw Error('unverified '+name);
 return new vm.Script(bytes.toString(),{filename:name});
});
for(const row of data.cases){
 const request={tex:row.tex,display:row.display,options:{display:row.display,Display:row.display}};
 const context=vm.createContext({request,console});scripts.forEach(script=>script.runInContext(context));
 let original;
 try{original=vm.runInContext('({svg:adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}))})',context);}
 catch(error){original={error:String(error.stack||error)};}
 row.original=original;
}
fs.writeFileSync(file,zlib.gzipSync(JSON.stringify(data,null,2)+'\n'));
console.log(JSON.stringify({cases:data.cases.length,svg:data.cases.filter(c=>c.original.svg).length,runtime:data.cases.filter(c=>c.original.error).length}));
