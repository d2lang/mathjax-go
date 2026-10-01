#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Capture unmodified original stackbin and escaped nonbreaking-space references."""
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

contexts = [
    ('plain', lambda s: s),
    ('bin-spacing', lambda s: 'a' + s + 'b'),
    ('add-spacing', lambda s: 'a+' + s + '+b'),
    ('group', lambda s: '{' + s + '}'),
    ('sup', lambda s: 'Q^{' + s + '}'),
    ('sub', lambda s: 'Q_{' + s + '}'),
    ('scripts', lambda s: s + '_i^j'),
    ('boxed', lambda s: r'\boxed{' + s + '}'),
    ('color', lambda s: r'\color{red}' + s),
    ('bold', lambda s: r'\mathbf{' + s + '}'),
    ('text', lambda s: r'\text{A $' + s + '$ B}'),
    ('matrix', lambda s: r'\begin{matrix}' + s + r'&Z\end{matrix}'),
]
for name in ['stackbin', 'stackrel']:
    for args in [r'{n}{+}', r'{\frac{a}{b}}{=}', r'{}{x}', r'{a+b}{\sum}', r'ab']:
        for context_name, wrap in contexts:
            add(name + '-' + str(len(requests)) + '-' + context_name, wrap('\\' + name + args))
    for call in ['\\' + name, '\\' + name + '{a}', '\\' + name + '{a', '\\' + name + '{a}{', '\\' + name + '*{a}{b}']:
        add(name + '-arguments-' + str(len(requests)), call)
    for declaration in [r'\DeclarePairedDelimiter{' + '\\' + name + '}{[}{]}', r'\DeclareMathOperator{' + '\\' + name + '}{Q}']:
        add(name + '-override-' + str(len(requests)), declaration + '\\' + name + '{x}')

for call_name, call in [('escaped-nbsp', '\\\u00a0'), ('space', r'\space'), ('tilde', '~'), ('raw-nbsp', '\u00a0'), ('nobreakspace', r'\nobreakspace')]:
    for body in [call, 'A' + call + 'B', call * 3, 'A+' + call + 'B', call + ' x', call + '\n' + 'x']:
        for context_name, wrap in contexts:
            add(call_name + '-' + str(len(requests)) + '-' + context_name, wrap(body))

add('d2-stackbin-witness', r'A\stackbin{k}{+}B = C')
add('d2-escaped-nbsp-witness', 'A\\\u00a0B\\\u00a0C = D')
assert len(requests) == 992
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
with open(directory / "base_command_registration_mathjax_3_2_2.json.gz", "wb") as output:
    with gzip.GzipFile(fileobj=output, mode="wb", mtime=0, filename="") as compressed:
        compressed.write(payload)
print("Original SVGs:", len(cases), "Valid:", sum("data-mjx-error" not in c["original"]["svg"] for c in cases),
      "Rendered errors:", sum("data-mjx-error" in c["original"]["svg"] for c in cases))
