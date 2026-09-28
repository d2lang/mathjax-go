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

The current inventory has **2,154 distinct inputs** against merged main149
`e3d4006b093073ad46555f92e94df4c85dd46331`. **1,990 complete original SVGs**
are exact (1,658 valid representations and 332 original error controls):
1,548 newly exact inputs and 442 unchanged controls, with no formerly exact
regression. Of these inputs, **1,912 original references are being published
for the first time**; 20 promote published HFill residuals and 58 repeat
published shared-handler controls. `eqnarray_current_public_overlap.json`
binds all 78 existing inputs to their unchanged original SVGs and files.
The historical Eqnarray inventories below were held locally and never
published, but some common controls also appeared in other published suites.
Both display modes cover zero through eleven columns, repeating
alignment/spacing, binary and embellished operators, empty/interior/final
rows, HFill-only entries, explicit and missing tags/labels, fonts/styles/colors,
nested arrays and guarded equation environments, registered operator/paired
overrides, and existing AMS/Matrix/CD controls.

`eqnarray_mathjax_3_2_2.json` preserves the initial **1,374 original SVGs**
unchanged. Its historical baseline `33ddd0e` was main131 plus a separate
initial-operator prototype. That prototype is not part of this change: the
final implementation builds on the merged array and Nonscript fixes.
`eqnarray_current_mathjax_3_2_2.json` adds **616 distinct exact assertions**:
532 newly captured original SVGs and 84 promoted existing references (34
historical Eqnarray residuals, 30 historical inherited controls, and 20
published HFill/Eqnarray residuals). These include fresh
Arrow/Aboxed/Cramped/FrameBox/BuildRel/Pmb/Skew/Nonscript compositions, tag and
nesting controls. The fresh capture has 556 distinct inputs in total: those
532 exact originals and 24 raw residual/analogue controls. This includes 192
ordinary/prime script-delivery intersections across Eqnarray, Align and Matrix:
184 exact originals and eight inherited original-error cases.
Independent historical review supplied 264 fresh comparisons; its 216 distinct
inputs remain included without altering the originals.

The historical `eqnarray_residuals.json` also remains unchanged, including its
114 cases and 90 inherited-handler controls. It is a historical receipt, not
a statement that every case still fails. `eqnarray_current_residuals.json`
records the remaining **164 raw results** against main149: 68 original-valid
inputs and 96 original error renderings. None is a passing golden. Of these,
92 are byte-unchanged existing-handler controls; 72 are newly reachable
Eqnarray inputs, comprising 32 valid representations and 40 error renderings.

`eqnarray_current_qualification.json` binds all 72 newly reachable residuals
to unchanged existing `align`/`align*` inputs. Every valid residual is also
byte-identical to the held historical Eqnarray candidate. The 20 valid inputs
that still produce errors have exactly the same error SVG as the existing
control. The other 12 retain the same glyph shapes and paint, with precisely
the same per-glyph vertical displacement from original as the existing
control. These are shared row-spacing options or macro-generated row breaks.
The four additional Eqnarray script cases are original errors that the shared
script parser incorrectly accepts, as do the unchanged Align controls; no
valid representation is qualified by that exception. All 1,962 earlier
Eqnarray candidate outputs remain byte-identical after the main149 rebase.

The remaining shared follow-ups are `\hline`, `\cr`/`\newline`, row-spacing
options, macro-produced row boundaries, environment-close/tag-argument
diagnostics, unbraced named-function script acceptance, and safe XML ampersand
escaping. Original bare ampersands in
error attributes remain untouched in the receipts; Go continues to escape
them. Earlier missing `\mspace`, nested-array, tag-font and initial-operator
cases that now match are included in exact assertions rather than waived.

The complete replay of **4,306 published residual inputs** finds **20 genuine
newly exact cases and no regressions or changed nonexact representations**.
Three apparent cancellation fixes and two apparent cancellation regressions
were attribute-order variation only. The qualification records 48 samples per
binary and input: both binaries emit both orders, and every complete parsed
XML tree equals the original. Exact golden assertions perform no such
normalization.

Regenerate the historical and current passing/raw inventories using only the
frozen original D2 runtime:

```sh
node --jitless testdata/generate_eqnarray.cjs /path/to/pinned/assets
```

The generator validates all three asset hashes, creates a fresh VM for each
input, and regenerates the inherited controls too. It neither executes Go
nor selects cases based on Go output. No reference is normalized or replaced
with a candidate SVG. The regression test compares complete SVG strings.
