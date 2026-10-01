# Original CSS border painting and resets

This witness exercises functional RGB border color, compact shorthand, and independent side width. The final source also preserves empty resets, JavaScript whitespace, first-split validation boundaries, and original zero-count dashed strokes. The 2,364 complete SVG assertions cover 2,308 distinct inputs; 1,512 assertions differ from the initial baseline and 852 are unchanged controls. An additional 446 direct source Styles/TRBL observations verify scanning and normalization.

[An additional empty-reset comparison](reset/README.md) includes complete before, fixed, and original SVGs and screenshots.

`comparison.png` is a Chromium screenshot of three complete, untouched D2 SVGs at one shared scale. The fixed/original SVGs and PNGs are byte-identical. The original is independently rendered through D2's hash-verified MathJax 3.2.2 bundle in a fresh runtime. See `manifest.json` for renderer/source revisions and all artifact hashes.

To regenerate, build D2 at the recorded renderer revision with a replacement for the before or fixed mathjax-go checkout, then render the exact source:

```sh
go mod edit -replace=github.com/d2lang/mathjax-go=/absolute/path/to/mathjax-checkout
go build -o /tmp/d2-parity .
/tmp/d2-parity --pad 24 --scale 1 --omit-version witness.d2 output.svg
```

For the independent original, use `testdata/parity/accent-italic-correction/oracle-adapter` as the replacement. Set `MATHJAX_GO_NODE`, `MATHJAX_GO_ORACLE_SCRIPT`, and `MATHJAX_GO_ORACLE_DIR` to Node, `testdata/differential/oracle.mjs`, and the pinned assets. The regression generator uses only the original runtime for expected outputs. These finite checks do not establish universal MathJax parity.
