# Plain TeX numbered matrix commands

The public fixture compares complete, unmodified SVG from D2's original MathJax 3.2.2 runtime at `ad8f5c21cb810236551da8c6512ba733e67357ee`. The generator verifies all three frozen asset hashes and runs the original runtime with Node `--jitless`.

Source: `base/BaseMappings.ts`, `BaseMethods.Matrix`, and `BaseItems.ArrayItem.EndRow`. `\eqalignno` and `\leqalignno` share the existing `\eqalign` display style and spacing, with labels on the right or left. Only rows with exactly three entries become `mlabeledtr`; their third entry moves to the label position. Rows with other lengths remain ordinary rows.

The 258 cases cover right/left labels, zero through five columns, mixed labeled/unlabeled rows, empty labels, scripts, fractions, large labels, styles, colors, nested matrices/environments, macros, row gaps, tags, missing arguments/braces, malformed dimensions, and invalid scripts. Shared Matrix script boundaries are also covered: an unbraced ArrayItem cannot satisfy a pending script, and an alignment-cell boundary is distinct from the matrix's closing brace. Ordinary matrix, eqalign, and cases-environment controls remain exact.

Regenerate with `node --jitless testdata/generate_matrix_numbered.cjs /path/to/pinned/assets`. All 258 cases match the original; 236 differed on the preceding main commit `729d74d`.

D2 witness math: `\eqalignno{x+y&=3&(1)\\2x-y&=0&(2)}` (and `\leqalignno` for left labels).
