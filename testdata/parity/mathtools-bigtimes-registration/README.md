# Mathtools bigtimes macro registration

A documented Mathtools declaration supplies a bold formatter, then bigtimes renders a full product equality with lower and upper operator limits. Before the fix Go reports an undefined bigtimes command; after restoring the exact original macro registration, the complete D2 output matches the original.

`comparison.png` is a Chromium screenshot of three complete, untouched D2 SVGs at one shared scale. The fixed/original SVGs and PNGs are byte-identical. The original is independently rendered through D2's hash-verified MathJax 3.2.2 bundle in a fresh runtime. See `manifest.json` for renderer/source revisions and all artifact hashes.

To regenerate, build D2 at the recorded renderer revision with a replacement for the before or fixed mathjax-go checkout, then render the exact source:

```sh
go mod edit -replace=github.com/d2lang/mathjax-go=/absolute/path/to/mathjax-checkout
go build -o /tmp/d2-parity .
/tmp/d2-parity --pad 24 --scale 1 --omit-version witness.d2 output.svg
```

For the independent original, use `testdata/parity/accent-italic-correction/oracle-adapter` as the replacement. Set `MATHJAX_GO_NODE`, `MATHJAX_GO_ORACLE_SCRIPT`, and `MATHJAX_GO_ORACLE_DIR` to Node, `testdata/differential/oracle.mjs`, and the pinned assets. The regression generator uses only the original runtime for expected outputs. These finite checks do not establish universal MathJax parity.

The singular documented `\DeclarePairedDelimiterX` declaration supplies the authored formatter. The `\Huge` group makes the complete product equality legible at the shared screenshot scale. The before renderer uses verified canonical main `fe29d3ad1b32f04ce864e0b410267e5b6d666c99`; the fixed registration is source commit `e21f53151ff810ed20879c97e5760cf5cf9705e2`, with the full witness bound into the frozen references at `35ca91f962b2b67fe598da3aed6b243a1a0e6f61`.

The [regression scope and generator](../../mathtools_bigtimes_README.md) preserve 358 valid-original SVGs: 170 repaired formulas and 188 independent controls. Ten bare-command original diagnostics remain separately retained and receive zero original-valid or diagnostic parity credit: the existing Go-only `boldsymbol` formatter remains available, while the original frozen package set omits that dependency. The production change restores only the exact original `bigtimes` registration.
