# D2 array bottom positioning evidence

The witness contains `A+\begin{array}[b]{r|l}1&2\\3&4\end{array}+B` at font size 32. Before the fix, the optional position and column setup appear as math content. The fixed renderer consumes `[b]` before the column specification and aligns the surrounding `A+B` with the array's second row. The complete fixed D2 SVG is byte-identical to the independent original MathJax 3.2.2 render.

Both new positioning witnesses and all 100 previously published D2 witnesses match their complete original SVGs with this production source. This is 102 D2 comparisons; the two new witnesses cover four existing original106 records across both display modes.

`comparison.png` shows the three untouched SVGs at one shared scale. Individual PNGs, raw SVGs and the D2 source are included. Root and independent reviewers inspected both comparisons: labels and source are readable, all formulas are unclipped, and the fixed and original panels match. The individual fixed and original PNGs are also byte-identical.

- D2 renderer: `01bc7ecdbdd04c13d6fe5df1967d2d9aa14ae579`.
- Before Go: actual merged main `16801f1f0e8e5c0b1582224cce85140b1093ed38`.
- Fixed production source: `d114575e4d54542cf119eafa77750a6c36b1a3f9`.
- Original: D2's unchanged, SHA-256-verified frozen MathJax 3.2.2 assets, with a fresh original runtime for each conversion.

## Reproduction

Use separate Go checkouts at the before and fixed revisions and the same D2 renderer revision. In a dedicated D2 checkout, replace its mathjax-go dependency with each checkout, build, and run the included witness:

```sh
go mod edit -replace=github.com/d2lang/mathjax-go=/absolute/path/to/mathjax-checkout
go build -o /tmp/d2-array-position .
/tmp/d2-array-position --pad 24 --scale 1 --omit-version /absolute/path/to/witness.d2 /tmp/output.svg
```

The witness uses `|||latex` and `|||` delimiters so the column-rule pipe in `{r|l}` remains inside the LaTeX block. Keep those delimiters and the exact TeX unchanged.

For the original render, replace the dependency with the unchanged `testdata/parity/accent-italic-correction/oracle-adapter` in this repository. Set `MATHJAX_GO_NODE` to the Node executable, `MATHJAX_GO_ORACLE_SCRIPT` to `testdata/differential/oracle.mjs`, and `MATHJAX_GO_ORACLE_DIR` to the directory containing the three verified original assets. Original MathJax produces the reference independently. Hashes and completed render/visual receipts are recorded in `manifest.json`.

These artifacts attest completed D2 and visual checks. Full Go, race, vet, WASM compilation and exact-head hosted checks must pass before merge; no final publication or merge revision is asserted here.
