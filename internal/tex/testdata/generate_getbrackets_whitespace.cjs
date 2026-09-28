// SPDX-License-Identifier: Apache-2.0
// Observe the frozen original TexParser.GetBrackets/GetNext methods directly.
// Usage: node --jitless generate_getbrackets_whitespace.cjs PINNED_ASSETS
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const assets = process.argv[2];
if (!assets) throw new Error('expected pinned D2 MathJax assets');
const hashes = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
const context = vm.createContext({console});
for (const [file, hash] of Object.entries(hashes)) {
  const data = fs.readFileSync(path.join(assets, file));
  if (crypto.createHash('sha256').update(data).digest('hex') !== hash) throw new Error(`unverified ${file}`);
  vm.runInContext(data.toString(), context, {filename: file});
}
const spaces = [9, 10, 11, 12, 13, 32, 0x85, 0xa0, 0x1680,
  ...Array.from({length: 11}, (_, i) => 0x2000 + i), 0x2028, 0x2029, 0x202f,
  0x205f, 0x3000, 0xfeff, 0x200b, 0x180e, 0x2060];
const sources = [];
for (const point of spaces) {
  for (const tail of ['[n]x', '{x}', '']) {
    sources.push({name: point.toString(16) + '-' + tail, source: String.fromCodePoint(point) + tail});
  }
}
for (const source of ['[n]x', '[]x', '[a{]}b]x', '[\\]]x', '[\\{]x', '[🙂]x',
  '[\uFEFF\u0085]x', '\uFEFF \t\u2003[n]x', '\uFEFF\u0085[n]x',
  '\u0085\uFEFF[n]x', '%comment\n[n]x', '\\relax[n]x']) {
  sources.push({name: 'content-' + sources.length, source});
}
context.requests = sources.flatMap(item => [null, 'fallback'].map(defaultValue => ({...item, defaultValue})));
const result = vm.runInContext(`(() => {
  const proto = MathJax._.input.tex.TexParser.default.prototype;
  const records = requests.map(request => {
    const parser = Object.create(proto);
    parser.string = request.source;
    parser.i = 0;
    parser.currentCS = '\\\\sample';
    const value = proto.GetBrackets.call(parser, '\\\\sample', request.defaultValue === null ? undefined : request.defaultValue);
    return {...request, value: value === undefined ? null : value,
      present: parser.i > 0 && parser.string.charAt(parser.i - 1) === ']',
      cursorBytes: unescape(encodeURIComponent(parser.string.slice(0, parser.i))).length,
      remaining: parser.string.slice(parser.i)};
  });
  return {methods: {GetBrackets: proto.GetBrackets.toString(), GetNext: proto.GetNext.toString(), nextIsSpace: proto.nextIsSpace.toString()}, cases: records};
})()`, context);
fs.writeFileSync(path.join(__dirname, 'getbrackets_whitespace_mathjax_3_2_2.json'), JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee', assetsSHA256: hashes, ...result
}, null, 2) + '\n');
console.log(result.cases.length);
