#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Capture fresh frozen originals for the exact Mathtools bigtimes macro."""
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
for name, digest in hashes.items():
    assert hashlib.sha256((assets / name).read_bytes()).hexdigest() == digest

body = r'\mathop{\Large\kern-.1em\boldsymbol{\times}\kern-.1em}'
registry_script = r'''
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
const dir=process.argv[1],context=vm.createContext({console});context.globalThis=context;
for(const name of ['polyfills.js','mathjax.js','setup.js'])new vm.Script(fs.readFileSync(path.join(dir,name),'utf8'),{filename:name}).runInContext(context);
const handler=context.html.inputJax[0].parseOptions.handlers.get('macro'),symbol=handler.lookup('bigtimes');
if(!symbol)throw Error('Missing original registration');
const maps=Array.from(handler._configuration).filter(x=>x.item.contains('bigtimes')).map(x=>({map:x.item.name,priority:x.priority}));
console.log(JSON.stringify({name:symbol.symbol,methodIsBaseMacro:symbol.func===context.MathJax._.input.tex.base.BaseMethods.default.Macro,
  methodArguments:symbol.args,maps,functionSource:symbol.func.toString()}));
'''
registry_process = subprocess.run([node, "--jitless", "-e", registry_script, str(assets)],
                                  text=True, capture_output=True, check=True)
registration = json.loads(registry_process.stdout)
assert registration["name"] == "bigtimes" and registration["methodIsBaseMacro"]
assert registration["methodArguments"] == [body]
assert any(m["map"] == "mathtools-macros" for m in registration["maps"])
contexts = [
    ("plain", lambda s: s),
    ("operands", lambda s: "a" + s + " b"),
    ("sum", lambda s: "a+" + s + "+b"),
    ("scripts", lambda s: s + "_i^j x_i"),
    ("limits", lambda s: s + r'\limits_{i=1}^{n} x_i'),
    ("nolimits", lambda s: s + r'\nolimits_{i=1}^{n} x_i'),
    ("superscript", lambda s: "Q^{" + s + "_i^j}"),
    ("subscript", lambda s: "Q_{" + s + "_i^j}"),
    ("scriptstyle", lambda s: r'{\scriptstyle ' + s + "_i^j}"),
    ("fraction", lambda s: r'\frac{' + s + "_i^j x_i}{1+x}"),
    ("color", lambda s: r'\textcolor{red}{a' + s + " b}"),
    ("mathfont", lambda s: r'\mathbfit{' + s + "}"),
    ("fences", lambda s: r'\left(' + s + r'_i^j\right)'),
    ("matrix", lambda s: r'\begin{matrix}' + s + r'&x\\y&' + s + r'\end{matrix}'),
]
requests = []
seen = set()

def add(name, tex, group="valid", expansion_control=None):
    for display in [False, True]:
        key = (tex, display)
        if key in seen:
            continue
        seen.add(key)
        request = {"name": name + ("-display" if display else "-inline"),
                   "tex": tex, "display": display, "group": group}
        if expansion_control is not None:
            request["expansionControl"] = expansion_control
        requests.append(request)

for alias in ["DeclarePairedDelimiterX", "DeclarePairedDelimitersX"]:
    for formatter_name, formatter in [("bold", r'\mathbf{#1}'), ("plain", "#1"), ("bold-italic", r'\mathbfit{#1}')]:
        declaration = "\\" + alias + r'{\boldsymbol}[1]{}{}{' + formatter + "}"
        for context_name, wrap in contexts:
            add(alias + "-" + formatter_name + "-" + context_name, declaration + wrap(r'\bigtimes'))
            add(alias + "-" + formatter_name + "-" + context_name + "-expansion-control", declaration + wrap(body))
        add(alias + "-" + formatter_name + "-formatter-control", declaration + r'\boldsymbol{\times}')

# Runtime definitions must retain precedence over the newly installed macro.
for alias in ["DeclarePairedDelimiterX", "DeclarePairedDelimitersX"]:
    add(alias + "-bigtimes-override", "\\" + alias + r'{\bigtimes}[1]{}{}{\mathbf{#1}}\bigtimes{x}')
add("operator-control", r'a+\sum_{i=1}^{n}x_i+b')
add("ordinary-times-control", r'a\times b')
add("d2-witness", r'\DeclarePairedDelimiterX{\boldsymbol}[1]{}{}{\mathbf{#1}}{\Huge\bigtimes_{i=1}^{n} x_i = x_1\times x_2\times\cdots\times x_n}')

# The frozen original lacks boldsymbol, while Go intentionally provides it.
# These complete original diagnostic objects are preserved, never SVG credit.
for context_name, wrap in contexts[:5]:
    add("bare-bigtimes-" + context_name, wrap(r'\bigtimes'), "original-diagnostic-go-extension-boundary", wrap(body))

wire = "".join(json.dumps({"tex": r["tex"], "options": {"Display": r["display"]}}, ensure_ascii=True) + "\n" for r in requests)
process = subprocess.run([node, "--jitless", str(directory / "differential/oracle.mjs"), "--asset-dir", str(assets)],
                         input=wire, text=True, capture_output=True, check=True)
responses = [json.loads(line) for line in process.stdout.split("\n") if line]
assert len(responses) == len(requests)
cases, boundaries = [], []
for request, response in zip(requests, responses):
    assert set(response) == {"svg"}, (request, response)
    diagnostic = 'data-mml-node="merror"' in response["svg"]
    if request["group"] == "valid":
        assert not diagnostic, (request, response)
        cases.append({**request, "original": response})
    else:
        assert diagnostic and 'data-mjx-error="Undefined control sequence \\boldsymbol"' in response["svg"]
        boundaries.append({**request, "original": response,
                           "qualification": "Original diagnostic caused by unavailable boldsymbol package; Go-only boldsymbol is intentionally retained. Zero original-valid/SVG parity credit."})
metadata = {"mathjaxGitCommit": "ad8f5c21cb810236551da8c6512ba733e67357ee", "primaryAssets": hashes,
            "registration": registration,
            "scope": "Complete unmodified APIs; fresh frozen runtime per render; authored active Mathtools declarations provide the boldsymbol dependency.",
            "originalValidCases": len(cases), "originalDiagnosticBoundaries": len(boundaries), "originalRuntimeFailures": 0}
for name, digest in hashes.items():
    assert hashlib.sha256((assets / name).read_bytes()).hexdigest() == digest
payload = (json.dumps({**metadata, "cases": cases, "originalDiagnosticGoExtensionBoundaries": boundaries}, ensure_ascii=True, indent=2) + "\n").encode()
with (directory / "mathtools_bigtimes_mathjax_3_2_2.json.gz").open("wb") as output:
    with gzip.GzipFile(fileobj=output, mode="wb", mtime=0, filename="") as compressed:
        compressed.write(payload)
print(json.dumps(metadata, indent=2))
