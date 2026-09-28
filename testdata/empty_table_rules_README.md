# Empty-table attribute lists

The frozen MathJax 3.2.2 source at `ad8f5c21cb810236551da8c6512ba733e67357ee`
uses `splice(n)` in CommonMtable.getColumnAttributes/getRowAttributes, where
`n` is the actual column or row count minus one. For a zero-column/row table,
`splice(-1)` retains every authored list entry except its last entry. Go
previously clamped the count to zero and discarded every spacing and line
entry. This changes the width of an empty `array{c|cc}` and the geometry of
surrounding boxes and fences.

The change retains the source negative-index behavior only at the four
columnspacing, rowspacing, columnlines and rowlines callers. Other attribute
callers still pass nonnegative counts. SVGmtable's line-position arithmetic
can read a nonexistent cell dimension for an empty table: JavaScript converts
that undefined numeric value to NaN. The bounded rule-position accessor keeps
that source output instead of indexing beyond a Go slice. Original `NaN`
line coordinates are preserved verbatim; they are not presented as visible
lines. The width reserved by the retained rule entries is still visible in
surrounding boxes and adjacent content.

The public fixture contains 284 complete original SVGs with no `hline` or
`hdashline` commands: empty tables, an explicit empty row, a real empty group,
partial frames, boxes, fractions, scripts, fences, accents and nested tables.
Ninety differ from main158, and 194 are unchanged controls. All fixture inputs
are independent of the separate horizontal-rule parser implementation.

The renderer fixture contains 80 observations of the actual frozen
getColumnAttributes/getRowAttributes methods: all four lists, authored zero
through three entries, and zero through four rows/columns. It preserves both
getAttributeArray's input list and the final method return. Tests compare the
negative-splice/list-repetition result to those observations without deriving
expected results from Go.

```sh
node --jitless testdata/generate_empty_table_rules.cjs PINNED_ASSETS
go test ./... -run 'TestEmptyTable(RuleOriginalReferences|ListOriginalMethods)'
```

The generator verifies all three pinned asset hashes, uses fresh original
MathJax runtimes and never runs Go or chooses references by the candidate
result. Complete original SVGs, including empty-table NaN coordinates, remain
unmodified. The original corpus and intermediate failed prototype outputs
are retained in the task audit; no failure is hidden by replacing a reference.
