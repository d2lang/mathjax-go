// SPDX-License-Identifier: Apache-2.0
// node --jitless generate_synthetic_operator_factory.cjs PINNED_ASSETS
// Original-only constructed-MathML observations; no candidate invocation.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const vm = require('node:vm');
const {spawnSync} = require('node:child_process');
const assets = process.argv[2];
if (!assets) throw Error('expected pinned asset directory');
const hashes = {
  'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'
};
const sha = data => crypto.createHash('sha256').update(data).digest('hex');
const scripts = Object.entries(hashes).map(([name, hash]) => {
  const data = fs.readFileSync(path.join(assets, name));
  if (sha(data) !== hash) throw Error(`unverified ${name}`);
  return new vm.Script(data.toString(), {filename:name});
});
const observe = new vm.Script(`(() => {
  const factory = html.inputJax[0].mmlFactory;
  const fakeObservations = [];
  const prototype = Object.getPrototypeOf(factory.create('mfenced'));
  const original = prototype.setChildInheritedAttributes;
  prototype.setChildInheritedAttributes = function(attributes, display, level, prime) {
    const before = {attributes, display, level, prime};
    const result = original.call(this, attributes, display, level, prime);
    fakeObservations.push({before,
      parentExplicit:this.attributes.getAllAttributes(),
      parentInherited:this.attributes.getAllInherited(),
      fake:[this.open,this.close,...this.separators].filter(Boolean).map(n => ({
        kind:n.kind, text:n.getText(), explicit:n.attributes.getAllAttributes(),
        inherited:n.attributes.getAllInherited(), properties:n.getAllProperties()
      }))});
    return result;
  };
  function build(s) {
    if (s.kind === 'text') return factory.create('text').setText(s.text);
    return factory.create(s.kind,s.attributes || {},(s.children || []).map(build));
  }
  const root = build(request.spec);
  root.setInheritedAttributes({},request.display,0,false);
  root.setTeXclass(null);
  const typeset = html.outputJax.typeset;
  html.outputJax.typeset = function(math, doc) {math.root=root;return typeset.call(this,math,doc)};
  const svg = adaptor.innerHTML(html.convert('x',{display:request.display,em:16,ex:8}));
  return {svg,fakeObservations};
})()`);
if (process.argv[3] === '--batch') {
  const requests = JSON.parse(fs.readFileSync(0,'utf8'));
  const output = requests.map(request => {
    const context = vm.createContext({console,request});
    for (const script of scripts) script.runInContext(context);
    return observe.runInContext(context);
  });
  process.stdout.write(JSON.stringify(output));
} else {
  const file = path.join(__dirname,'synthetic_operator_factory_mathjax_3_2_2.json');
  const fixture = JSON.parse(fs.readFileSync(file,'utf8'));
  if (fixture.mathjaxGitCommit !== 'ad8f5c21cb810236551da8c6512ba733e67357ee') throw Error('unbound primary revision');
  for (let start=0;start<fixture.cases.length;start+=24) {
    const batch = fixture.cases.slice(start,start+24);
    const run = spawnSync(process.execPath,['--jitless',__filename,assets,'--batch'],{
      input:JSON.stringify(batch),encoding:'utf8',maxBuffer:32*1024*1024
    });
    if (run.status !== 0) throw Error(run.stderr || String(run.error));
    const results = JSON.parse(run.stdout);
    if (results.length !== batch.length) throw Error('incomplete original response');
    results.forEach((result,i) => {
      const c = batch[i];
      c.original = {svg:result.svg};
      c.originalSHA256 = sha(result.svg);
      if (c.fakeObservations) c.fakeObservations = result.fakeObservations;
      // The historical CSS partition and complete baseline bindings are
      // retained. This generator never chooses or rewrites that partition.
    });
  }
  fs.writeFileSync(file,JSON.stringify(fixture,null,2)+'\n');
}
