// SPDX-License-Identifier: Apache-2.0
// Actual unchanged primary handlers and SubsupItem checkItem; no model replaces them.
const fs = require('node:fs'), vm = require('node:vm'), path = require('node:path'), crypto = require('node:crypto');
const base = process.argv[2], output = process.argv[3];
if (!base || !output) throw Error('usage: node generate_script_argument_items.cjs pinned-assets output.json');
const hashes = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'
};
const context = vm.createContext({console});
for (const [name, hash] of Object.entries(hashes)) {
  const data = fs.readFileSync(path.join(base, name));
  if (crypto.createHash('sha256').update(data).digest('hex') !== hash) throw Error('Unpinned ' + name);
  vm.runInContext(data.toString(), context, {filename: name});
}
const result = vm.runInContext(fs.readFileSync(path.join(__dirname, 'script_argument_items_body.js'), 'utf8'), context);
fs.writeFileSync(output, JSON.stringify({mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee', assetsSHA256: hashes, ...result}, null, 2) + '\n');
console.log(result.rows.length + ' actual primary handler/item records');
