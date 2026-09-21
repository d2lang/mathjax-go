// SPDX-License-Identifier: Apache-2.0
// Test-only overlays of compiled MathML. No mmlToken parser support is assumed.
const fs = require("node:fs"),
  vm = require("node:vm"),
  crypto = require("node:crypto"),
  path = require("node:path");
const base = process.argv[2];
if (!base)
  throw Error(
    "usage: node generate_explicit_spacing.cjs /path/to/pinned/d2latex [evidence-directory]"
  );
const evidence = process.argv[3];
if (evidence) fs.mkdirSync(evidence, { recursive: true });
const hash = (x) => crypto.createHash("sha256").update(x).digest("hex");
const assets = Object.fromEntries(
  ["polyfills.js", "mathjax.js", "setup.js"].map((f) => [
    f,
    fs.readFileSync(path.join(base, f)),
  ])
);
const hashes = {
  "polyfills.js":
    "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01",
  "mathjax.js":
    "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869",
  "setup.js":
    "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881",
};
for (const [f, b] of Object.entries(assets))
  if (hash(b) !== hashes[f]) throw Error(f + " not pinned");
const a = (index, explicit = {}, inherited = {}, removeClass = false) => ({
  index,
  explicit,
  inherited,
  removeClass,
});
const cases = [
  ["empty-before-sum", String.raw`\sum_{i}^{n}`, [], true],
  [
    "empty-before-zero-sum",
    String.raw`\sum_{i}^{n}`,
    [a(1, { lspace: "0", rspace: "0" })],
    true,
  ],
  ["leading-explicit", "+x", [a(0, { lspace: "1em", rspace: "0em" })]],
  ["internal-zero", "x+y", [a(0, { lspace: "0em", rspace: "0em" })]],
  ["isolated", "+", [a(0, { lspace: "1em", rspace: "2em" })]],
  ["left-only-default-right", "+x", [a(0, { lspace: "1em" })]],
  ["right-only-default-left", "x+y", [a(0, { rspace: "1em" })]],
  ["inherited", "x+y", [a(0, {}, { lspace: ".5em", rspace: ".25em" })]],
  [
    "explicit-over-inherited",
    "x+y",
    [a(0, { lspace: "0", rspace: ".1em" }, { lspace: "1em", rspace: "2em" })],
  ],
  ["negative-clamped", "x+y", [a(0, { lspace: "-.5em", rspace: "-1em" })]],
  [
    "adjacent-max",
    "x++y",
    [
      a(0, { lspace: ".2em", rspace: ".6em" }),
      a(1, { lspace: ".9em", rspace: ".1em" }),
    ],
  ],
  [
    "adjacent-covered",
    "x++y",
    [a(0, { rspace: "1em" }), a(1, { lspace: ".2em" })],
  ],
  [
    "adjacent-embellished",
    String.raw`x+_{i}+^{j}y`,
    [a(0, { rspace: ".4em" }), a(1, { lspace: ".7em" })],
  ],
  ["script-default-large", String.raw`x^{y+z}`, [a(0, { lspace: "0" })]],
  ["script-default-small", String.raw`x^{+y}`, [a(0, { rspace: "0" })]],
  [
    "script-explicit",
    String.raw`x^{y+z}`,
    [a(0, { lspace: "1em", rspace: ".5em" })],
  ],
  ["script-secondlevel", String.raw`x^{y^{a+b}}`, [a(0, { lspace: "0" })]],
  [
    "embellished-subscript",
    String.raw`x+_{i}y`,
    [a(0, { lspace: ".4em", rspace: ".6em" })],
  ],
  [
    "embellished-single-row",
    String.raw`x\mathord{+}y`,
    [a(0, { lspace: ".4em", rspace: ".6em" })],
  ],
  [
    "single-child-fraction",
    String.raw`\frac{+}{x}`,
    [a(0, { lspace: "1em", rspace: "2em" })],
  ],
  [
    "multi-child-fraction",
    String.raw`\frac{+x}{y}`,
    [a(0, { lspace: "1em", rspace: ".2em" })],
  ],
  ["form-only-control", "x+y", [a(0, { form: "prefix" })]],
  ["no-spacing-control", "x+y", []],
  [
    "remove-class-explicit",
    "x+y",
    [a(0, { lspace: "0em", rspace: "0em" }, {}, true)],
  ],
];
const records = [];
for (const [name, tex, overlays, prependEmpty = false] of cases)
  for (const display of prependEmpty ? [true] : [false, true]) {
    const c = vm.createContext({
      console: { log() {}, warn() {}, error() {} },
    });
    for (const [f, b] of Object.entries(assets))
      vm.runInContext(b.toString(), c, { filename: f });
    c.request = { tex, overlays, display, prependEmpty };
    const data = vm.runInContext(
      `(()=>{let info;const original=html.outputJax.typeset;html.outputJax.typeset=function(math,doc){if(request.prependEmpty){const row=math.root.childNodes[0];if(!row.isKind('mrow'))throw Error('Expected inferred root row');const empty=math.root.factory.create('mo',{rspace:'0'});row.childNodes.unshift(empty);empty.parent=row;empty.setInheritedAttributes({},request.display,0,false);}const nodes=[];math.root.walkTree(n=>{if(n.kind==='mo')nodes.push(n)});for(const o of request.overlays){const n=nodes[o.index];if(!n)throw Error('Missing mo');for(const [k,v]of Object.entries(o.explicit))n.attributes.set(k,v);for(const [k,v]of Object.entries(o.inherited))n.attributes.setInherited(k,v);if(o.removeClass)n.removeProperty('texClass');}math.root.setTeXclass(null);info=nodes.map(n=>({text:n.getText(),lspace:n.lspace,rspace:n.rspace,scriptlevel:n.attributes.get('scriptlevel')}));return original.call(this,math,doc)};const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));return{svg,operators:info};})()`,
      c
    );
    if (data.svg.includes('data-mml-node="merror"'))
      throw Error(name + " merror");
    const id = name + "-" + (display ? "display" : "inline");
    if (evidence) fs.writeFileSync(path.join(evidence, id + ".svg"), data.svg);
    records.push({
      name: id,
      tex,
      display,
      overlays,
      prependEmpty,
      sha256: hash(data.svg),
      operators: data.operators,
    });
  }
const out = {
  oracle:
    "Unmodified pinned D2 MathJax3.2.2, pure post-compile MML overlays before output; no mmlToken parser claim",
  assetsSHA256: hashes,
  cases: records,
};
fs.writeFileSync(
  path.join(__dirname, "explicit_spacing_mathjax_3_2_2.json"),
  JSON.stringify(out, null, 2) + "\n"
);
console.log(records.length + " exact SVG cases");
