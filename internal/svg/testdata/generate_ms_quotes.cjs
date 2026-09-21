// SPDX-License-Identifier: Apache-2.0
// Pure MathML through unmodified pinned CommonMs/SVGms; no mmlToken parser.
const fs = require("node:fs"),
  vm = require("node:vm"),
  path = require("node:path"),
  crypto = require("node:crypto");
const base = process.argv[2],
  evidence = process.argv[3];
if (!base)
  throw Error(
    "usage: node generate_ms_quotes.cjs /path/to/pinned/d2latex [evidence-directory]"
  );
if (evidence) fs.mkdirSync(evidence, { recursive: true });
const hash = (b) => crypto.createHash("sha256").update(b).digest("hex");
const hashes = {
  "polyfills.js":
    "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01",
  "mathjax.js":
    "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869",
  "setup.js":
    "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881",
};
const assets = Object.entries(hashes).map(([f, h]) => {
  const b = fs.readFileSync(path.join(base, f));
  if (hash(b) !== h) throw Error("Unpinned " + f);
  return [f, b];
});
const t = (text) => ({ kind: "text", text });
const n = (kind, attributes = {}, children = [], inherited = {}) => ({
  kind,
  attributes,
  children,
  inherited,
});
const ms = (attrs = {}, text = "x", inherited = {}) =>
  n("ms", attrs, [t(text)], inherited);
const cases = [
  ["default", ms()],
  ["explicit-angle", ms({ lquote: "<", rquote: ">" })],
  ["explicit-straight", ms({ lquote: '"', rquote: '"' })],
  ["explicit-left-only", ms({ lquote: '"' })],
  ["explicit-right-only", ms({ rquote: '"' })],
  ["empty-both", ms({ lquote: "", rquote: "" })],
  ["empty-left", ms({ lquote: "" })],
  ["empty-right", ms({ rquote: "" })],
  ["empty-text", ms({}, "")],
  ["empty-all", ms({ lquote: "", rquote: "" }, "")],
  ["monospace", ms({ mathvariant: "monospace" })],
  [
    "monospace-angle",
    ms({ mathvariant: "monospace", lquote: "<", rquote: ">" }),
  ],
  ["bold", ms({ mathvariant: "bold" }, "xy")],
  ["italic", ms({ mathvariant: "italic" }, "xy")],
  ["multi-character-quotes", ms({ lquote: "<<", rquote: ">>" }, "text")],
  ["literal-quote-text", ms({}, '"x"')],
  ["escaped-xml-text", ms({ lquote: "<&", rquote: '>"' }, "<x&y>")],
  ["literal-tex", ms({}, String.raw`\alpha {x}`)],
  ["inherited-straight", ms({}, "x", { lquote: '"', rquote: '"' })],
  ["inherited-left-only", ms({}, "x", { lquote: '"' })],
  [
    "explicit-over-inherited",
    ms({ lquote: "[" }, "x", { lquote: '"', rquote: "]" }),
  ],
  ["scaled", ms({ mathsize: "2em" }, "xy")],
  ["script", n("msub", {}, [n("mi", {}, [t("a")]), ms()])],
  [
    "fraction",
    n("mfrac", {}, [ms({ lquote: "[", rquote: "]" }), n("mn", {}, [t("2")])]),
  ],
  [
    "inherited-style",
    n("mstyle", { mathvariant: "monospace", mathcolor: "red" }, [ms()]),
  ],
  ["colored", ms({ mathcolor: "red", mathbackground: "yellow" })],
  ["multiple-text-children", n("ms", {}, [t("a"), t("b")])],
  ["adjacent", n("mrow", {}, [ms(), ms({ lquote: "[", rquote: "]" }, "y")])],
  ["mtext-control", n("mtext", {}, [t("x")])],
  ["mi-control", n("mi", {}, [t("x")])],
];
const records = [];
for (const [name, spec] of cases)
  for (const display of [false, true]) {
    const c = vm.createContext({
      console: { log() {}, warn() {}, error() {} },
    });
    for (const [f, b] of assets)
      vm.runInContext(b.toString(), c, { filename: f });
    c.request = { spec, display };
    const result = vm.runInContext(
      `(()=>{
  const factory=html.inputJax[0].mmlFactory;
  const overlays=[];
  function build(s){if(s.kind==='text')return factory.create('text').setText(s.text);const node=factory.create(s.kind,s.attributes,s.children.map(build));if(Object.keys(s.inherited).length)overlays.push([node,s.inherited]);return node}
  const root=factory.create('math',request.display?{display:'block'}:{},[build(request.spec)]);
  root.setInheritedAttributes({},request.display,0,false);
  for(const [node,attrs]of overlays)for(const [k,v]of Object.entries(attrs))node.attributes.setInherited(k,v);
  root.setTeXclass(null);
  const pairs=o=>Object.keys(o).map(name=>({name,value:o[name]}));
  const project=node=>({kind:node.kind,text:node.kind==='text'?node.getText():null,
   attributes:node.attributes?{explicit:pairs(node.attributes.getAllAttributes()),inherited:pairs(node.attributes.getAllInherited()),defaults:pairs(node.attributes.getAllDefaults()),global:pairs(node.attributes.getAllGlobals())}:{explicit:[],inherited:[],defaults:[],global:[]},
   properties:pairs(node.getAllProperties()),
   flags:{token:node.isToken,embellished:node.isEmbellished,spacelike:node.isSpacelike,linebreakContainer:node.linebreakContainer,hasNewline:node.hasNewline,inferred:node.isInferred,notParent:node.notParent,arity:node.arity===Infinity?-2:node.arity,coreIndex:node.coreIndex||0},
   texClass:Number.isFinite(node.texClass)?node.texClass:-1,prevClass:Number.isFinite(node.prevClass)?node.prevClass:-1,prevLevel:Number.isFinite(node.prevLevel)?node.prevLevel:0,
   children:node.childNodes.map(project)});
  const tree=project(root), before=JSON.stringify(tree),original=html.outputJax.typeset;
  html.outputJax.typeset=function(math,doc){math.root=root;return original.call(this,math,doc)};
  const svg=adaptor.innerHTML(html.convert('x',{display:request.display,em:16,ex:8}));
  if(JSON.stringify(project(root))!==before)throw Error('Output mutated original MathML');
  return{tree,svg};
 })()`,
      c
    );
    const id = name + "-" + (display ? "display" : "inline");
    if (evidence)
      fs.writeFileSync(path.join(evidence, id + ".svg"), result.svg);
    records.push({
      name: id,
      display,
      spec,
      tree: result.tree,
      sha256: hash(result.svg),
    });
  }
fs.writeFileSync(
  path.join(__dirname, "ms_quotes_mathjax_3_2_2.json"),
  JSON.stringify(
    {
      oracle:
        "Unmodified pinned D2 MathJax 3.2.2, direct registered MathML trees; no parser command dependency",
      assetsSHA256: hashes,
      cases: records,
    },
    null,
    2
  ) + "\n"
);
console.log(records.length + " exact same-MathML SVG cases");
