# Match JavaScript whitespace in delimiters

The D2 source contains an actual U+FEFF BOM immediately after left and right. Go treated those characters as delimiters and rendered an error; the original skips them and renders scalable parentheses around the fraction. The fixed and independently rendered original results match exactly.

`comparison.png` is a Chromium screenshot of three complete, untouched D2 SVGs at one shared scale. The fixed/original SVGs and PNGs are byte-identical. The original is independently rendered through D2's hash-verified MathJax 3.2.2 bundle in a fresh runtime. See `manifest.json` for renderer/source revisions and all artifact hashes.

To regenerate, build D2 at the recorded renderer revision with a replacement for the before or fixed mathjax-go checkout, then render the exact source:

```sh
go mod edit -replace=github.com/d2lang/mathjax-go=/absolute/path/to/mathjax-checkout
go build -o /tmp/d2-parity .
/tmp/d2-parity --pad 24 --scale 1 --omit-version witness.d2 output.svg
```

For the independent original, use `testdata/parity/accent-italic-correction/oracle-adapter` as the replacement. Set `MATHJAX_GO_NODE`, `MATHJAX_GO_ORACLE_SCRIPT`, and `MATHJAX_GO_ORACLE_DIR` to Node, `testdata/differential/oracle.mjs`, and the pinned assets. The regression generator uses only the original runtime for expected outputs. These finite checks do not establish universal MathJax parity.

The source contains actual U+FEFF immediately after `\left` and `\right`; BOM is invisible in the screenshot's source block. An escaped spelling of its formula is:

```text
F(x) = \left[U+FEFF](\frac{a}{b}\right[U+FEFF])
```

`[U+FEFF]` above denotes the actual character rather than text to enter. See [`getdelimiter_whitespace_README.md`](../../getdelimiter_whitespace_README.md) for the complete original reference inventory, diagnostic exclusions, and regeneration command.
