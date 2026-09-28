# Array column rules and partial frames

The references use D2's frozen MathJax 3.2.2 component at
`ad8f5c21cb810236551da8c6512ba733e67357ee`. Exact tests compare complete original
SVG strings. The separate raw inventory preserves all unresolved originals and
baseline/candidate observations; it does not normalize SVGs or make those
observations pass.

## Source behavior

[BaseMethods.Array](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts#L1385)
filters the column specification with a leading `c`, then replaces each
alignment letter followed by a run of separators with that run's final marker.
Surviving letters represent gaps with no rule. The first and last positions
are outer edges; the remaining positions become `columnlines` values. This
preserves gaps such as the second gap in `c|cc` and the final-marker behavior
of mixed `|:` runs. Ignored characters do not consume a column position.

[ArrayItem.createMml](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseItems.ts#L964)
sets an empty table frame and wraps partial left/right frames in `menclose`.
Its padding is zero only when the internal column-line string is nonempty and
is not exactly `none`. Partial frames use ordinary solid enclosure sides even
when authored with colons; the source dashed-frame flag applies to a full
four-edge frame. Smallmatrix properties remain on the inner table before this
wrapper is added.

The old Go code collected separator markers without their positions. An
`array{c|cc}` therefore repeated its rule over the following unmarked gap;
`array{|c:c|}` lost both outside rules and used a solid internal rule in place
of the dashed one. This change translates the positional separator logic and
returns the source enclosure through all three existing Array call sites.

The parser prerequisites already preserve an explicitly empty column alignment,
consume the next argument for literally empty optional alignment, and start an
Array's lexical environment correctly. This change retains those behaviors.
The shared UnderOver arithmetic prerequisite also resolves the independently
reproduced brace crop rounding difference. No renderer or row-ending source
changes are included in the column-rule commit.

## Complete inventory and remaining differences

The inventory has 1,252 distinct TeX/display inputs: 1,068 complete original SVG
assertions and 184 raw observations. Against the published PR157 head
`538d879f57839c1bfa2cd4f3c81c56d685c43109`, 828 inputs become exact and 240 remain
exact. All 1,068 assertions are valid original renderings. The raw inventory
retains 112 original-valid expressions and 72 original error renderings; there
are no original runtime exceptions in this inventory.

The 1,208-input source and independent inventory is preserved in full. Four
new no-rule controls explicitly bind the inherited unbraced-array script error.
Forty additional distinct inputs promote genuine fixes from the published raw
corpus; overlapping source and upstream receipts remain recorded in `origins`.
Coverage includes separator positions, mixed repeated rules, partial frames,
ignored characters, ragged and empty rows, nested arrays, all starred matrix
registrations, Physics matrix generators, fonts, scripts, pending recipients,
and decoration widths. The 144 optional-alignment intersections include 72
pairs whose direct and fallback argument forms have identical original SVGs.

Of the raw observations, 168 outputs are unchanged. These cover the separate
missing `hline`/`hdashline` handler and inherited unbraced-array script behavior.
Sixteen changed outputs are original errors: the original rejects an unbraced
Array BeginItem in a superscript or subscript, while Go already accepts it.
Each is linked by `inheritedControl` to one of the four unchanged no-rule
controls with a byte-identical original error SVG. Correcting lines after that
inherited acceptance does not resolve the script error. No original-valid
nonexact output changes in the final comparison. The complete original and
historical Go strings remain in the raw file.

The column change is checked against all 5,062 published raw inputs and 4,326
upstream test inputs rendered under the frozen D2 configuration. It produces
100 genuine exact fixes in the former and 12 in the latter. Cancel/enclose
attribute ordering can change serialized output without changing XML or
geometry: both binaries produced both orders over 48 repeats per affected
input. One apparent published fix and one apparent published regression, plus
two apparent upstream fixes, are excluded from the behavioral counts. No real
rendering regression or changed-valid unresolved output was found.

`array_column_rules_inventory.json` records overlap against every tracked JSON
under any `testdata` directory at the exact published PR157 head: 385 files.
Of the 1,068 exact inputs, 756 are first input publications, 212 already had
complete original references, and 100 promote prior raw originals. All
previous complete original strings agree. These are per-change coverage
counts, not globally unique additions to the project.

## Regeneration

```sh
node --jitless testdata/generate_array_column_rules.cjs PINNED_ASSETS
go test ./... -run TestArrayColumnRuleOriginalReferences
```

The generator verifies all three frozen asset SHA256 hashes and calls only the
original component, in a fresh runtime for each expression. It regenerates
both exact and raw original fields without invoking Go or selecting inputs by
their result. The regression test compares complete SVGs in both stored display
modes. Historical original fixtures and their expectations are unchanged.
