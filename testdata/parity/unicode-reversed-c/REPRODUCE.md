# Reproduction

Use D2 revision `01bc7ecdbdd04c13d6fe5df1967d2d9aa14ae579`. For Go output, build that D2 renderer with `github.com/d2lang/mathjax-go` replaced by the exact before or candidate source in `manifest.json`, then run:

```sh
d2 --pad 24 --scale 1 --omit-version witness.d2 output.svg
```

For the original reference, use the evidence-only adapter in `../accent-italic-correction/oracle-adapter`. Set `MATHJAX_GO_NODE` to the pinned Node executable, `MATHJAX_GO_ORACLE_SCRIPT` to the unchanged `testdata/differential/oracle.mjs`, `MATHJAX_GO_ORACLE_DIR` to the verified MathJax 3.2.2 asset directory, and `NODE_OPTIONS=--jitless`. Verify all hashes in `manifest.json` first. The adapter invokes the original oracle, which creates fresh VM/document/handler state per conversion; it does not read Go-produced SVGs. Use the same D2 flags above.

Compare complete files: `after.svg` must equal `original.svg` byte for byte. The before must differ, with the central U+2183 glyph slanted. The original reference and before capture were not regenerated during the candidate validation.

For the screenshot, display the three saved SVGs without editing their contents. Use their common 121×93 intrinsic dimensions and one shared scale of 300/93, rendering each at390×300 CSS pixels with deviceScaleFactor2. Keep labels for merged main181 (6d88373 / sourceb254536), candidate0158db9, and original MathJax3.2.2. The exact retained comparison is3000×1562 PNG pixels. A browser screenshot cannot replace the full-SVG byte comparison.
