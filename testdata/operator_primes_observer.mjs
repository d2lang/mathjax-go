// SPDX-License-Identifier: Apache-2.0
// Read-only observation immediately before the original SVG output pass.
import {readFileSync} from 'node:fs';
import {createContext, Script} from 'node:vm';
import {createInterface} from 'node:readline';
import {createHash} from 'node:crypto';
import path from 'node:path';

const assets = process.argv[2];
const hashes = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
const scripts = Object.entries(hashes).map(([name, hash]) => {
  const bytes = readFileSync(path.join(assets, name));
  if (createHash('sha256').update(bytes).digest('hex') !== hash) throw Error(`unverified ${name}`);
  return new Script(bytes.toString(), {filename: name});
});
for await (const line of createInterface({input: process.stdin, crlfDelay: Infinity})) {
  if (!line.trim()) continue;
  const request = JSON.parse(line);
  const context = createContext({console, request});
  context.globalThis = context;
  for (const script of scripts) script.runInContext(context);
  const result = new Script(`(() => {
    const operators = [];
    const typeset = html.outputJax.typeset;
    html.outputJax.typeset = function(math, document) {
      const visit = node => {
        if (node.kind === 'mo') {
          const all = node.getAllProperties();
          const properties = {};
          for (const key of ['pseudoscript', 'primes']) {
            if (Object.hasOwn(all, key)) properties[key] = all[key];
          }
          const core = node.coreParent();
          const entry = {
            text: node.getText(), properties,
            rawParent: node.parent ? node.parent.kind : null,
            logicalParent: core.Parent ? core.Parent.kind : null,
            coreParent: core.kind,
          };
          if (all.pseudoscript === true) {
            const inherited = node.attributes.getAllInherited();
            entry.inheritedSpacing = {lspace: inherited.lspace, rspace: inherited.rspace};
          }
          operators.push(entry);
        }
        for (const child of node.childNodes) visit(child);
      };
      visit(math.root);
      return typeset.call(this, math, document);
    };
    const svg = adaptor.innerHTML(html.convert(request.tex, {display: request.display, em: 16, ex: 8}));
    return {svg, operators};
  })()`).runInContext(context);
  process.stdout.write(JSON.stringify(result) + '\n');
}
