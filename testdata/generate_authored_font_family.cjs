// SPDX-License-Identifier: Apache-2.0
// Regenerate using the unmodified, hash-pinned D2 v0.8.1 MathJax3.2.2 assets.
const fs = require("node:fs");
const vm = require("node:vm");
const path = require("node:path");
const crypto = require("node:crypto");
const base = process.argv[2], evidence = process.argv[3];
if (!base) throw Error("usage: node generate_authored_font_family.cjs /path/to/d2latex [svg-output-directory]");
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
const inputs = [["mi-monospace", "\\mmlToken{mi}[fontfamily=monospace]{xyz}"], ["mi-serif", "\\mmlToken{mi}[fontfamily=serif]{xyz}"], ["mi-single", "\\mmlToken{mi}[fontfamily=monospace]{x}"], ["mn-family", "\\mmlToken{mn}[fontfamily=monospace]{123}"], ["mo-family", "\\mmlToken{mo}[fontfamily=monospace]{+}"], ["mtext-family", "\\mmlToken{mtext}[fontfamily=monospace]{hello}"], ["ms-family", "\\mmlToken{ms}[fontfamily=monospace]{abc}"], ["mi-empty-family", "\\mmlToken{mi}[fontfamily=\"\"]{xyz}"], ["family-bold", "\\mmlToken{mi}[fontfamily=monospace,fontweight=bold]{xyz}"], ["family-600", "\\mmlToken{mi}[fontfamily=monospace,fontweight=600]{xyz}"], ["family-700", "\\mmlToken{mi}[fontfamily=monospace,fontweight=700]{xyz}"], ["family-italic", "\\mmlToken{mi}[fontfamily=monospace,fontstyle=italic]{xyz}"], ["family-bold-italic", "\\mmlToken{mi}[fontfamily=monospace,fontweight=bold,fontstyle=italic]{xyz}"], ["explicit-variant", "\\mmlToken{mi}[fontfamily=monospace,mathvariant=normal]{xyz}"], ["explicit-bold-variant", "\\mmlToken{mi}[fontfamily=monospace,mathvariant=bold]{xyz}"], ["family-greek", "\\mmlToken{mi}[fontfamily=monospace]{α}"], ["family-cjk", "\\mmlToken{mi}[fontfamily=monospace]{漢字}"], ["family-smp", "\\mmlToken{mi}[fontfamily=monospace]{𝕜}"], ["family-in-bold", "\\mathbf{\\mmlToken{mi}[fontfamily=monospace]{xyz}}"], ["family-script", "x^{\\mmlToken{mi}[fontfamily=monospace]{xyz}}"], ["family-delimiter", "\\mmlToken{mo}[fontfamily=monospace]{(}"], ["style-family", "\\mmlToken{mi}[style=\"font-family:monospace\"]{xyz}"], ["style-priority", "\\mmlToken{mi}[fontfamily=serif,style=\"font-family:monospace; font-weight:bold; font-style:italic\"]{xyz}"], ["style-variant", "\\mmlToken{mi}[mathvariant=bold,style=\"font-family:monospace\"]{xyz}"], ["style-padding", "\\mmlToken{mi}[style=\"font-family:monospace; padding:1px; color:red\"]{xyz}"], ["ordinary", "x+y"], ["error-control", "x^2^3"], ["largeop-family", "\\mmlToken{mo}[fontfamily=monospace,largeop=true]{∑}"], ["largeop-style-family", "\\mmlToken{mo}[style=\"font-family:monospace\",largeop=true]{∑}"], ["stretch-family", "\\mmlToken{mo}[fontfamily=monospace,minsize=3em]{(}"], ["horizontal-family", "\\mmlToken{mo}[fontfamily=monospace,stretchy=true,minsize=3em]{→}"], ["family-attribute-style-priority", "\\mmlToken{mi}[fontfamily=serif,fontweight=normal,fontstyle=normal,style=\"font-family:monospace; font-weight:bold; font-style:italic\"]{xyz}"], ["family-empty-css", "\\mmlToken{mi}[fontfamily=\"\",style=\"font-family:monospace\"]{xyz}"], ["family-explicit-empty-variant", "\\mmlToken{mi}[fontfamily=monospace,mathvariant=\"\"]{xyz}"], ["family-weight-huge", "\\mmlToken{mi}[fontfamily=monospace,fontweight=999999999999999999999999999999999999999999999999999999]{xyz}"], ["family-list", "\\mmlToken{mi}[fontfamily=\"Arial, serif\"]{xyz}"], ["family-mathsize", "\\mmlToken{mi}[fontfamily=monospace,mathsize=150%]{xyz}"], ["family-ms-custom", "\\mmlToken{ms}[fontfamily=monospace,lquote=\"<\",rquote=\">\"]{xyz}"], ["family-scripts-control", "\\mmlToken{mi}[fontfamily=monospace]{xyz}_i^2"]];
if (evidence) fs.mkdirSync(evidence, {recursive: true});
const records = [];
for (const [name, tex] of inputs) for (const display of [false, true]) {
  const item = {name: name + (display ? "-display" : "-inline"), tex, display};
  const context = vm.createContext({console: {log(){}, warn(){}, error(){}}});
  for (const [name, bytes] of assets) vm.runInContext(bytes.toString(), context, {filename: name});
  context.input = item;
  const svg = vm.runInContext("adaptor.innerHTML(html.convert(input.tex,{display:input.display,em:16,ex:8}))", context);

  if (evidence) fs.writeFileSync(path.join(evidence, item.name + ".svg"), svg);
  records.push({...item, svgSHA256: hash(svg)});
}
fs.writeFileSync(path.join(__dirname, "authored_font_family_mathjax_3_2_2.json"), JSON.stringify({
  oracle: "Unmodified pinned D2 v0.8.1 MathJax3.2.2; fresh VM per case/mode",
  mathjaxGitCommit: "ad8f5c21cb810236551da8c6512ba733e67357ee",
  assetsSHA256: hashes,
  cases: records
}, null, 2) + "\n");
console.log(records.length + " complete SVG references");
