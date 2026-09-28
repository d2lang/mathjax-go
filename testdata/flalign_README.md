# Flalign / XalignAt rows and percentage-width lifecycle

These references come from D2's original MathJax 3.2.2 bundle at source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. The comparison baseline is merged
MathJax-Go main137, `6f14f09f1e15ffb5a4e94cca0eb6bfa4a790a139`.

The source policies are:

- `ts/input/tex/ams/AmsMethods.ts`: `XalignAt` retains twice the validated pair
  count; `FlalignArray` selects padding, centering and label behavior.
- `ts/input/tex/ams/AmsItems.ts`, `FlalignItem`: `EndEntry` rejects excess authored
  entries immediately; `EndRow` pads the declared entries and inserts fit cells
  between equation pairs, with exterior fit cells only for padded XalignAt.
  `EndTable` repeats column definitions and makes centered single-pair Flalign
  natural width. XalignAt labels have zero width and negative-width lspace.
- `ts/output/common/Wrappers/mrow.ts`: row stretching precedes propagation of
  child percentage markers.
- `ts/output/common/Wrapper.ts`, `ts/output/svg/Wrappers/math.ts` and
  `ts/output/common/Wrappers/mtable.ts`: the root percentage pass follows natural
  bbox computation, supplies the output container width, preserves markers
  through transparent wrappers, and lets the table clear its own/container
  markers and invalidate cached ancestors before one root recomputation.

The renderer corrections and table parser correction are coupled. Fixing only
column positions exposed inherited responsive-root scaling inside fences; fixing
only scaling retained incorrect column placement and natural widths. Neither
intermediate version is claimed as a complete fix.

## References and comparison policy

`flalign_mathjax_3_2_2.json` contains 3,005 complete original SVGs: 2,931 compare
byte-for-byte, and 74 overflow diagnostics compare with one precisely bound XML
attribute escape. The original serializer emits a bare `&` inside
`data-mjx-error="Extra & in row of ENV"`. The test changes only that exact
attribute to `&amp;`, preserves the original reference verbatim, requires valid
output XML, and separately checks the parser's original `XalignOverflow` ID and
literal message for all 74 inputs. No other XML or SVG normalization is allowed.
Main137 fails 1,412 of these checks (1,338 newly byte-exact and 74 diagnostics).

`flalign_nested_mathjax_3_2_2.json` contains 432 additional complete original
references; all compare byte-for-byte and main137 fails 180. These include
public nested-table inputs and source output controls applying listed table
attributes in preorder before typesetting. They cover natural/fixed/percentage
outer and inner widths, fit/percentage columns, equal columns/rows, labels,
fixed and stretchy fences, fractions, scripts and accents. The same attributes
are applied through the original post-filter and the Go shared MathML tree.

The broader inventories retain 207 unresolved cases verbatim:

- `flalign_residuals.json`: 199 complete original/baseline/candidate triples,
  including existing initial-operator spacing, explicit row gaps, row separator
  commands, Mathtools command/context/override handling and a fraction/table
  rounding difference.
- `flalign_nested_output_observations.json`: eight nested percentage-table/overbrace
  controls. These carry explicit MathML attribute inputs, so they are replayed
  by the output-control generator rather than a public-TeX-only residual runner.
  The brace width and tag placement improve. Existing nested 150%
  placement differences remain (about 5.9 CSS pixels in the sampled input),
  together with subpixel position differences up to 0.17 CSS pixel at the
  audit's 20px font size. They are not passing references.

All 63 changed original-valid residuals in the main inventory were rendered in
Chromium at the same scale. Their glyph paths are unchanged and no individual
bounding-box component moves farther from the original by more than 0.01 CSS
pixel. The eight additional nested controls were inspected separately; their
raw differences remain recorded. This visual check is diagnostic evidence, not
an SVG comparison normalization.

A replay of all 2,088 previously published residual inputs fixes 142 exactly;
the other 16 changed outputs are the already recorded initial-operator spacing
and incorrectly accepted Aboxed cases. An independent 260-input review finds
227 exact outputs, 141 fixes and no formerly exact regression. Its remaining
caller/override controls retain their original evidence.

## Regeneration

Run from the repository root:

```sh
node --jitless testdata/generate_flalign.cjs /path/to/pinned/D2/assets
```

The generator verifies all three frozen asset hashes, creates a fresh original
VM for each input, and updates only primary SVG results. It never reads Go
output to generate expectations. All four fixture files regenerated
byte-identically. The historical primary fixtures are unchanged.

The final candidate is rebased onto merged main138,
`aaa365ec3a653db84b9f625013a2e2350f76cd4d`. All 3,204 public audit outputs
are unchanged across that rebase on both the baseline and candidate. Full
frozen-oracle tests (including the new BuildRel and 432 nested-width references),
race tests, vet and the WebAssembly build pass on that combined source.
Replaying all 2,142 published residual inputs gives the same 142 new exact
outputs and 16 already-reviewed changed residuals.

The two existing isolated table geometry tests now explicitly invoke the parent
percentage pass before asserting the same final numeric dimensions. Their old
assumption that table construction itself resolved percentages was precisely the
lifecycle bug; the assertions were preserved.

Validation passed on the combined main137 candidate: full frozen-original suite,
targeted checks for the appended nested references and original diagnostic
messages, race tests, vet, and the WebAssembly build. The early targeted gate
failed only on the isolated tests' obsolete eager-resolution timing; their final
numeric assertions now pass after the source parent-width invocation.
