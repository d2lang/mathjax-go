# drcases display-style evidence

The same D2 source and renderer show compact Go fractions and side-positioned sum limits before the fix, then display-size fractions and vertically stacked limits after the fix. The fixed and original complete D2 SVGs are byte-identical. The three-panel screenshot uses one shared scale and was visually inspected.

Before is actual merged mathjax-go `5985de47cd782beeee58344eb1ddaa0bbd5cc51e`; its production files match the source-bound `3c9e0bd` binary. After uses source `b25453663d1e4b6f12be0943d367bbb2aab03ae8`. Both use D2 renderer `01bc7ecdbdd04c13d6fe5df1967d2d9aa14ae579`. The original reference is independently rendered through D2 with its unmodified, hash-verified MathJax 3.2.2 bundle. Asset and artifact hashes are in `manifest.json`.

To reproduce, build D2 at the renderer revision with `github.com/d2lang/mathjax-go` replaced by a checkout of the before or after source, then run `d2 --pad 24 --scale 1 --omit-version witness.d2 output.svg`. For the original, use the evidence-only oracle adapter under `../accent-italic-correction/oracle-adapter`; set `MATHJAX_GO_NODE`, `MATHJAX_GO_ORACLE_SCRIPT` and `MATHJAX_GO_ORACLE_DIR` to the pinned Node binary, unchanged oracle script and verified original assets. The adapter is not production code and does not copy candidate SVGs.

All 96 previous D2 witnesses also remain byte-identical to their original references. Eight unrelated displaylines alignment gaps remain separately held in the source corpus; this demonstration does not establish complete MathJax parity.
