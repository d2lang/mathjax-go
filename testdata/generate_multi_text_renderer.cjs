// SPDX-License-Identifier: Apache-2.0
// Constructed MML is supplied through a compile stub; output methods are unmodified.
const fs=require('node:fs'), path=require('node:path'), vm=require('node:vm'), crypto=require('node:crypto');
const A=process.argv[3] || __dirname, base=process.argv[2];
fs.mkdirSync(A,{recursive:true});
if(!base)throw Error('usage: node capture.cjs /pinned/d2latex');
const sha=b=>crypto.createHash('sha256').update(b).digest('hex');
const hashes={
 'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
 'mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
 'setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'};
const assets=Object.entries(hashes).map(([name,h])=>{let b=fs.readFileSync(path.join(base,name));if(sha(b)!==h)throw Error('asset drift '+name);return[name,b.toString()]});
function context(){const c=vm.createContext({console});for(const[n,s]of assets)vm.runInContext(s,c,{filename:n});return c;}
function save(name,value){let p=path.join(A,name);fs.mkdirSync(path.dirname(p),{recursive:true});fs.writeFileSync(p,typeof value==='string'?value:JSON.stringify(value,null,2)+'\n');}

// An explicit value codec retains own undefined, null, false, zero and strings.
function observeRuntime(){
 const ids=new Map();let serial=0;
 const id=n=>n==null?null:(ids.has(n)?ids.get(n):(ids.set(n,'n'+(++serial)),ids.get(n)));
 const encode=v=>v===undefined?{$type:'undefined'}:typeof v==='number'&&!Number.isFinite(v)?{$type:'number',value:String(v)}:v;
 const pairs=o=>Object.keys(o).map(name=>({name,value:encode(o[name])}));
 const tree=n=>({id:id(n),kind:n.kind,text:n.kind==='text'?n.getText():null,
  attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:null,
  properties:pairs(n.getAllProperties()),texClass:encode(n.texClass),parent:id(n.parent),children:n.childNodes.map(tree)});
 const plain=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,
  attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:null,
  properties:pairs(n.getAllProperties()),texClass:encode(n.texClass),children:n.childNodes.map(plain)});
 const list=root=>{const out=[];function walk(n){if(n.kind==='mo')out.push(n);n.childNodes.forEach(walk)}walk(root);return out};
 const live=(n,root)=>{while(n&&n!==root)n=n.parent;return !!n};
 const snapshot=(options,tracked=[])=>JSON.parse(JSON.stringify({tree:tree(options.root),plain:plain(options.root),
  registered:(options.nodeLists.mo||[]).map(id),liveRegistered:(options.nodeLists.mo||[]).filter(n=>live(n,options.root)).map(id),
  preorder:list(options.root).map(id),tracked:tracked.map(([label,n])=>({label,...tree(n)}))}));
 return {ids,id,encode,pairs,tree,plain,list,live,snapshot};
}


function renderRuntime(spec){
 const O=observeRuntime(),options=html.inputJax[0].parseOptions; options.clear();
 const nf=options.nodeFactory;
 const texts=spec.parts.map((s,i)=>{const n=nf.create('text',s);O.ids.set(n,'text'+i);return n});
 const token=nf.create('node',spec.kind,texts,spec.attrs);O.ids.set(token,'token');
 const root=nf.create('node','math',[token],spec.display?{display:'block'}:{});O.ids.set(root,'math');
 root.setInheritedAttributes({},spec.display,0,false); options.root=root;
 const before=O.snapshot(options,[['token',token],...texts.map((n,i)=>['text'+i,n])]);
 const compile=html.inputJax[0].compile;html.inputJax[0].compile=()=>root;
 const svg=adaptor.innerHTML(html.convert('',{display:spec.display,em:16,ex:8}));
 const repeated=adaptor.innerHTML(html.convert('',{display:spec.display,em:16,ex:8}));
 if(svg!==repeated)throw Error('Repeated primary renderer output changed: '+spec.name);
 html.inputJax[0].compile=compile;
 const after=O.snapshot(options,[['token',token],...texts.map((n,i)=>['text'+i,n])]);
 return {spec,before,after,svg};
}
const records=[];
const project=n=>({kind:n.kind,text:n.text,attributes:n.attributes?n.attributes.explicit:[],properties:n.properties,children:n.children.map(project)});
for(const spec of [
  {
    "name": "mo-multi-inline",
    "kind": "mo",
    "parts": [
      "=",
      "<"
    ],
    "display": false,
    "attrs": {}
  },
  {
    "name": "mo-multi-display",
    "kind": "mo",
    "parts": [
      "=",
      "<"
    ],
    "display": true,
    "attrs": {}
  },
  {
    "name": "mo-single-inline",
    "kind": "mo",
    "parts": [
      "=<"
    ],
    "display": false,
    "attrs": {}
  },
  {
    "name": "mo-single-display",
    "kind": "mo",
    "parts": [
      "=<"
    ],
    "display": true,
    "attrs": {}
  },
  {
    "name": "mi-multi-inline",
    "kind": "mi",
    "parts": [
      "a",
      "b"
    ],
    "display": false,
    "attrs": {}
  },
  {
    "name": "mi-multi-display",
    "kind": "mi",
    "parts": [
      "a",
      "b"
    ],
    "display": true,
    "attrs": {}
  },
  {
    "name": "mi-single-inline",
    "kind": "mi",
    "parts": [
      "ab"
    ],
    "display": false,
    "attrs": {}
  },
  {
    "name": "mi-single-display",
    "kind": "mi",
    "parts": [
      "ab"
    ],
    "display": true,
    "attrs": {}
  },
  {
    "name": "mn-multi-inline",
    "kind": "mn",
    "parts": [
      "1",
      "2"
    ],
    "display": false,
    "attrs": {}
  },
  {
    "name": "mn-multi-display",
    "kind": "mn",
    "parts": [
      "1",
      "2"
    ],
    "display": true,
    "attrs": {}
  },
  {
    "name": "mn-single-inline",
    "kind": "mn",
    "parts": [
      "12"
    ],
    "display": false,
    "attrs": {}
  },
  {
    "name": "mn-single-display",
    "kind": "mn",
    "parts": [
      "12"
    ],
    "display": true,
    "attrs": {}
  },
  {
    "name": "mtext-multi-inline",
    "kind": "mtext",
    "parts": [
      "a",
      "b"
    ],
    "display": false,
    "attrs": {}
  },
  {
    "name": "mtext-multi-display",
    "kind": "mtext",
    "parts": [
      "a",
      "b"
    ],
    "display": true,
    "attrs": {}
  },
  {
    "name": "mtext-single-inline",
    "kind": "mtext",
    "parts": [
      "ab"
    ],
    "display": false,
    "attrs": {}
  },
  {
    "name": "mtext-single-display",
    "kind": "mtext",
    "parts": [
      "ab"
    ],
    "display": true,
    "attrs": {}
  },
  {
    "name": "ms-quotes-inline",
    "kind": "ms",
    "parts": [
      "ab"
    ],
    "display": false,
    "attrs": {}
  },
  {
    "name": "ms-quotes-display",
    "kind": "ms",
    "parts": [
      "ab"
    ],
    "display": true,
    "attrs": {}
  },
  {
    "name": "mo-empty-first-inline",
    "kind": "mo",
    "parts": [
      "",
      "="
    ],
    "display": false,
    "attrs": {}
  },
  {
    "name": "mo-empty-first-display",
    "kind": "mo",
    "parts": [
      "",
      "="
    ],
    "display": true,
    "attrs": {}
  },
  {
    "name": "mo-explicit-font-inline",
    "kind": "mo",
    "parts": [
      "=",
      "<"
    ],
    "display": false,
    "attrs": {
      "fontfamily": "serif"
    }
  },
  {
    "name": "mo-explicit-font-display",
    "kind": "mo",
    "parts": [
      "=",
      "<"
    ],
    "display": true,
    "attrs": {
      "fontfamily": "serif"
    }
  },
  {
    "name": "mo-three-children-inline",
    "kind": "mo",
    "parts": [
      "=",
      "<",
      ">"
    ],
    "attrs": {},
    "display": false
  },
  {
    "name": "mo-three-children-display",
    "kind": "mo",
    "parts": [
      "=",
      "<",
      ">"
    ],
    "attrs": {},
    "display": true
  },
  {
    "name": "mo-empty-middle-inline",
    "kind": "mo",
    "parts": [
      "=",
      "",
      ">"
    ],
    "attrs": {},
    "display": false
  },
  {
    "name": "mo-empty-middle-display",
    "kind": "mo",
    "parts": [
      "=",
      "",
      ">"
    ],
    "attrs": {},
    "display": true
  },
  {
    "name": "mo-empty-last-inline",
    "kind": "mo",
    "parts": [
      "=",
      ""
    ],
    "attrs": {},
    "display": false
  },
  {
    "name": "mo-empty-last-display",
    "kind": "mo",
    "parts": [
      "=",
      ""
    ],
    "attrs": {},
    "display": true
  },
  {
    "name": "mo-multichar-first-inline",
    "kind": "mo",
    "parts": [
      "=<",
      ">"
    ],
    "attrs": {},
    "display": false
  },
  {
    "name": "mo-multichar-first-display",
    "kind": "mo",
    "parts": [
      "=<",
      ">"
    ],
    "attrs": {},
    "display": true
  },
  {
    "name": "mo-bold-inline",
    "kind": "mo",
    "parts": [
      "=",
      "<"
    ],
    "attrs": {
      "mathvariant": "bold"
    },
    "display": false
  },
  {
    "name": "mo-bold-display",
    "kind": "mo",
    "parts": [
      "=",
      "<"
    ],
    "attrs": {
      "mathvariant": "bold"
    },
    "display": true
  }
]){
 const c=context(); c.spec=spec;vm.runInContext('const observeRuntime='+observeRuntime.toString(),c);
 const r=vm.runInContext('('+renderRuntime.toString()+')(spec)',c);save('primary/'+spec.name+'.svg',r.svg);const svg=r.svg;delete r.svg;save('primary/'+spec.name+'.json',r);records.push({...spec,svg,svgSHA256:sha(svg),before:project(r.before.tree),after:project(r.after.tree)});
}
save('multi_text_renderer_mathjax_3_2_2.json',{oracle:'Unmodified pinned MathJax3.2.2 output jax; explicitly constructed token graph, no TeX relation filter',primaryCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records});console.log(records.length+' actual primary constructed-token observations; no filter or method substitution');
