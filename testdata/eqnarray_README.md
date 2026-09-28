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

The current inventory has **1,962 distinct inputs** against merged main148
`6f388b9618a2347151b4120e58c938811104d567`. **1,806 complete original SVGs**
are exact (1,482 valid representations and 324 original error controls):
1,456 newly exact inputs and 350 unchanged controls, with no formerly exact
regression. Of the 1,806 exact references, **1,786 are being published for the
first time**; the other 20 promote existing published HFill residuals. The
historical Eqnarray inventories below were held locally and never published.
Both display modes cover zero through eleven columns, repeating
alignment/spacing, binary and embellished operators, empty/interior/final
rows, HFill-only entries, explicit and missing tags/labels, fonts/styles/colors,
nested arrays and guarded equation environments, registered operator/paired
overrides, and existing AMS/Matrix/CD controls.

`eqnarray_mathjax_3_2_2.json` preserves the initial **1,374 original SVGs**
unchanged. Its historical baseline `33ddd0e` was main131 plus a separate
initial-operator prototype. That prototype is not part of this change: the
final implementation builds on the merged array and Nonscript fixes.
`eqnarray_current_mathjax_3_2_2.json` adds **432 distinct exact assertions**:
348 newly captured original SVGs and 84 promoted existing references (34
historical Eqnarray residuals, 30 historical inherited controls, and 20
published HFill/Eqnarray residuals). These include fresh
Arrow/Aboxed/Cramped/FrameBox/BuildRel/Pmb/Skew/Nonscript compositions, tag and
nesting controls. The fresh capture has 364 distinct inputs in total: those
348 exact originals and 16 raw residual/analogue controls.
Independent historical review supplied 264 fresh comparisons; its 216 distinct
inputs remain included without altering the originals.

The historical `eqnarray_residuals.json` also remains unchanged, including its
114 cases and 90 inherited-handler controls. It is a historical receipt, not
a statement that every case still fails. `eqnarray_current_residuals.json`
records the remaining **156 raw results** against main148: 68 original-valid
inputs and 88 original error renderings. None is a passing golden. Of these,
88 are byte-unchanged existing-handler controls; 68 are newly reachable
Eqnarray inputs, comprising 32 valid representations and 36 error renderings.

`eqnarray_current_qualification.json` binds all 68 newly reachable residuals
to unchanged existing `align`/`align*` inputs. Every valid residual is also
byte-identical to the held historical Eqnarray candidate. The 20 valid inputs
that still produce errors have exactly the same error SVG as the existing
control. The other 12 retain the same glyph shapes and paint, with precisely
the same per-glyph vertical displacement from original as the existing
control. These are shared row-spacing options or macro-generated row breaks,
not a change to the environment's column geometry.

The remaining shared follow-ups are `\hline`, `\cr`/`\newline`, row-spacing
options, macro-produced row boundaries, environment-close/tag-argument
diagnostics, and safe XML ampersand escaping. Original bare ampersands in
error attributes remain untouched in the receipts; Go continues to escape
them. Earlier missing `\mspace`, nested-array, tag-font and initial-operator
cases that now match are included in exact assertions rather than waived.

The complete replay of **4,271 published residual inputs** finds **20 genuine
newly exact cases and no regressions or changed nonexact representations**.
One apparent cancellation fix and one apparent cancellation regression were
attribute-order variation only. The qualification records 48 samples per
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
