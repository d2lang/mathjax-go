# SPDX-License-Identifier: Apache-2.0
"""Regenerate original references: python3 SCRIPT PINNED_ASSETS [NODE]."""
import hashlib
import json
from pathlib import Path
import subprocess
import sys

HERE = Path(__file__).resolve().parent
ASSETS = Path(sys.argv[1]).resolve()
NODE = sys.argv[2] if len(sys.argv) > 2 else "node"
FILES = ["cramped_substack_mathjax_3_2_2.json", "cramped_substack_residuals.json"]
for name in FILES:
    path = HERE / name
    fixture = json.loads(path.read_text())
    for asset, expected in fixture["oracleAssetsSHA256"].items():
        actual = hashlib.sha256((ASSETS / asset).read_bytes()).hexdigest()
        if actual != expected:
            raise RuntimeError("unverified " + asset)
    # Bound process memory; oracle.mjs creates a fresh runtime per expression.
    for start in range(0, len(fixture["cases"]), 24):
        batch = fixture["cases"][start:start + 24]
        requests = "".join(json.dumps({"tex": c["tex"], "options": {"display": c["display"]}}) + "\n" for c in batch)
        result = subprocess.run([NODE, "--jitless", str(HERE / "differential/oracle.mjs"), "--asset-dir", str(ASSETS)], input=requests, text=True, capture_output=True, check=True)
        responses = [json.loads(line) for line in result.stdout.rstrip('\n').split('\n')]
        if len(responses) != len(batch):
            raise RuntimeError("incomplete original response batch")
        for c, response in zip(batch, responses):
            if "original" in c:
                c["original"] = response
            else:
                if "svg" not in response:
                    raise RuntimeError(response)
                c["svg"] = response["svg"]
    path.write_text(json.dumps(fixture, indent=2, ensure_ascii=False) + "\n")
