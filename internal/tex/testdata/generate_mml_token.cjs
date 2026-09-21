// SPDX-License-Identifier: Apache-2.0
// Uses unmodified pinned D2 MathJax 3.2.2; no parser or MML overlays.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_mml_token.cjs /path/to/pinned/d2latex [evidence-directory]"
  );
if (evidence) fs.mkdirSync(evidence, { recursive: true });
const hash = (x) => crypto.createHash("sha256").update(x).digest("hex");
const hashes = {
  "polyfills.js":
    "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01",
  "mathjax.js":
    "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869",
  "setup.js":
    "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881",
};
const assets = Object.entries(hashes).map(([f, sha]) => {
  const b = fs.readFileSync(path.join(base, f));
  if (hash(b) !== sha) throw Error("Unpinned " + f);
  return [f, b];
});
const token = (kind, attrs, text) =>
  String.raw`\mmlToken{${kind}}` +
  (attrs === null ? "" : "[" + attrs + "]") +
  "{" +
  text +
  "}";
const cases = [
  ["mi", token("mi", null, "x"), true],
  ["mn", token("mn", null, "123"), true],
  ["mo", token("mo", null, "+"), true],
  ["mtext", token("mtext", null, "two words"), true],
  ["ms", token("ms", null, "text")],
  ["mglyph", token("mglyph", 'alt="literal" width="1em" height="1em"', "")],
  ["literal-tex", token("mtext", null, String.raw`\alpha {x} ^2`), true],
  ["literal-xml", token("mtext", null, "<&>"), true],
  ["literal-spaces", token("mtext", null, "  a  b  "), true],
  ["empty-text", token("mi", null, ""), true],
  ["quoted", token("mi", `mathvariant='bold', mathcolor="red"`, "x"), true],
  ["unquoted", token("mi", "mathvariant=bold,mathcolor=red", "x"), true],
  [
    "boolean",
    token("mo", "stretchy=FaLsE,symmetric=TrUe,movablelimits=TRUE", "∑"),
  ],
  ["keep-default", token("mi", "mathvariant=normal", "x"), true],
  ["empty-attrs", token("mi", "id=,class=\"\",mathvariant=''", "x"), true],
  [
    "duplicate",
    token("mi", "id=first,id=second,mathvariant=bold,mathvariant=normal", "x"),
    true,
  ],
  ["duplicate-empty", token("mi", 'id=first,id=,id=""', "x"), true],
  [
    "allowlist",
    token(
      "mtext",
      'fontfamily="serif",fontsize=16px,fontweight=bold,fontstyle=italic,color=red,background=white,id=label,class=token,href="#local",style="color: blue"',
      "x"
    ),
  ],
  ["global-default", token("mi", "scriptlevel=2,displaystyle=false", "x")],
  ["token-specific", token("ms", 'lquote="<",rquote=">"', "x")],
  [
    "explicit-spaces-shape-only",
    token("mo", "lspace=1em,rspace=0em", "+") + "x",
  ],
  ["separator-no-comma", token("mi", "id=one class=two", "x"), true],
  ["single-quote-space", token("mtext", "id='two words'", "x"), true],
  ["unquoted-tab-value", token("mi", "id=one\tclass=two", "x")],
  ["unclosed-quote-is-unquoted", token("mi", "id='abc", "x")],
  ["linebreak-inside-quote", token("mi", 'id="first\nsecond"', "x")],
  ["boolean-string-space", token("mi", 'id=" true "', "x")],
  ["boolean-id", token("mi", "id=FALSE", "x")],
  ["unicode-not-boolean", token("mo", "stretchy=falſe", "(")],
  [
    "js-whitespace",
    token(
      "mi",
      '\u00a0\ufeffid\u2003=\u2028"x"\u2029,\u1680class=y\u3000',
      "x"
    ),
  ],
  ["nel-not-js-whitespace", token("mi", "\u0085id=x", "x")],
  ["case-sensitive-attribute", token("mi", "ID=x", "x")],
  ["unicode-fold-name", token("mi", "claſs=x", "x")],
  ["hyphen-name", token("mi", "data-id=x", "x")],
  ["digit-name", token("mi", "id2=x", "x")],
  ["unknown", token("mi", "unknown=x", "x")],
  ["unknown-empty", token("mi", "unknown=", "x")],
  ["wrong-token-attribute", token("mi", "stretchy=true", "x")],
  ["invalid-tail", token("mi", "id=ok, broken", "x")],
  ["trailing-double-comma", token("mi", "id=ok,,", "x")],
  ["space-in-unquoted", token("mi", "id=two words", "x")],
  ["missing-equals", token("mi", "id", "x")],
  ["mrow-not-token", token("mrow", null, "x")],
  ["mspace-ignores-text", token("mspace", null, "x"), true],
  ["text-not-token", token("text", null, "x")],
  ["xml-not-token", token("XML", null, "x")],
  ["unknown-kind", token("unknown", null, "x")],
  ["kind-case", token("MI", null, "x")],
  ["kind-whitespace", token(" mi ", null, "x")],
  ["missing-kind", String.raw`\mmlToken`],
  ["missing-text", String.raw`\mmlToken{mi}`],
  ["invalid-kind-missing-text", String.raw`\mmlToken{mrow}`],
  ["missing-bracket", String.raw`\mmlToken{mi}[id=x{x}`],
  ["extra-brace-in-bracket", String.raw`\mmlToken{mi}[id=x}]{x}`],
  ["missing-brace", String.raw`\mmlToken{mi}{x`],
  ["unbraced-text", String.raw`\mmlToken{mi}x`, true],
  ["control-sequence-text", String.raw`\mmlToken{mtext}\alpha`, true],
  ["scripted-token", String.raw`\mmlToken{mi}{x}_i`, true],
  [
    "styled-context",
    String.raw`\mathbf{\mmlToken{mi}[mathvariant=normal]{x}}`,
    true,
  ],
  [
    "nested-explicit-variant",
    String.raw`\mathbf{\mathit{\mmlToken{mi}[mathvariant=double-struck]{R}}}`,
    true,
  ],
  [
    "declaration-explicit-variant",
    String.raw`{\bf \mmlToken{mi}[mathvariant=normal]{x}}`,
    true,
  ],
  ["ordinary-font-control", String.raw`\mathbf{x}`, true],
];
const records = [];
for (const [name, tex, svgControl = false] of cases) {
  const c = vm.createContext({ console: { log() {}, warn() {}, error() {} } });
  for (const [f, b] of assets)
    vm.runInContext(b.toString(), c, { filename: f });
  c.request = { tex, svgControl };
  const result = vm.runInContext(
    `(()=>{
    let tree, error=null;const keptNodes=[];
    const nodeUtil=MathJax._.input.tex.NodeUtil.default,setProperties=nodeUtil.setProperties;
    nodeUtil.setProperties=function(n,defs){setProperties(n,defs);if(Object.hasOwn(defs,'mjx-keep-attrs'))keptNodes.push({kind:n.kind,attributes:Object.fromEntries(n.attributes.getExplicitNames().sort().map(k=>[k,n.attributes.getExplicit(k)])),movablelimits:n.getProperty('movablelimits')??null})};
    const format=html.inputJax[0].formatError;
    html.inputJax[0].formatError=function(e){error={id:e.id,message:e.message};return format.call(this,e)};
    const project=n=>({kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?Object.fromEntries(n.attributes.getExplicitNames().sort().map(k=>[k,n.attributes.getExplicit(k)])):{},children:n.childNodes.map(project)});
    const typeset=html.outputJax.typeset;
    html.outputJax.typeset=function(math,doc){tree=project(math.root);return request.svgControl?typeset.call(this,math,doc):adaptor.node('g')};
    const svg=adaptor.innerHTML(html.convert(request.tex,{display:true,em:16,ex:8}));
    return {tree,error,svg,keptNodes};
  })()`,
    c
  );
  if (evidence && svgControl)
    fs.writeFileSync(path.join(evidence, name + ".svg"), result.svg);
  records.push({
    name,
    tex,
    tree: result.tree,
    error: result.error,
    keptNodes: result.keptNodes,
    svgSHA256: svgControl ? hash(result.svg) : null,
  });
}
const result = {
  oracle:
    "Unmodified D2 v0.8.1 embedded MathJax 3.2.2, fresh VM per case; parser and output unchanged",
  mathjaxGitCommit: "ad8f5c21cb810236551da8c6512ba733e67357ee",
  assetsSHA256: hashes,
  cases: records,
};
fs.writeFileSync(
  path.join(__dirname, "mml_token_mathjax_3_2_2.json"),
  JSON.stringify(result, null, 2) + "\n"
);
console.log(
  records.length +
    " parser cases, " +
    records.filter((r) => r.svgSHA256).length +
    " complete SVG controls"
);
