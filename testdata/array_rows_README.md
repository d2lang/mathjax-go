# Array row and final-entry references

These fixtures cover source `ArrayItem` row finalization across base arrays,
AMS alignments, Cases and Mathtools environments. An authored row separator
closes an entry and a row even when no nodes were emitted. At `EndTable`, the
pending entry becomes a row only if it has emitted nodes or the row already
contains an explicitly closed entry. Raw whitespace is not a reliable test:
font declarations, labels and HFill emit no nodes; empty groups, styles and
leaf spacing nodes do emit nodes.

The implementation also preserves JavaScript `slice(0, -1)` semantics in
`EqnArrayItem.extendArray`. A table with zero columns must not index a Go slice
with a negative length. Before this fix, the valid input
`\begin{aligned}\\\end{aligned}` panicked. The visual witness
`\begin{matrix}a\\\\b\end{matrix}` lost its authored empty middle row.

Primary source is MathJax 3.2.2 commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`:

- [ArrayItem checkItem, EndEntry, EndRow and EndTable](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseItems.ts#L900-L1050)
- [EqnArrayItem.EndTable and extendArray](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseItems.ts#L1165-L1193)
- [MultlineItem.EndTable](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/ams/AmsItems.ts)
- [MultlinedItem inheritance and finalization](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/mathtools/MathtoolsItems.ts)

`array_rows_mathjax_3_2_2.json` contains **2,197 complete original SVGs**:
1,981 valid representations and 216 original error SVGs. The main129 baseline
`d8ce2ffc052369e564bdd26b41987bb0ad4ac229` fails 687 of these references,
including eight valid inputs that panic. The other 1,510 are unchanged controls.
`TestEmptyAlignedRowsDoNotPanic` directly exercises all eight formerly
crashing Render calls and display-mode Measure calls, and checks the authored
blank-row count.
The inventory includes both display modes, empty/leading/interior/trailing
rows, repeated row separators, multiple empty cells, tags/labels, HFill-only
cells, pending functions and negations, fonts, colors, size declarations,
empty groups, leaf mspaces, fractions, nested command matrices, and Mathtools
special rows. CD and command-matrix controls use their existing finalization.
All 36 former HFill-only final-row residuals are now exact; 24 additional
unique inputs from that historical inventory are included here. The original
HFill residual file stays unchanged as a record of its earlier audit.

The complete deduplicated exploratory inventory was 2,596 inputs. Its other
**399** references are retained in `array_rows_residuals.json` with raw original,
baseline and candidate results; they are not passing expectations. There were
zero formerly exact regressions. The residuals comprise 209 original-valid
representations, 152 error SVGs and 38 exceptions thrown by original MathJax
for empty NumCases tables. They exercise previously unresolved alignedat and
flalign geometry, Mathtools special-row layouts, Cases text cells, and
unsupported or differently diagnosed tag/alignment contexts. Fifty-three
residual outputs change because their row structure is corrected but another
difference remains. None are replaced with Go output as an expected SVG.

The source row scanner and unrelated geometry remain outside this fix. In
particular, this change does not reinterpret nested raw row delimiters or add
new `\cr`/`\newline` ownership.

Regenerate both files directly from the original runtime:

```sh
node --jitless testdata/generate_array_rows.cjs /path/to/pinned/assets
```

The generator validates SHA-256 for all three D2 assets, initializes a fresh
VM for every expression and regenerates every retained input, including raw
residuals and original exceptions. It never invokes Go or chooses entries by
comparison with the candidate. The exact test compares the whole SVG string,
without normalization.
