// Read-only D077 observations of the unmodified, pinned MathJax 3.2.2 bundle.
const fs = require('node:fs'), vm = require('node:vm'), path = require('node:path'), crypto = require('node:crypto');
const base = process.argv[2], output = process.argv[3];
if (!base || !output) throw Error('usage: node registered-dimensions.cjs PINNED_ASSETS OUTPUT');
const sha = b => crypto.createHash('sha256').update(b).digest('hex');
const hashes = {
  'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
const context = vm.createContext({console});
for (const [name, hash] of Object.entries(hashes)) {
  const bytes = fs.readFileSync(path.join(base,name));
  if (sha(bytes) !== hash) throw Error('Unpinned '+name);
  vm.runInContext(bytes.toString(), context, {filename:name});
}
// Deliberately finite: rounding, preserved spellings, parser/cursor boundaries,
// invalid syntax, and non-mu controls. No candidate parser replaces the oracle.
context.inputs = [
  '8mu','6mu','18mu','-8mu','+6mu','0mu','-0mu','00.000mu',
  '.010799mu','.0108mu','-.010799mu','-.0108mu','.027mu','.045mu','.063mu',
  '1,5mu',',5mu','8.mu','8 mu',' 8mu ','8mu  x','8mu\tx','8mu\nx','8mu𝑥',
  '8mumu','8mu%comment','8MU','1e2mu','++8mu','8','mu','.mu','8qu','',
  '1pc','+1.250em','1,5pt','8px','8ex','8mm','8cm','8in','\ufeff8mu','\u00858mu',
];
context.getInputs = [
  '8mu x','6mu x','18mu x','-8mu x','+6mu x','0mu x',
  '.010799mu x','.0108mu x','.027mu x','1,5mu x',',5mu x','8.mu x',
  ' 8mu  x','8mu\tx','8mu\nx','8mu𝑥','8mumu','8mu%comment',
  '{8mu}x','{ 8 mu }x','{8mu x}z','{1,5mu}x','{8mu  }x','{}x','{8mu',
  '8MU x','1e2mu x','++8mu x','8x','mu x','.mu x','8qu x','',
  '1pc x','+1.250em x','{1,5pt}x','\ufeff8mu x','\u00858mu x',
];
context.accepted = JSON.parse(fs.readFileSync(path.join(__dirname,'mu_dimension_accepted.json'),'utf8')).rows;
const data = vm.runInContext(`(() => {
  const util = MathJax._.input.tex.ParseUtil.default;
  const proto = MathJax._.input.tex.TexParser.default.prototype;
  const methods = MathJax._.input.tex.base.BaseMethods.default;
  const nf = html.inputJax[0].parseOptions.nodeFactory;
  const error = e => ({id:e.id ?? null, message:e.message, name:e.name ?? null});
  const matching = inputs.flatMap(source => [false,true].map(rest => ({source,rest,result:util.matchDimen(source,rest)})));
  function parser(source) {
    const p = Object.create(proto);
    p.string = source; p.i = 0; p.currentCS = '\\\\kern';
    p.configuration = html.inputJax[0].parseOptions;
    return p;
  }
  const dimensions = getInputs.map(source => {
    const p = parser(source); let value = null, failure = null;
    try { value = proto.GetDimen.call(p,'\\\\kern'); } catch(e) { failure = error(e); }
    return {source,value,error:failure,cursorUTF16:p.i,cursorBytesUTF8:unescape(encodeURIComponent(source.slice(0,p.i))).length,remaining:source.slice(p.i),sourceUnchanged:p.string===source};
  });
  const hskip = ['8mu x','6mu x','18mu x','{8mu}x','8MU x'].map(source => {
    const p = parser(source); const made = [], pushed = []; let failure=null;
    p.create = (...args) => { const node=nf.create(...args); made.push(node); return node; };
    p.Push = n => pushed.push(n);
    try { methods.Hskip(p,'\\\\kern'); } catch(e) { failure=error(e); }
    return {source,error:failure,cursorUTF16:p.i,remaining:source.slice(p.i),created:made.map(n=>({kind:n.kind,attributes:n.attributes.getAllAttributes(),properties:n.getAllProperties()})),pushedIdentity:pushed.length===made.length&&pushed.every((n,i)=>n===made[i]),pushCount:pushed.length};
  });
  // Same renderer's length parser, bypassing TeX matchDimen intentionally.
  const length2em = MathJax._.util.lengths.length2em;
  const directMathMLLengths = ['8mu','6mu','18mu','-8mu','.010799mu','.0108mu','1.250em'].map(source=>({source,em:length2em(source)}));
  const normalization = matching.filter(r=>!r.rest).map(r=>({source:r.source,primary:r.result,expected:r.result[0]!==null&&r.source.trim().endsWith('mu')?r.result[0]+r.result[1]:r.source}));
  const extraction = accepted.map(r=>{const m=util.matchDimen(r.value);return {...r,primaryFullReturnedValue:m,expectedValue:r.error===null&&m[0]!==null&&r.value.trim().endsWith('mu')?m[0]+m[1]:r.value};});
  const extremes=['18000000000000000000000mu','18000000000000000000000000000000mu','1'+'0'.repeat(310)+'mu'].map(source=>({source,expected:util.matchDimen(source).slice(0,2).join('')}));
  return {normalization,extraction,extremes,matching,dimensions,hskip,directMathMLLengths,methods:{matchDimen:util.matchDimen.toString(),Em:util.Em.toString(),GetDimen:proto.GetDimen.toString(),GetNext:proto.GetNext.toString(),GetArgument:proto.GetArgument.toString(),nextIsSpace:proto.nextIsSpace.toString(),getCodePoint:proto.getCodePoint.toString(),Hskip:methods.Hskip.toString(),length2em:length2em.toString()}};
})()`, context);
fs.writeFileSync(output,JSON.stringify({assetsSHA256:hashes,methodIdentity:'Functions called from the actual registered unmodified bundle; synthetic parser only supplies source/cursor/configuration and observes Hskip pushes',...data},null,2)+'\n');
console.log(JSON.stringify({matching:data.matching.length,dimensions:data.dimensions.length,hskip:data.hskip.length,directMathMLLengths:data.directMathMLLengths.length}));
