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

The public fixture contains 376 complete original SVGs with no `hline` or
`hdashline` commands: empty tables, an explicit empty row, a real empty group,
partial frames, boxes, fractions, scripts, fences, accents and nested tables.
One hundred four differ from actual main163 `c09d0460db08b943c49003fbe2409402f9c2fca4`,
and 272 are unchanged controls. All fixture inputs
are independent of the separate horizontal-rule parser implementation.

The renderer fixture contains 80 observations of the actual frozen
getColumnAttributes/getRowAttributes methods: all four lists, authored zero
through three entries, and zero through four rows/columns. It preserves both
getAttributeArray's input list and the final method return. Tests compare the
negative-splice/list-repetition result to those observations without deriving
expected results from Go.

```sh
node --jitless testdata/generate_empty_table_rules.cjs PINNED_ASSETS
go test ./... -run TestEmptyTable
```

The generator verifies all three pinned asset hashes, uses fresh original
MathJax runtimes and never runs Go or chooses references by the candidate
result. Complete original SVGs, including empty-table NaN coordinates, remain
unmodified. The original corpus and intermediate failed prototype outputs
are retained in the task audit; no failure is hidden by replacing a reference.

## Standalone qualification

The standalone renderer on actual main163 preserves all 376 original SVGs
exactly. The independent review contributes 92 distinct source controls, all
exact, with 14 fixes. A broader 1,858-input replay includes this public corpus,
the separately preserved HLine inventory, supported NumCases overrides and
determinant intersections: 838 inputs are exact, 104 become exact, and no
previously exact input or unresolved output regresses or changes. Every output
is byte-identical to the previously qualified main161 renderer candidate.
Missing HLine and inherited NumCases errors remain outside this renderer fix.

All 4,326 upstream input outputs are byte-unchanged. Of 5,758 published residual
inputs, four serialized cancellation outputs differ: two apparent fixes and two
apparent regressions. Both binaries emit both attribute orders across 64 repeats
per input, and all 512 complete parsed XML trees equal the originals. No
behavioral fix or regression is counted for those serialization differences.
Strict original SVG assertions perform no XML normalization. These standalone
counts do not include any fixes from the separately preserved HLine parser.

A recursive scan of 418 tracked testdata JSON files at actual main163 finds
eight existing complete references among these 376 input pairs; all eight
originals agree. The other 368 are first input publications. Full overlap
identities are in `empty_table_rules_inventory.json`. Counts are per-change
coverage, not a claim of globally unique additions across later parity changes.

The additional controlled renderer inventory injects authored table attributes
at the same pre-output boundary in both implementations. Its full identity
is sourceTeX plus display and tableAttributes, not just a public TeX pair.
It has 66 complete original SVGs, including four empty equal-row controls.
The first list-retention prototype exposed a separate source maximum error:
CommonMtable.getEqualRowHeight applies Math.max to all row totals, so an empty
list returns -Infinity, negative totals retain their negative maximum, and
NaN takes precedence over positive Infinity. The Go maximum now follows that
contract. Twenty direct original method observations cover empty and nonempty
lists, negative totals, signed zero, infinities and NaN in both orders.

All 66 controlled SVGs are exact after this coupled helper correction; 16
differ from actual main163 and 50 are unchanged. All 66 candidate outputs are
byte-identical to the previously qualified main161 renderer. The current-base
authored baseline is built using only the original main163 mtable source as an
overlay; its outputs also preserve the prior baseline controls. The four prior intermediate outputs
are preserved in `internal/svg/testdata/empty_table_equalrows_intermediate_proof.json`,
alongside the unmodified originals and exact corrected outputs. In particular,
no finite frame is accepted as equivalent to an original non-finite frame.
