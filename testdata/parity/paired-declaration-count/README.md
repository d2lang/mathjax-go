# Empty paired-delimiter argument count

An explicitly empty argument-count string must skip argument consumption while an absent count still defaults to one. The before render swallowed the following fraction as a dummy argument; the original and fixed renders retain it. The complete reference corpus also covers string zero, JavaScript whitespace, pre/post templates and parameter substitution.

[String-zero parameter substitution](zero/README.md) has its own complete SVGs and screenshot comparison. The 494 references comprise 390 valid outputs and 104 original diagnostics; 94 valid and 104 diagnostic differences are fixed, with 296 controls. Independent review adds 108 fresh exact comparisons.

`comparison.png` is a Chromium screenshot of three complete, untouched D2 SVGs at one shared scale. The fixed/original SVGs and PNGs are byte-identical. The original is independently rendered through D2's hash-verified MathJax 3.2.2 bundle in a fresh runtime. See `manifest.json` for renderer/source revisions and all artifact hashes.

To regenerate, build D2 at the recorded renderer revision with a replacement for the before or fixed mathjax-go checkout, then render the exact source:

```sh
go mod edit -replace=github.com/d2lang/mathjax-go=/absolute/path/to/mathjax-checkout
go build -o /tmp/d2-parity .
/tmp/d2-parity --pad 24 --scale 1 --omit-version witness.d2 output.svg
```

For the independent original, use `testdata/parity/accent-italic-correction/oracle-adapter` as the replacement. Set `MATHJAX_GO_NODE`, `MATHJAX_GO_ORACLE_SCRIPT`, and `MATHJAX_GO_ORACLE_DIR` to Node, `testdata/differential/oracle.mjs`, and the pinned assets. The regression generator uses only the original runtime for expected outputs. These finite checks do not establish universal MathJax parity.
