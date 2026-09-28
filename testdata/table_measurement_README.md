# Table measurement references

`table_measurement_mathjax_3_2_2.json` stores 358 complete SVG strings from
D2's frozen MathJax 3.2.2 bundle, upstream commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. The fixture records the three
original asset hashes. Each case uses a fresh VM, font cache none, `em=16`,
`ex=8`, and the recorded display setting. No SVG values are normalized or
produced by the Go implementation.

There are 118 public TeX cases and 240 MathML layout controls. The public
cases cover CD arrow directions and minimum sizes, labels, tall and shifted
content, font sizes, long horizontal arrows, ordinary matrices, and tables in
scripts, fractions, roots, fences, and grouped nested tables. The controls
cover equal columns and rows, fixed widths, percentage widths, automatic and
fitted columns, mixed column specifications, frames, and rules.

For the controls, `tableAttributes` records the exact attributes applied to
each compiled `mtable`. The original generator uses the TeX input jax's
post-filter, before SVG wrappers exist. The Go test applies the same attributes
to its completed MathML tree before typesetting. These controls isolate table
layout without inventing public TeX syntax for MathML-only options.

The primary behavior is
[`CommonMtable`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/output/common/Wrappers/mtable.ts):
the constructor computes column specifications, stretches rows and columns,
and measures cells lazily through `getTableData`. Equal columns with automatic
width, percentage widths with fitted columns, fixed widths with automatic or
fitted columns, and equal rows request measurements before stretching. Other
tables measure afterward. The first measurement remains cached.

The former unconditional early measurement underestimated rows and columns
containing stretched vertical arrows. Before this fix, 108 of the 358 SVG
references fail. The two historical `cd-height` qualifications in
`tex_mu_dimensions_test.go` now require the original SVG and measurements;
their historical receipts remain unchanged on disk.

A visible public witness is
`\minCDarrowheight0pt\begin{CD}\raise3em{A}@VVV\\B\end{CD}`. The old layout
does not reserve the stretched arrow's depth, so it extends below the following
row. The corrected table reserves the same space as the original renderer.

Regenerate with
`node --jitless testdata/generate_table_measurement.cjs PINNED_ASSETS`.
Run `go test ./... -run 'TestTableMeasurementReferences|TestTeXMuDimensionsPinnedReferences'`.
