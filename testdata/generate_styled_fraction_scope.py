#!/usr/bin/env python3
"""Regenerate full styled-fraction SVGs from D2's frozen original only."""
import gzip
import hashlib
import json
import pathlib
import subprocess
import sys

directory = pathlib.Path(__file__).resolve().parent
assets = pathlib.Path(sys.argv[1])
node = sys.argv[2]
hashes = {
    "mathjax.js": "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869",
    "polyfills.js": "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01",
    "setup.js": "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881",
}
for name, expected in hashes.items():
    assert hashlib.sha256((assets / name).read_bytes()).hexdigest() == expected, name

cases = []
def append(tex, display):
    cases.append({"name": "styled-frac-" + str(len(cases)), "tex": tex, "display": display})

for command in ["frac", "dfrac", "tfrac", "binom", "dbinom", "tbinom"]:
    for tail in [r"+\sum_{i=1}^n i", r"+\frac{a}{b}", r"^{a+b}", r"_{a+b}"]:
        for outer in ["%s", r"Q^{%s}", r"\scriptstyle %s"]:
            for display in [False, True]:
                append(outer % ("\\" + command + "{x}{y}" + tail), display)
for command in ["dfrac", "tfrac"]:
    for tex in ["\\DeclareMathOperator{\\" + command + "}{F}\\" + command + "+x",
                "\\DeclarePairedDelimiter{\\" + command + "}{(}{)}\\" + command + "{x}",
                "\\" + command + "{x}", "\\" + command + "{x}{y}_",
                "\\" + command + "{x}{y}^", "\\" + command + "{x}{y}^j^k"]:
        for display in [False, True]:
            append(tex, display)
for display in [False, True]:
    append(r"\tfrac{x}{y} + \sum_{i=1}^{n} i + Q^{\dfrac{x}{y}_{a+b}}", display)
assert len(cases) == 170
wire = "".join(json.dumps({"tex": c["tex"], "options": {"Display": c["display"]}}, ensure_ascii=True) + "\n" for c in cases)
process = subprocess.run([node, "--jitless", str(directory / "differential/oracle.mjs"), "--asset-dir", str(assets)],
                         input=wire, text=True, capture_output=True, check=True)
responses = [json.loads(line) for line in process.stdout.split("\n") if line]
assert len(responses) == len(cases)
for case, response in zip(cases, responses):
    assert set(response) == {"svg"}, case["name"]
    case["SVG"] = response["svg"]
    case["TeX"] = case.pop("tex")
fixture = {"mathjaxGitCommit": "ad8f5c21cb810236551da8c6512ba733e67357ee", "primaryAssets": hashes, "cases": cases}
payload = (json.dumps(fixture, ensure_ascii=True, indent=2) + "\n").encode()
with open(directory / "styled_fraction_scope_mathjax_3_2_2.json.gz", "wb") as output:
    with gzip.GzipFile(fileobj=output, mode="wb", mtime=0, filename="") as compressed:
        compressed.write(payload)
