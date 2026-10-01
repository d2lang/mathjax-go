#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Capture complete unchanged original Cancel option results, including runtime failures."""
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

options = [
    '', 'unknown=x', '__proto__=true', 'constructor=red', '0=red',
    'unknown=x,mathcolor=red', 'mathcolor=red,unknown=x',
    ',,mathcolor=red,,', 'mathcolor=red,', 'mathcolor={ red }',
    'mathcolor={{red}}', 'mathcolor=true', 'mathcolor=false',
    'mathcolor=\ufeffred\ufeff', 'mathcolor=\u0085red\u0085',
    'mathcolor=blue,mathcolor=red', 'mathbackground=yellow',
    'mathbackground=false', 'mathcolor=red,mathbackground=yellow',
    'color=red', 'background=yellow', 'data-padding=1em',
    'data-thickness=0.1em', 'data-arrowhead=0.2em 0.1em',
    'unknown={a,b=c}', 'unknown={{a,b}},mathcolor=red',
    'unknown={a\\,b},mathcolor=red', 'unknown={},mathcolor=red',
    'mathcolor=red,unknown={a,b}', 'unknown', 'unknown=false',
    '=red', 'false', '__proto__=false,unknown,mathcolor=red',
    'mathcolor=,mathcolor=red', 'mathcolor={}',
    'mathcolor=false,color=red', 'mathbackground=false,background=yellow',
    'mathcolor=true,color=red', 'mathbackground=true,background=yellow',
    'color=false', 'background=false',
]
contexts = [
    ('plain', lambda s: s),
    ('sup', lambda s: 'Q^{' + s + '}'),
    ('font-color', lambda s: r'\color{blue}\mathbf{' + s + '}'),
    ('text', lambda s: r'\text{A $' + s + '$ B}'),
    ('matrix', lambda s: r'\begin{matrix}' + s + r'&Z\end{matrix}'),
]
for command in ['cancel', 'bcancel', 'xcancel', 'cancelto']:
    for option in options:
        call = '\\' + command + '[' + option + ']' + ('{0}' if command == 'cancelto' else '') + r'{\frac{x}{y}}'
        for context_name, wrap in contexts:
            add(command + '-' + str(len(requests)) + '-' + context_name, wrap(call))
    for option in ['data-arrowhead=true', 'data-arrowhead={true}', 'data-arrowhead=false', 'data-arrowhead', 'data-padding=true', 'data-thickness=false']:
        add(command + '-typed-boundary-' + str(len(requests)), '\\' + command + '[' + option + ']' + ('{0}' if command == 'cancelto' else '') + '{x}')

add('d2-accepted-options-witness', r'\cancel[unknown=x]{x}+\bcancel[__proto__=true]{y}+\cancelto[unknown=x,mathcolor=red]{0}{z}')
add('d2-bom-color-witness', '\\cancel[mathcolor=\ufeffred\ufeff]{x}+\\cancel[mathcolor=false]{y}')
for attribute in ['mathcolor', 'color', 'mathbackground', 'background']:
    for value in ['false', '0', 'red']:
        add('mmltoken-string-color-' + str(len(requests)), r'\mmlToken{mi}[' + attribute + '="' + value + r'"]{x}')
assert len(requests) == 1756
wire = "".join(json.dumps({"tex": c["tex"], "options": {"Display": c["display"]}}, ensure_ascii=True) + "\n" for c in requests)
process = subprocess.run([node, "--jitless", str(directory / "differential/oracle.mjs"), "--asset-dir", str(assets)],
                         input=wire, text=True, capture_output=True, check=True)
responses = [json.loads(line) for line in process.stdout.split("\n") if line]
assert len(responses) == len(requests)
cases, runtime_cases = [], []
for request, response in zip(requests, responses):
    case = {"name": request["name"], "tex": request["tex"], "display": request["display"], "original": response}
    if set(response) == {"svg"}:
        cases.append(case)
    else:
        assert set(response) == {"error"}, (request, response)
        assert response["error"].startswith("TypeError: t.trim is not a function\n"), (request, response)
        assert "at e.getParameters (mathjax.js:4:506021)" in response["error"], (request, response)
        runtime_cases.append(case)
assert len(cases) == 1724 and len(runtime_cases) == 32
assert all("data-mjx-error" not in c["original"]["svg"] for c in cases)
metadata = {"mathjaxGitCommit": "ad8f5c21cb810236551da8c6512ba733e67357ee", "primaryAssets": hashes,
            "scope": "Complete unchanged original APIs; fresh frozen runtime per case; all original runtime failures retained separately."}
payload = (json.dumps({**metadata, "cases": cases, "runtimeCases": runtime_cases}, ensure_ascii=True, indent=2) + "\n").encode()
with open(directory / "cancel_options_mathjax_3_2_2.json.gz", "wb") as output:
    with gzip.GzipFile(fileobj=output, mode="wb", mtime=0, filename="") as compressed:
        compressed.write(payload)
print("Original SVGs:", len(cases), "Valid:", sum("data-mjx-error" not in c["original"]["svg"] for c in cases),
      "Rendered errors:", sum("data-mjx-error" in c["original"]["svg"] for c in cases))

print("Original runtime exceptions:", len(runtime_cases))
