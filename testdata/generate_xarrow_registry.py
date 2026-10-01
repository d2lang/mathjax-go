#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Capture unmodified original active xArrow registrations and padding references."""
import gzip
import hashlib
import json
import pathlib
import subprocess
import sys

directory = pathlib.Path(__file__).resolve().parent
assets, node = pathlib.Path(sys.argv[1]), sys.argv[2]
hashes = {
    "mathjax.js": "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869",
    "polyfills.js": "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01",
    "setup.js": "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881",
}
for name, expected in hashes.items():
    assert hashlib.sha256((assets / name).read_bytes()).hexdigest() == expected, name

names=['xrightarrow','xleftarrow','xleftrightarrow','xLeftarrow','xRightarrow','xLeftrightarrow','xhookleftarrow','xhookrightarrow','xmapsto','xrightharpoondown','xleftharpoondown','xrightleftharpoons','xRightleftharpoons','xLeftrightharpoons','xrightharpoonup','xleftharpoonup','xleftrightharpoons']
req=[];seen=set()
def add(name,tex):
 for d in [False,True]:
  if (tex,d) in seen:continue
  seen.add((tex,d));req.append({'name':name+('-display' if d else '-inline'),'tex':tex,'display':d})
for name in names:
 for label in ['',r'[g]',r'[\frac{c}{d}]']:
  call='\\'+name+label+r'{\frac{a}{b}}'
  for cname,ctx in [('plain',lambda s:s),('sup',lambda s:r'Q^{'+s+'}'),('group-scripts',lambda s:'{'+s+r'}_i^j'),('own-scripts',lambda s:s+r'_i^j'),('bold',lambda s:r'\mathbf{'+s+'}'),('color',lambda s:r'\color{red}'+s),('boxed',lambda s:r'\boxed{'+s+'}'),('text',lambda s:r'\text{A $'+s+'$ B}'),('matrix',lambda s:r'\begin{matrix}'+s+r'&Z\end{matrix}')]:
   add(name+'-'+str(len(req))+'-'+cname,ctx(call))
 for call in ['\\'+name+r'{f}', '\\'+name+r'[g]{}','\\'+name+'[]{}','\\'+name,'\\'+name+'[a','\\'+name+r'{}^','\\'+name+'{}_', '\\'+name+r'*{f}']:
  add(name+'-control-'+str(len(req)),call)
 for decl in [r'\DeclarePairedDelimiter{'+'\\'+name+r'}{[}{]}',r'\DeclareMathOperator{'+'\\'+name+r'}{Q}']:
  add(name+'-override-'+str(len(req)),decl+'\\'+name+r'{x}')
add('d2-alias-witness',r'A\xRightleftharpoons[k_2]{k_1}B\xLeftrightharpoons[k_4]{k_3}C')
add('d2-metrics-witness',r'A\xleftrightarrow{k}B\xrightleftharpoons{k}C')
assert len(req) == 1262
wire = "".join(json.dumps({"tex": c["tex"], "options": {"Display": c["display"]}}, ensure_ascii=True) + "\n" for c in req)
process = subprocess.run([node, "--jitless", str(directory / "differential/oracle.mjs"), "--asset-dir", str(assets)],
                         input=wire, text=True, capture_output=True, check=True)
responses = [json.loads(line) for line in process.stdout.split("\n") if line]
assert len(responses) == len(req)
cases = []
for request, response in zip(req, responses):
    assert set(response) == {"svg"}, (request, response)
    cases.append({"name": request["name"], "tex": request["tex"], "display": request["display"], "original": response})
metadata = {"mathjaxGitCommit": "ad8f5c21cb810236551da8c6512ba733e67357ee", "primaryAssets": hashes,
            "scope": "Complete unchanged original APIs; fresh frozen runtime per case; all rendered errors retained."}
payload = (json.dumps({**metadata, "cases": cases}, ensure_ascii=True, indent=2) + "\n").encode()
with open(directory / "xarrow_registry_mathjax_3_2_2.json.gz", "wb") as output:
    with gzip.GzipFile(fileobj=output, mode="wb", mtime=0, filename="") as compressed:
        compressed.write(payload)
print("Original SVGs:", len(cases), "Valid:", sum("data-mjx-error" not in c["original"]["svg"] for c in cases),
      "Rendered errors:", sum("data-mjx-error" in c["original"]["svg"] for c in cases))
