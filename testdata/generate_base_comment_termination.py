#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Capture complete unchanged original LF-only Base comment references."""
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

requests = []
seen = set()

def add(name, tex):
    for display in [False, True]:
        if (tex, display) in seen:
            continue
        seen.add((tex, display))
        requests.append({'name': name + ('-display' if display else '-inline'), 'tex': tex, 'display': display})

for ending_name, ending in [('lf', '\n'), ('cr', '\r'), ('crlf', '\r\n'), ('ls', '\u2028'), ('ps', '\u2029')]:
    for hidden in ['+z', r'+\frac{a}{b}', r'\unknown', '{}', '{}_i^j', r'\sum_{i=1}^n i']:
        body = 'a%comment' + ending + hidden + '\nb'
        for context_name, wrap in [
            ('plain', lambda s: s),
            ('group', lambda s: '{' + s + '}'),
            ('boxed', lambda s: r'\boxed{' + s + '}'),
            ('sup', lambda s: 'Q^{' + s + '}'),
            ('matrix', lambda s: r'\begin{matrix}' + s + r'&Z\end{matrix}'),
        ]:
            add(ending_name + '-' + str(len(requests)) + '-' + context_name, wrap(body))
add('d2-comment-witness', r'A+B% hidden' + '\r' + r'+\frac{1}{0}' + '\n' + '=C')
assert len(requests) == 302
wire = "".join(json.dumps({"tex": c["tex"], "options": {"Display": c["display"]}}, ensure_ascii=True) + "\n" for c in requests)
process = subprocess.run([node, "--jitless", str(directory / "differential/oracle.mjs"), "--asset-dir", str(assets)],
                         input=wire, text=True, capture_output=True, check=True)
responses = [json.loads(line) for line in process.stdout.split("\n") if line]
assert len(responses) == len(requests)
cases = []
for request, response in zip(requests, responses):
    assert set(response) == {"svg"}, (request, response)
    cases.append({"name": request["name"], "tex": request["tex"], "display": request["display"], "original": response})
metadata = {"mathjaxGitCommit": "ad8f5c21cb810236551da8c6512ba733e67357ee", "primaryAssets": hashes,
            "scope": "Complete unchanged original APIs; fresh frozen runtime per case; all rendered errors retained."}
payload = (json.dumps({**metadata, "cases": cases}, ensure_ascii=True, indent=2) + "\n").encode()
with open(directory / "base_comment_termination_mathjax_3_2_2.json.gz", "wb") as output:
    with gzip.GzipFile(fileobj=output, mode="wb", mtime=0, filename="") as compressed:
        compressed.write(payload)
print("Original SVGs:", len(cases), "Valid:", sum("data-mjx-error" not in c["original"]["svg"] for c in cases),
      "Rendered errors:", sum("data-mjx-error" in c["original"]["svg"] for c in cases))
