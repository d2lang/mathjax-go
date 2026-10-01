# Physics: accept JavaScript whitespace in matrix dimensions

The matrix size starts with U+FEFF before 2. The shared argument fix restores JavaScript whitespace, decimal-prefix Number parsing and strict dimension checks; 2588 complete frozen-original SVG assertions cover 118 valid fixes, 92 diagnostic fixes and 2378 unchanged controls.

`comparison.png` is a Chromium screenshot of three complete, untouched D2 SVGs at one shared scale. The fixed/original SVGs and PNGs are byte-identical. The original is independently rendered through D2's hash-verified MathJax 3.2.2 bundle in a fresh runtime. See `manifest.json` for renderer/source revisions and all artifact hashes.

To regenerate, build D2 at the recorded renderer revision with a replacement for the before or fixed mathjax-go checkout, then render the exact source:

```sh
go mod edit -replace=github.com/d2lang/mathjax-go=/absolute/path/to/mathjax-checkout
go build -o /tmp/d2-parity .
/tmp/d2-parity --pad 24 --scale 1 --omit-version witness.d2 output.svg
```

For the independent original, use `testdata/parity/accent-italic-correction/oracle-adapter` as the replacement. Set `MATHJAX_GO_NODE`, `MATHJAX_GO_ORACLE_SCRIPT`, and `MATHJAX_GO_ORACLE_DIR` to Node, `testdata/differential/oracle.mjs`, and the pinned assets. The regression generator uses only the original runtime for expected outputs. These finite checks do not establish universal MathJax parity.

Both witness arguments contain literal U+FEFF (BOM) whitespace; it is invisible in the screenshot source. The second strut witness has [its own comparison](strut/comparison.png), [source](strut/witness.d2), and complete before/after/original SVGs in `strut/`.
