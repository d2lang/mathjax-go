# Stretchable row children

The Go row renderer located the core `mo` of each embellished operator, then
compared those core pointers with the outer row children when deciding which
children contribute to the stretch target. A wrapper such as the `TeXAtom`
around `\bigl[` consequently counted as ordinary tall content. It incorrectly
enlarged the surrounding automatic fences.

For example, both of these now reproduce the complete original SVG:

```tex
\left[\bigl[x\bigr]\right]
\left[\Biggl[x\Biggr]\right]
```

Original MathJax measures the ordinary `x` when choosing the outer brackets;
the explicit inner brackets retain their authored size. The previous Go
render enlarged the outer brackets to match the inner brackets.

The implementation follows pinned MathJax 3.2.2 source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`:

- [CommonMrow.stretchChildren](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/output/common/Wrappers/mrow.ts#L86)
  asks each outer child whether it stretches, measures only children with no
  stretch state unless all children stretch, and delivers the final dimensions
  to each selected child's core operator.
- [CommonWrapper.canStretch](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/output/common/Wrapper.ts#L640)
  propagates the core's stretch state through embellished wrappers. The Go
  implementation already supports this delegation; the row now uses it.
- [CommonWrapper.getBBox/getOuterBBox](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/output/common/Wrapper.ts#L304)
  supports a temporary measurement when `save` is false. The all-stretch row
  branch uses that measurement, including its source default scale and style
  adjustments. Existing callers retain the usual cached-box path.

This also fixes the complete original rendering of the all-stretch scale
control `\left[\mmlToken{mo}[stretchy="true",mathsize="200%"]{[}\right]`.

## Complete original references

`row_stretch_mathjax_3_2_2.json` contains **1,020 complete original SVGs**, all
valid renderings. It records the source commit, the three verified frozen D2
bundle hashes, and baseline main131
`e9b0d9241c3df3562d56146ce04bd03c929fad1f`. Each expression is rendered in a
fresh VM with font cache none, `em=16`, `ex=8`, and its recorded display mode.
The baseline fails 538 of these references. The full deduplicated audit of
1,378 inputs has zero formerly exact regressions.

Coverage includes six fence pairs, four explicit delimiter sizes, ordinary
and tall content, display and script styles, mixed sizes, grouped and classed
operators, accents, scripts, positions, color, font and phantom wrappers,
fractions, roots, arrays, CD, Physics applications, and explicit `mmlToken`
stretch/minsize/maxsize/symmetry/scale/style attributes. Independent review
contributed 284 fresh comparisons: 218 exact, 152 newly exact, no regressions.

After rebasing onto main130, all 988 initial references still match, all 358 raw
observations remain byte-identical, and the baseline failure count remains
512. A further 36 fresh original comparisons cover the merged Physics vector
applications and array row-finalization changes with explicit nested fences:
all 36 match, including 30 newly exact results, with no regressions. Thirty-two
unique cases are included in the final 1,020-reference fixture; four overlap
existing cases. The final main131 recheck confirms all 1,020 exact references,
538 baseline discrepancies, and unchanged original/baseline/candidate outcomes
for all 358 residual controls.

## Preserved raw observations

`row_stretch_residuals.json` retains the other **358** inputs with complete
original, baseline and candidate responses. None was exact at baseline;
332 candidate responses remain byte-identical to baseline. The other 26
change as row classification improves but retain a separate difference.
They are not passing references, and no output normalization is used.

These controls primarily exercise rows containing only stretchable children,
fixed-size and mixed-scale early measurements, embellished scripts and
accents, authored spacing/padding, and existing Physics quantity parsing or
fence behavior. For example, `\left[\bigl[\bigr]\right]` and
`\Res[\Biggl[\Biggr]]` retain their previous differences. The original
`CommonMo.getStretchedVariant` does not invalidate cached boxes when choosing
a fixed glyph; the Go method currently does. This change preserves that
existing policy and records the remaining observations for a separate cache
and measurement audit. It does not replace valid original results with Go
output or drop the unsuccessful forms.

Regenerate both files with
`python3 testdata/generate_row_stretch.py PINNED_ASSETS NODE`. The generator
verifies every asset hash, uses batches of 24 fresh VMs, and updates only
original responses, preserving historical Go receipts. Run
`go test ./... -run TestRowStretchReferences` for the full exact-SVG check.
