# Restore legacy matrix fence spacing

The legacy `\cases` command loses the same original thin space before its brace. This secondary D2 witness uses `\Huge x\cases{1}y` with its caption outside the formula.

`comparison.png` shows three complete, untouched D2 SVGs at one shared scale. The fixed and independently rendered original SVGs and PNGs are byte-identical; the before output visibly lacks the spacing. Both comparisons were viewed and checked for clipping. See `manifest.json` for all source revisions and artifact hashes. The before binary was built from local preview `883f28d1180b49d1d1f9f7421eedc30f181d887d`; its entire tracked tree exactly matches canonical baseline `fc40841854f0b3cdad53ee346a71646c478b7ba8` (tree `090f9f0c9c35c7bdeeaec697177edbfb8107918b`).

The fresh original corpus has 1,156 valid SVGs: 98 fixes and 1,058 unchanged controls. Another 64 original diagnostic SVGs are unchanged exact controls, for 1,220 strict full-SVG assertions. Sixteen original-unsupported `boldsymbol` diagnostics are preserved separately and receive no parity credit. There were zero original runtime failures and zero SVG regressions. The corpus covers all eight legacy Matrix commands, neighboring text and atom classes, inline/display mode, fonts and scripts, and ten environment controls.

To regenerate, use D2 at the recorded renderer revision with an external modfile replacing `github.com/d2lang/mathjax-go` with the recorded before or fixed checkout, then render the exact source:

```sh
go build -modfile /absolute/path/to/replaced-d2.mod -o /tmp/d2-parity .
/tmp/d2-parity --pad 24 --scale 1 --omit-version witness.d2 output.svg
```

For the original, build D2 with `testdata/parity/accent-italic-correction/oracle-adapter` as the replacement, then set `MATHJAX_GO_NODE`, `MATHJAX_GO_ORACLE_SCRIPT`, and `MATHJAX_GO_ORACLE_DIR` to Node, `testdata/differential/oracle.mjs`, and the pinned assets. These finite checks do not establish universal MathJax parity.
