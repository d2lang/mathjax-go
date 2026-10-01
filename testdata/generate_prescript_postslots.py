#!/usr/bin/env python3
"""Capture complete original APIs for authored scripts after Mathtools prescript."""
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

cases, seen = [], set()
def append(tex, display):
    key = tex, display
    if key in seen:
        return
    seen.add(key)
    cases.append({"name": "prescript-postslots-" + str(len(cases)), "tex": tex, "display": display})

for sup, sub in [("a", "b"), ("a", ""), ("", "b"), ("", ""), ("{}", "{}")]:
    for base in ["C", r"\sum", "A+B", r"\prescript{x}{y}{C}"]:
        for tail in ["", "_d", "^e", "_d^e", "^e_d", "'", "'_d", "'^e", "_d'", "^e'", "_d_d", "^e^e", "^", "_"]:
            for group in [False, True]:
                for display in [False, True]:
                    tex = "\\prescript{" + sup + "}{" + sub + "}{" + base + "}"
                    append(("{" + tex + "}" if group else tex) + tail, display)
assert len(cases) == 1120

for outer in ["%s", r"\frac{%s}{x}", r"Q^{%s}", r"\ket{%s}", r"\left(%s\right)",
              r"\begin{matrix}%s&x\\y&z\end{matrix}", r"\underbrace{%s}_{k}", r"\smash{%s}",
              r"\pmb{%s}", r"\cancel{%s}", r"\boxed{%s}", r"\color{red}%s", r"\mathbf{%s}"]:
    for display in [False, True]:
        append(outer % r"\prescript{a}{b}{C}_d^e", display)
for tex in [r"\prescript{\frac{a}{b}}{\frac{c}{d}}{X}_{i}^{j}", r"\prescript{14}{6}{C}_{2}^{+}",
            r"\prescript{a}{b}{\sum}\limits_i^j", r"\prescript{a}{b}{\sum}\nolimits_i^j",
            r"\prescript{a}{b}{C}\limits_i^j"]:
    for display in [False, True]:
        append(tex, display)
for tail in [r"\limits", r"\limits_i", r"\limits^j", r"_i\limits", r"^j\limits", r"_i^j\limits",
             r"\limits_i^j", r"\limits\nolimits_i^j", r"\nolimits\limits_i^j", r"'_i\limits"]:
    for display in [False, True]:
        append(r"\prescript{a}{b}{\sum}" + tail, display)
assert len(cases) == 1172

wire = "".join(json.dumps({"tex": c["tex"], "options": {"Display": c["display"]}}, ensure_ascii=True) + "\n" for c in cases)
process = subprocess.run([node, "--jitless", str(directory / "differential/oracle.mjs"), "--asset-dir", str(assets)],
                         input=wire, text=True, capture_output=True, check=True)
responses = [json.loads(line) for line in process.stdout.split("\n") if line]
assert len(responses) == len(cases)
svg_cases, runtime_cases = [], []
for case, response in zip(cases, responses):
    case["TeX"] = case.pop("tex")
    case["original"] = response
    if "svg" in response:
        assert set(response) == {"svg"}
        svg_cases.append(case)
    else:
        assert set(response) == {"error"}
        runtime_cases.append(case)

metadata = {"mathjaxGitCommit": "ad8f5c21cb810236551da8c6512ba733e67357ee", "primaryAssets": hashes,
            "scope": "Complete unchanged original APIs; fresh frozen runtime per case; runtime-only cases have no SVG parity credit."}
payload = (json.dumps({**metadata, "cases": svg_cases}, ensure_ascii=True, indent=2) + "\n").encode()
with open(directory / "prescript_postslots_mathjax_3_2_2.json.gz", "wb") as output:
    with gzip.GzipFile(fileobj=output, mode="wb", mtime=0, filename="") as compressed:
        compressed.write(payload)
(directory / "prescript_postslots_runtime_mathjax_3_2_2.json").write_text(
    json.dumps({**metadata, "cases": runtime_cases}, ensure_ascii=True, indent=2) + "\n")
print("Original SVGs:", len(svg_cases), "Original runtimes:", len(runtime_cases))
