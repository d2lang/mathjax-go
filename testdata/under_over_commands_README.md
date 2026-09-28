# Under and over decoration commands

These references use the original D2 MathJax 3.2.2 component at
`ad8f5c21cb810236551da8c6512ba733e67357ee`. Exact tests compare complete,
unmodified original SVGs. The separate raw inventory retains every unresolved
original and the observed baseline/candidate outputs. No SVG attribute or
numeric value is normalized for the exact tests.

## Source contract

[BaseMappings](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMappings.ts#L516)
registers `overparen` and `underparen` with U+23DC/U+23DD, and
`overleftrightarrow` and `underleftrightarrow` with U+2194. All four use the
existing UnderOver handler, including its argument parsing, accent defaults,
and script behavior. Dynamic registered commands retain their existing
precedence.

Restoring these commands exposed two shared implementation differences:

- [MmlMunderover](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/core/MmlTree/MmlNodes/munderover.ts)
  asks the base's node-specific `coreMO()` whether limits move. A compound
  row is its own core, even when its first descendant is a sum. Conversely,
  an embellished row can have a nonzero core index because preceding spaces
  are ignored. The inherited script-level decision now uses the existing
  source-defined core lookup, including wrapper and selected-action rules.
- [SVGmo](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/output/svg/Wrappers/mo.ts)
  uses `(W+w)/2` for the right edge of a middle glyph and `1.5*(X/w)` for
  horizontal extender scaling. Its clipping calculation rounds `s*w` before
  subtracting X. These operation boundaries affect rounded SVG viewBoxes.
  An explicit `float64` conversion preserves the original separate rounding;
  the [Go floating-point operator rules](https://go.dev/ref/spec#Floating-point_operators)
  otherwise permit fused multiply-add/subtract operations across statements.
  The live operand/IEEE-bit observations and arm64 instruction receipt in
  `under_over_arithmetic_receipts.json` demonstrate that distinction without
  changing serialization tolerances.

The same explicit product rounding is preserved in CommonScriptbase's
italic-correction/skew calculation. It prevents an otherwise identical
accent width from rounding a fraction's denominator position differently.

The corpus includes existing line, brace, arrow and wide-accent controls;
compound, embellished and spacelike bases; square roots, fractions, tables,
font/color and script scopes; explicit movablelimits attributes; optional
limits; authored widths on both sides of rounding ties; and dynamic command
replacement. Historical primary references remain unchanged.

## Reproduction and qualification

The generator verifies all three frozen asset hashes and constructs a fresh
original runtime per expression, in bounded subprocess batches. It regenerates
original fields only and never invokes Go or selects an assertion subset.

The inventory has 3,790 distinct TeX/display inputs: 3,766 complete original
SVG assertions (including 32 original error renderings) and 24 raw residuals.
Against the published PR154 head `a5c6c64d` (internal tree
`74787f5d`), 2,174 inputs
become byte-exact and no formerly exact input changes to a nonexact result.
The 144 new empty-option controls cover decorations around `matrix*`,
`smallmatrix*`, and `bmatrix*` with absent, empty, whitespace, and explicit
alignment options. All are exact. The 16 low-float brace controls and 520
variable-limit literal controls are also exact.

The 24 remaining original-valid differences are the existing MathFont
identifier grouping of `AB`. Eight are newly reachable through the four new
commands; sixteen are unchanged existing-decoration controls. Each candidate
has the same complete outlined glyph geometry as its original. These remain
raw: no tree or SVG normalization makes them pass an exact test.
`under_over_residual_qualification.json` binds the geometry review to the
unmodified original, baseline, and candidate SVG hashes.

`under_over_commands_inventory.json` records the public-input overlap using
both TeX and display mode. This per-fix inventory is not a globally unique
MathJax coverage count. The 92 already-published input pairs are explicitly listed; the other 3,674
asserted input pairs are first published in this change. The overlap scan
uses the complete published PR154 fixture tree before its merge. The previous
original observations are unchanged; only the classification of now-exact
inputs differs after the source prerequisites compose.

Regenerate the complete original SVG fields with:

```sh
node --jitless testdata/generate_under_over_commands.cjs PINNED_ASSETS
```

The exact regression test is `TestUnderOverCommandReferences`. The shared
source changes are also checked against the historical accent, fraction,
limits, prime, and stretchy-operator references without changing their
original expectations.
