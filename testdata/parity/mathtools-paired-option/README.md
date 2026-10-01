# Mathtools pairedDelimiters TeX setting

The original accepts the documented pairedDelimiters key through mathtoolsset. Before the fix, Go stops with UnknownKeyVal. The fixed renderer accepts the setting and keeps the separately declared scalable norm intact, exactly matching the frozen original. This witness tests the active TeX setter; custom JavaScript delimiter-object initialization remains a separate D2 API boundary.

`comparison.png` is a Chromium screenshot of three complete, untouched D2 SVGs at one shared scale. The fixed/original SVGs and PNGs are byte-identical. The original is independently rendered through D2's hash-verified MathJax 3.2.2 bundle in a fresh runtime. See `manifest.json` for renderer/source revisions and all artifact hashes.

To regenerate, build D2 at the recorded renderer revision with a replacement for the before or fixed mathjax-go checkout, then render the exact source:

```sh
go mod edit -replace=github.com/d2lang/mathjax-go=/absolute/path/to/mathjax-checkout
go build -o /tmp/d2-parity .
/tmp/d2-parity --pad 24 --scale 1 --omit-version witness.d2 output.svg
```

For the independent original, use `testdata/parity/accent-italic-correction/oracle-adapter` as the replacement. Set `MATHJAX_GO_NODE`, `MATHJAX_GO_ORACLE_SCRIPT`, and `MATHJAX_GO_ORACLE_DIR` to Node, `testdata/differential/oracle.mjs`, and the pinned assets. The regression generator uses only the original runtime for expected outputs. These finite checks do not establish universal MathJax parity.

The standalone PNGs were independently captured in separate fresh Chromium pages with GPU disabled, after image decoding, document fonts, and two animation frames. Two complete capture runs produced identical PNGs; fixed/original decoded RGBA pixels and PNG bytes are equal. `manifest.json` retains the raster provenance. The tracked `capture-screenshots.cjs` reproduces the comparison and individual screenshots with Playwright (set `PLAYWRIGHT_MODULE` and `CHROMIUM_EXECUTABLE` if necessary). The original SVGs are unchanged.
