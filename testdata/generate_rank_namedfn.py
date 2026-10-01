#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Capture unmodified original Physics rank NamedFn references."""
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

requests=[]
seen=set()
def add(name,tex):
 for display in [False,True]:
  key=(tex,display)
  if key in seen:continue
  seen.add(key);requests.append({'name':name+('-display' if display else '-inline'),'tex':tex,'display':display})
for tail in ['', ' x', '+Z', '-Z', '=Z', r'\times Z', r'\frac{x}{y}', r'\left(A\right)', r'(A)', r'_i^j+A', r'\limits_i^j+A', r"'_i+A", r'\not\leq A', r'\text{A}', r'\hspace{1em}Z']:
 add('tail-'+str(len(requests)),r'\rank'+tail)
styles=['',r'\bf ',r'\it ',r'\sf ',r'\mathfrak',r'\mathcal',r'\mathbf',r'\mathit',r'\mathrm',r'\mathnormal',r'\color{red}',r'\scriptstyle ',r'\displaystyle ']
for style in styles:
 for body in [r'\rank+Z',r'\rank(A)+\rank(B)',r'\rank_i^j+Z',r'\rank\limits_{i=1}^{n}A']:
  tex=style+'{'+body+'}' if style and not style.endswith(' ') else style+body
  add('style-'+str(len(requests)),tex)
for ctx in [lambda b:r'Q^{'+b+'}',lambda b:r'Q^'+b,lambda b:r'\sqrt{'+b+'}',lambda b:r'\frac{'+b+r'}{x}',lambda b:r'\text{Value: $'+b+r'$}',lambda b:r'\boxed{'+b+r'}',lambda b:r'\begin{matrix}'+b+r'&Z\end{matrix}',lambda b:r'\not '+b,lambda b:r'\dots '+b,lambda b:r"Q'"+b,lambda b:r'\sin '+b]:
 for body in [r'\rank+Z',r'\bf\rank(A)',r'\rank_i^j+Z',r'\rank\limits_{i=1}^{n}A']:
  add('context-'+str(len(requests)),ctx(body))
for name in ['rank','sinh','sin','tr','trace','Re','Im']:
 add('function-control-'+name,'\\'+name+'+Z')
 add('function-control-bold-'+name,r'\mathbf{'+'\\'+name+'+Z}')
for declaration in [r'\DeclarePairedDelimiter{\rank}{[}{]}',r'\DeclareMathOperator{\rank}{Q}',r'\DeclareMathOperator*{\rank}{Q}']:
 for invoke in [r'\rank{x}+Z',r'\rank*{x}+Z',r'\rank_{i}^{j}x']:
  add('override-'+str(len(requests)),declaration+invoke)
add('d2-witness',r'\rank+\rank+\rank+\rank = 4\rank')
assert len(requests) == 264
wire = "".join(json.dumps({"tex": c["tex"], "options": {"Display": c["display"]}}, ensure_ascii=True) + "\n" for c in requests)
process = subprocess.run([node, "--jitless", str(directory / "differential/oracle.mjs"), "--asset-dir", str(assets)],
                         input=wire, text=True, capture_output=True, check=True)
responses = [json.loads(line) for line in process.stdout.split("\n") if line]
assert len(responses) == len(requests)
cases, deferred = [], []
for request, response in zip(requests, responses):
    assert set(response) == {"svg"}, (request, response)
    case = {"name": request["name"], "tex": request["tex"], "display": request["display"], "original": response}
    # These two original observations are retained without an equality claim:
    # the pre-existing FnItem/GetForm gap adds an invisible ApplyFunction node.
    (deferred if request["tex"] == r"\rank\times Z" else cases).append(case)
assert len(cases) == 262 and len(deferred) == 2
metadata = {"mathjaxGitCommit": "ad8f5c21cb810236551da8c6512ba733e67357ee", "primaryAssets": hashes,
            "scope": "Complete unchanged original APIs; fresh frozen runtime per case; deferred original observations receive no SVG parity credit."}
payload = (json.dumps({**metadata, "cases": cases, "deferred": deferred}, ensure_ascii=True, indent=2) + "\n").encode()
with open(directory / "rank_namedfn_mathjax_3_2_2.json.gz", "wb") as output:
    with gzip.GzipFile(fileobj=output, mode="wb", mtime=0, filename="") as compressed:
        compressed.write(payload)
print("Original strict SVGs:", len(cases), "Valid:", sum("data-mjx-error" not in c["original"]["svg"] for c in cases),
      "Rendered errors:", sum("data-mjx-error" in c["original"]["svg"] for c in cases), "Deferred valid originals:", len(deferred))
