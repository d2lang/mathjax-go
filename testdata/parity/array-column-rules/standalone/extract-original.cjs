const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const {parseArgs} = require('node:util');
const {values} = parseArgs({options:{
 'asset-dir':{type:'string'},
 'output-dir':{type:'string'},
 help:{type:'boolean'},
}});
if(values.help){
 console.log('Usage: node --jitless extract-original.cjs --asset-dir PATH --output-dir PATH');
 process.exit(0);
}
if(!values['asset-dir'] || !values['output-dir']){
 throw Error('--asset-dir and --output-dir are required');
}
const here=path.resolve(values['output-dir']);
const assets=path.resolve(values['asset-dir']);
if(here===path.resolve(__dirname))throw Error('Use a separate output directory to preserve the published evidence');
fs.mkdirSync(here,{recursive:true});
const expected={
 'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
 'mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
 'setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
const hash=data=>crypto.createHash('sha256').update(data).digest('hex');
const context=vm.createContext({console});context.globalThis=context;
for(const [name,sha]of Object.entries(expected)){
 const data=fs.readFileSync(path.join(assets,name));if(hash(data)!==sha)throw Error(`unverified ${name}`);
 new vm.Script(data.toString(),{filename:name}).runInContext(context);
}
const cases=[
 {id:'outer-and-dashed',tex:String.raw`\begin{array}{|c:c|}a&c\\b&d\end{array}`},
 {id:'unmarked-gap',tex:String.raw`\begin{array}{c|cc}a&b&c\\d&e&f\end{array}`},
];
for(const c of cases){
 const container=context.html.convert(c.tex,{em:16,ex:8,display:true});
 const svg=context.adaptor.innerHTML(container);
 fs.writeFileSync(path.join(here,c.id+'-original.svg'),svg);
 c.originalSVG_SHA256=hash(svg);c.originalContainerAttributes=context.adaptor.allAttributes(container);
}
const styleNode=context.html.outputJax.styleSheet(context.html);
const styleHTML=context.adaptor.outerHTML(styleNode);
const css=context.adaptor.textContent(styleNode);
fs.writeFileSync(path.join(here,'original-mathjax-svg.css'),css);
fs.writeFileSync(path.join(here,'original-mathjax-svg-style.html'),styleHTML);
const receipt={mathjaxGitCommit:'ad8f5c21cb810236551da8c6512ba733e67357ee',assets:expected,api:'html.outputJax.styleSheet(html); adaptor.textContent(styleNode)',stylesheetSHA256:hash(css),stylesheetHTML_SHA256:hash(styleHTML),styleSheetMethodSource:String(context.html.outputJax.styleSheet),renderOptions:{em:16,ex:8,display:true},cases};
fs.writeFileSync(path.join(here,'original-css-extraction.json'),JSON.stringify(receipt,null,2)+'\n');
console.log(JSON.stringify({stylesheetBytes:Buffer.byteLength(css),stylesheetSHA256:hash(css),hasSolid:css.includes('mjx-solid'),hasDashed:css.includes('mjx-dashed'),cases:cases.map(c=>({id:c.id,attributes:c.originalContainerAttributes}))},null,2));
