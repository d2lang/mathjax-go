# D2 subnumcases shared-frame evidence

Original MathJax accepts this exact input and renders `xyxyz + Z`. Before the fix, Go reports `Missing \end{subnumcases}`. The fixed complete D2 SVG and its PNG screenshot are byte-identical to the independently rendered original MathJax 3.2.2 reference.

`comparison.png` includes untouched full D2 SVGs at one shared scale. The before SVG paints a dark diagnostic block; the caption transcribes its exact `data-mjx-error` value. The raw SVGs, individual screenshots, and exact D2 source are included. Both new witnesses and all 103 previously committed witnesses match their complete original D2 references. Root inspected both comparisons and every individual PNG.

This witness directly pops the equation array and then both references to the same CasesBegin object. It therefore bypasses normal cases decoration. No visible `F=` or left brace is expected: the original reference determines the result.

- D2 renderer: `01bc7ecdbdd04c13d6fe5df1967d2d9aa14ae579`.
- Before Go: actual merged main `2a19a9626f9136931a5866ca14288d2d64998f0c`.
- Fixed Go: the five exact production hashes in `manifest.json`, recorded through a Go build overlay.
- Original: D2's unchanged frozen MathJax 3.2.2 assets; a fresh original VM per conversion.

## Reproduction

Use separate mathjax-go checkouts at the before revision and this fix, with the D2 renderer revision above. In a dedicated D2 checkout, point its replacement at each mathjax-go checkout and render the exact witness:

```sh
go mod edit -replace=github.com/d2lang/mathjax-go=/absolute/path/to/mathjax-checkout
go build -o /tmp/d2-cases-alias .
/tmp/d2-cases-alias --pad 24 --scale 1 --omit-version /absolute/path/to/witness.d2 /tmp/output.svg
```

For the independent original, use the unchanged `testdata/parity/accent-italic-correction/oracle-adapter` as the replacement. Set `MATHJAX_GO_NODE` to the pinned Node executable, `MATHJAX_GO_ORACLE_SCRIPT` to `testdata/differential/oracle.mjs`, and `MATHJAX_GO_ORACLE_DIR` to the three verified frozen assets. The recorded builds used scratch `-modfile` copies and preserved the D2 checkout's existing module bytes.

All 196 new complete SVG references and the earlier cases tests pass: 136 fixes and 60 diagnostic controls. The original-only generator reproduces all 252 complete original APIs, with 56 original runtime failures stored separately and excluded from SVG parity. The broader replay completed 146,891 distinct requests and retained 164,971 source/consumer associations with no regression flags. Original-runtime response changes and cached SVG attribute-order variation are recorded separately and receive no Cases fix credit. Independent qualification and exact-commit full/race/vet/WASM and hosted checks are separate validation gates. These finite results do not establish full MathJax parity.
