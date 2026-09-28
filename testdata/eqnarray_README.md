# Base and AMS equation-array environments

This change connects the original `eqnarray` and `eqnarray*` registrations to
`BaseMethods.EqnArray`'s existing Go implementation. Both environments are
taggable; only the unstarred registration requests default numbering. D2's
frozen `NoTags` configuration suppresses automatic equation numbers, while
explicit `\tag`, `\tag*`, labels and references keep their source behavior.

The source registration uses repeating `right center left` alignment with
alternating `0em 0.278em` column spacing. `BaseMethods.EqnArray` sets **3pt** row
spacing: the extra `.5em` argument in both source maps is unused. These
registrations do not consume an optional vertical-alignment argument.
The existing EqnArray implementation supplies subsequent-cell operator
spacing, row finalization, maximum-column repetition, tagging and the nested
equation-environment guard.

Primary source is MathJax 3.2.2 commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`:

- [BaseMappings.eqnarray](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMappings.ts#L706-L707)
- [AmsMappings.eqnarray*](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/ams/AmsMappings.ts#L104-L105)
- [BaseMethods.EqnArray](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts#L1487-L1508)
- [EqnArrayItem](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseItems.ts#L1105-L1180)

The visible witness is
`\begin{eqnarray}a&=&b\\c&=&d\end{eqnarray}`. Previously this produced an
unknown-environment error; it now matches the complete original SVG.

`eqnarray_mathjax_3_2_2.json` contains **1,374 complete original SVGs**:
1,132 valid representations and 242 original error controls. Baseline
`33ddd0e` (merged main131 plus the separate initial-operator correction)
fails 1,180 references; 194 remain exact controls. The 1,488-input exploration
has no formerly exact regressions or candidate panics. Both display modes
cover zero through eleven columns, repeating alignment/spacing, binary and
embellished operators, empty/interior/final rows, HFill-only entries, explicit
and missing tags/labels, fonts/styles/colors, nested arrays and guarded
equation environments, registered operator/paired-delimiter overrides, and
existing AMS/Matrix/CD controls.

`eqnarray_residuals.json` retains the other **114 raw original/baseline/candidate
results**: 64 original-valid representations and 50 error SVGs. They are not
passing goldens. The 22 existing-environment controls remain byte-identical
to baseline. The other 92 newly reachable residuals reproduce existing handler
differences. The report retains 90 analogous existing `align`, `align*`, or
plain-command inputs in `inheritedControls`; all already differ from original
MathJax and remain unchanged by this patch. Those raw controls (58 valid,
32 errors) are preserved in the same file. Two successful grouped-nesting
controls are already included in the passing regression inventory. Independent
review supplied 264 fresh original comparisons (256 exact, 196 fixed, zero
regressions); its 216 distinct inputs are included in these inventories.

The documented shared follow-ups are parser-owned row/cell boundaries
(`\cr`, `\newline`, row-spacing options, `\hline`, unbraced text arguments and
fences), MathFont inheritance into explicit tags, and existing closing/tag
argument diagnostics. The missing `\mspace` handler and ungrouped nested
EqnArray capture are also evidenced separately by unchanged existing
controls. Safe XML escaping remains intact. These source
handlers are separate from the two missing environment registrations.

Regenerate both the passing and raw residual inventories using only the
frozen original D2 runtime:

```sh
node --jitless testdata/generate_eqnarray.cjs /path/to/pinned/assets
```

The generator validates all three asset hashes, creates a fresh VM for each
input, and regenerates the inherited controls too. It neither executes Go
nor selects cases based on Go output. No reference is normalized or replaced
with a candidate SVG. The regression test compares complete SVG strings.
