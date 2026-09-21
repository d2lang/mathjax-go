// SPDX-License-Identifier: Apache-2.0
// Regenerate using the unmodified, hash-pinned D2 v0.8.1 MathJax3.2.2 assets.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2], evidence = process.argv[3];
if (!base) throw Error("usage: node generate_unicode_fallback.cjs /path/to/d2latex [svg-output-directory]");
const hash = (value) => crypto.createHash("sha256").update(value).digest("hex");
const hashes = {
  "polyfills.js": "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01",
  "mathjax.js": "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869",
  "setup.js": "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881"
};
const assets = Object.entries(hashes).map(([name, digest]) => {
  const bytes = fs.readFileSync(path.join(base, name));
  if (hash(bytes) !== digest) throw Error("Unpinned asset: " + name);
  return [name, bytes];
});
const inputs = [
  [
    "double-struck-k",
    "\\mathbb{k}"
  ],
  [
    "raw-bold-alpha",
    "\\mathbf{α}"
  ],
  [
    "fixed-Bbbk",
    "\\Bbbk"
  ],
  [
    "mixed-raw",
    "\\mathbf{α+Γ}"
  ],
  [
    "double-struck-R-control",
    "\\mathbb{R}"
  ],
  [
    "bold-gamma-control",
    "\\mathbf{\\Gamma}"
  ],
  [
    "ordinary-control",
    "x+1"
  ],
  [
    "cjk-control",
    "\\text{漢}"
  ],
  [
    "unmapped-symbol-control",
    "\\text{☃}"
  ],
  [
    "normal",
    "\\mmlToken{mi}[mathvariant=normal]{ABhkαΓϕ19!?}"
  ],
  [
    "bold",
    "\\mmlToken{mi}[mathvariant=bold]{ABhkαΓϕ19!?}"
  ],
  [
    "italic",
    "\\mmlToken{mi}[mathvariant=italic]{ABhkαΓϕ19!?}"
  ],
  [
    "bold-italic",
    "\\mmlToken{mi}[mathvariant=bold-italic]{ABhkαΓϕ19!?}"
  ],
  [
    "double-struck",
    "\\mmlToken{mi}[mathvariant=double-struck]{ABhkαΓϕ19!?}"
  ],
  [
    "fraktur",
    "\\mmlToken{mi}[mathvariant=fraktur]{ABhkαΓϕ19!?}"
  ],
  [
    "bold-fraktur",
    "\\mmlToken{mi}[mathvariant=bold-fraktur]{ABhkαΓϕ19!?}"
  ],
  [
    "script",
    "\\mmlToken{mi}[mathvariant=script]{ABhkαΓϕ19!?}"
  ],
  [
    "bold-script",
    "\\mmlToken{mi}[mathvariant=bold-script]{ABhkαΓϕ19!?}"
  ],
  [
    "sans-serif",
    "\\mmlToken{mi}[mathvariant=sans-serif]{ABhkαΓϕ19!?}"
  ],
  [
    "bold-sans-serif",
    "\\mmlToken{mi}[mathvariant=bold-sans-serif]{ABhkαΓϕ19!?}"
  ],
  [
    "sans-serif-italic",
    "\\mmlToken{mi}[mathvariant=sans-serif-italic]{ABhkαΓϕ19!?}"
  ],
  [
    "sans-serif-bold-italic",
    "\\mmlToken{mi}[mathvariant=sans-serif-bold-italic]{ABhkαΓϕ19!?}"
  ],
  [
    "monospace",
    "\\mmlToken{mi}[mathvariant=monospace]{ABhkαΓϕ19!?}"
  ],
  [
    "literal-smp-fallback",
    "\\mmlToken{mi}[mathvariant=double-struck]{𝕜}"
  ],
  [
    "non-smp-supplementary",
    "\\mmlToken{mi}[mathvariant=double-struck]{𝄞}"
  ],
  [
    "smp-upper-bound",
    "\\mmlToken{mi}[mathvariant=double-struck]{𝟿}"
  ],
  [
    "after-smp-bound",
    "\\mmlToken{mi}[mathvariant=double-struck]{𝠀}"
  ]
];
if (evidence) fs.mkdirSync(evidence, {recursive: true});
const records = [];
for (const [name, tex] of inputs) for (const display of [false, true]) {
  const item = {name: name + (display ? "-display" : "-inline"), tex, display};
  const context = vm.createContext({console: {log(){}, warn(){}, error(){}}});
  for (const [name, bytes] of assets) vm.runInContext(bytes.toString(), context, {filename: name});
  context.input = item;
  const svg = vm.runInContext("adaptor.innerHTML(html.convert(input.tex,{display:input.display,em:16,ex:8}))", context);
  if (svg.includes('data-mml-node="merror"')) throw Error(item.name + " has reference error");
  if (evidence) fs.writeFileSync(path.join(evidence, item.name + ".svg"), svg);
  records.push({...item, svgSHA256: hash(svg)});
}
fs.writeFileSync(path.join(__dirname, "unicode_fallback_mathjax_3_2_2.json"), JSON.stringify({
  oracle: "Unmodified pinned D2 v0.8.1 MathJax3.2.2; fresh VM per case/mode",
  mathjaxGitCommit: "ad8f5c21cb810236551da8c6512ba733e67357ee",
  assetsSHA256: hashes,
  cases: records
}, null, 2) + "\n");
console.log(records.length + " complete SVG references");
