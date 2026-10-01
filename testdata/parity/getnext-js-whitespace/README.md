# JavaScript whitespace preserves matrix, braket and derivative arguments

Each formula contains actual U+FEFF after an active argument boundary. The original GetNext reader skips it: the fixed renderer restores the matrix, stretches the braket correctly, and reads the derivative variable as its second argument. Family captions are ordinary D2 labels outside the mathematical expressions.

`comparison.png` is a Chromium screenshot of three complete, untouched D2 SVGs at one shared scale. The fixed/original SVGs and PNGs are byte-identical. The original is independently rendered through D2's hash-verified MathJax 3.2.2 bundle in a fresh runtime. See `manifest.json` for renderer/source revisions and all artifact hashes.

To regenerate, build D2 at the recorded renderer revision with a replacement for the before or fixed mathjax-go checkout, then render the exact source:

```sh
go mod edit -replace=github.com/d2lang/mathjax-go=/absolute/path/to/mathjax-checkout
go build -o /tmp/d2-parity .
/tmp/d2-parity --pad 24 --scale 1 --omit-version witness.d2 output.svg
```

For the independent original, use `testdata/parity/accent-italic-correction/oracle-adapter` as the replacement. Set `MATHJAX_GO_NODE`, `MATHJAX_GO_ORACLE_SCRIPT`, and `MATHJAX_GO_ORACLE_DIR` to Node, `testdata/differential/oracle.mjs`, and the pinned assets. The regression generator uses only the original runtime for expected outputs. These finite checks do not establish universal MathJax parity.

The additional [Physics comparison](physics/README.md) demonstrates Differential, Expression and DiagonalMatrix. Both diagrams contain literal U+FEFF characters at the selected boundaries. Their ordinary D2 family captions stay outside the mathematical expressions. The public corpus binds all six active families to complete fresh original SVGs and records original runtime failures separately without parity credit.
