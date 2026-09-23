// SPDX-License-Identifier: Apache-2.0
// Fresh unmodified pinned primary conversion; registered macros are labelled separately.
const fs = require('node:fs'), vm = require('node:vm'), path = require('node:path'), crypto = require('node:crypto');
const base = process.argv[2], evidence = process.argv[3];
if (process.env.NODE_OPTIONS) throw Error('Normal Node required');
if (!base || !evidence) throw Error('usage: node generate_unbraced_prime.cjs pinned-assets evidence-directory');
fs.mkdirSync(evidence, {recursive: true});
const hash = x => crypto.createHash('sha256').update(x).digest('hex');
const hashes = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'
};
const assets = Object.entries(hashes).map(([name, h]) => {
  const b = fs.readFileSync(path.join(base, name));
  if (hash(b) !== h) throw Error('Unpinned ' + name);
  return [name, b.toString()];
});
const project = n => ({kind:n.kind, text:n.text,
  attributes:Object.fromEntries(n.attributes.explicit.map(x=>[x.name,x.value])),
  properties:Object.fromEntries(n.properties.map(x=>[x.name,x.value])), children:n.children.map(project)});
const records = [];
for (const request of JSON.parse(fs.readFileSync(path.join(__dirname, 'numeric_script_inputs.json')))) {
  const context = vm.createContext({console});
  for (const [name, data] of assets) vm.runInContext(data, context, {filename:name});
  context.request = request;
  const result = vm.runInContext(`(() => {
    const r = request.registration, methods = MathJax._.input.tex.base.BaseMethods.default;
    if (r) { const {Macro} = MathJax._.input.tex.Symbol;
      html.inputJax[0].configuration.handlers.retrieve('ams-declare-ops').add(r.name, new Macro(r.name,methods.Macro,[r.body,r.arguments])); }
    const pairs = o => Object.keys(o).map(name=>({name,value:o[name]}));
    const full = n => ({kind:n.kind,text:n.kind==='text'?n.getText():null,
      attributes:n.attributes?{explicit:pairs(n.attributes.getAllAttributes()),inherited:pairs(n.attributes.getAllInherited()),defaults:pairs(n.attributes.getAllDefaults()),global:pairs(n.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},
      properties:pairs(n.getAllProperties()),children:n.childNodes.map(full)});
    let tree; const original = html.outputJax.typeset;
    html.outputJax.typeset = function(math,doc) { tree=full(math.root); return original.call(this,math,doc); };
    const svg = adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));
    return {tree,svg};
  })()`, context);
  fs.writeFileSync(path.join(evidence,request.name+'.svg'),result.svg);
  fs.writeFileSync(path.join(evidence,request.name+'.json'),JSON.stringify(result.tree,null,2)+'\n');
  const dimensions = result.svg.match(/^<svg[^>]+width="([0-9.]+)ex" height="([0-9.]+)ex"[^>]+>/);
  records.push({...request,svgSHA256:hash(result.svg),propertiesTree:project(result.tree),
    measureAvailable:!!dimensions, width:dimensions?Math.ceil(Number(dimensions[1])*8):0,
    height:dimensions?Math.ceil(Number(dimensions[2])*8):0});
}
fs.writeFileSync(path.join(evidence,'numeric_script_mathjax_3_2_2.json'),JSON.stringify({
  mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assetsSHA256:hashes,cases:records},null,2)+'\n');
console.log(records.length+' complete primary SVG/tree records');
