# D2 cases array-pop evidence

Original MathJax accepts the included input and renders pending `x`, the later matrix, and `+Z`. Go before the fix reports `Missing \end{numcases}`. The fix continues the same input after the EqnArray is popped and preserves the CasesBegin two-End decoration sequence. The complete fixed D2 SVG is byte-identical to the independently generated original MathJax 3.2.2 reference.

The helper declaration in this witness supplies `empheqlbrace` through the public package profile. Keep it and the exact input unchanged. Original MathJax decorates the oldest pending node internally; no visible left brace or `F=` is claimed here.

`comparison.png` shows the three untouched full D2 SVGs at one shared scale. The before SVG paints its diagnostic as a dark block; the caption transcribes the exact `data-mjx-error` value. Individual PNGs, raw SVGs, and the D2 source are included. The fixed and original PNGs are also byte-identical. The new witness and all 102 previously published witnesses match their original full D2 SVGs.

- D2 renderer: `01bc7ecdbdd04c13d6fe5df1967d2d9aa14ae579`.
- Before Go: actual merged main `c24058161bb5df6af3f269ce65030c94079b952a`.
- Fixed Go: production source hashes in `manifest.json`, applied as a three-file Go build overlay for the recorded D2 build.
- Original: unchanged, hash-verified D2-frozen MathJax 3.2.2 assets, with a fresh original runtime for every conversion.

## Reproduction

Use separate Go checkouts at the before revision and this fix and the same D2 renderer revision. In a dedicated D2 checkout, point the mathjax-go replacement at each checkout and render the exact witness:

```sh
go mod edit -replace=github.com/d2lang/mathjax-go=/absolute/path/to/mathjax-checkout
go build -o /tmp/d2-cases-array-pop .
/tmp/d2-cases-array-pop --pad 24 --scale 1 --omit-version /absolute/path/to/witness.d2 /tmp/output.svg
```

The three-pipe LaTeX block delimiters preserve the input verbatim. For the independent original, point the replacement at the unchanged `testdata/parity/accent-italic-correction/oracle-adapter`. Set `MATHJAX_GO_NODE` to Node's executable, `MATHJAX_GO_ORACLE_SCRIPT` to `testdata/differential/oracle.mjs`, and `MATHJAX_GO_ORACLE_DIR` to the directory containing the three verified original assets.

The recorded root D2 build used a separate `-modfile` copy and the production overlay; the D2 checkout's existing module bytes were preserved. Its first prior-witness reader attempted to resolve the merged SHA in a different clone; that reader failed after the new render completed. The 102 prior renders resumed from the actual main clone, using the already built binary and without regenerating originals. Both receipts are retained by hash.

These artifacts attest completed D2 and visual comparisons and are installed with the exact production bytes used by the recorded build. The 100 strict original-reference cases and four native continuation tests passed, and fresh original-only regeneration preserved all 112 complete source APIs. Protected replay was independently qualified. Complete regression checks and hosted CI are recorded separately for the exact published commit before merge; these finite comparisons do not establish full MathJax parity.
