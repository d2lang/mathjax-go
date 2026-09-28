# Array entry operators and gathered alignment

The references use D2's original MathJax 3.2.2 bundle, source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. The comparison baseline is merged
main144, `a9fcab295ac962d46f533c5d3e4993a50f1e7bf4`.

## Source behavior

- `ParseUtil.fixInitialMO` skips leading mspace and empty TeXAtoms, then prepends
  an empty mi when the first remaining node is embellished or a relation
  TeXAtom. This preserves operator spacing at the beginning of an array entry.
- `BaseItems.EqnArrayItem.EndEntry` applies that operation to each authored
  entry after the first. `AmsItems.MultlineItem.EndEntry` applies it to entries
  after the first row. The Go call sites run before label cells, declared-count
  checks and generated Flalign layout cells; those are not authored entries.
- `MathtoolsMappings` passes `l` and `r` to AmsEqnArray for lgathered and
  rgathered. Their columns therefore use left and right alignment, rather than
  the centered policy of gathered. The existing literal `3pt` row spacing is
  unchanged; the extra mapping argument is ignored by the source handler.
- `AmsMethods.AmsEqnArray` reads an optional bracket argument before the body and
  calls `ParseUtil.setArrayAlign`. Its t/b/c forms become baseline 1, baseline
  -1 and axis. Other nonempty strings pass through, after the source JavaScript
  whitespace/control-space trimming. Aligned, gathered and both side-aligned
  variants now share the same already-tested helper as alignedat.

These corrections are coupled. Initial-operator repair alone widened rows with
incorrect centered alignment and moved dots farther from the original.
Alignment repair alone removed accidental compensation for missing operator
spacing. Missing optional-position handling rendered literal `[t]`, likewise
masking spacing errors. The separate source commits are reviewed together;
none of those intermediate outputs is a passing reference.

## Inventory and qualifications

`array_initial_mathjax_3_2_2.json` has 6,840 complete original SVGs: 6,196 valid
renders and 644 original error renders, compared byte for byte. Main144 fails
3,742. The deduplicated audit includes all 4,768 original held InitialMO cases,
550 gathered cases, 120 parser-entry compositions, 540 independent InitialMO
cases, 100 independent gathered cases and 1,274 optional-alignment cases.

The total audit has 7,280 unique inputs. `array_initial_residuals.json` preserves
all 440 unresolved original/baseline/candidate outcomes, without turning Go
output into a golden. Of these, 434 are byte-unchanged from baseline. The six
changed residuals are three multlined/tag formulas in both modes: math glyphs
now have exactly the original positions, while the pre-existing tag-height
error remains unchanged (17.5973 CSS px at the audit's 20px scale).

The unchanged residuals include existing Multline column/shove diagnostics,
original NEL runtime failures, nested environment capture, and CD font/operator
controls. Original runtime failures are retained as errors and are not counted
as rendered SVG references. The exact tests perform no SVG normalization.

All 520 previously changed nonexact InitialMO cases were reconsidered after
intervening prerequisites: 514 are now complete-original exact; the remaining
six are the tag controls above. Every changed valid family was rendered at the
same scale and matched by glyph identity, not DOM position. The literal `[t]`
regression is now exact. All 640 independent InitialMO/gathered references are
exact, including the seven independently quantified alignment-only regressions.

`array_initial_intermediate_observations.json` preserves 36 complete original,
baseline, isolated-candidate and final-candidate counterexamples. These show
why the changes must be composed; their intermediate outputs are not test
expectations. The old source branches and original audit receipts remain
available independently.

The replay of 3,138 previously published residual inputs fixes 334 exactly.
Fourteen changed outputs retain existing macro-generated row-break geometry;
all matched glyph positions are unchanged. The cancel controls also expose
pre-existing nondeterministic data-padding/data-thickness attribute order.
Both baseline and candidate repeatedly produce both variants. Their raw
observations are retained without normalizing the source references or calling
that unrelated serialization behavior an array regression.

## Regeneration

```sh
python3 testdata/generate_array_initial.py /path/to/pinned/D2/assets /path/to/node
```

The generator verifies all three asset SHA256 hashes, creates a fresh original
runtime for each input, and replaces only primary results. It never reads Go
output to generate expected SVGs. It regenerates the exact, unresolved and
intermediate-observation files, preserving their explicit inventories.

## Independent review and final validation

A separate optional-alignment review added 202 inputs beyond the 1,274-case
optional inventory: 160 complete-original exact, 126 fixes, no regressions and
no changed nonexact outputs. Its remaining 42 outcomes are byte-unchanged
ungrouped nested-environment capture failures. The source review confirmed
GetBrackets ordering, override priority and trimSpaces semantics.

All 7,280 baseline and candidate outputs, plus those 202 independent controls,
are unchanged after rebasing onto main144. All three original fixture files
regenerated byte-identically before the baseline metadata update. On main144,
the full frozen-oracle suite, race tests, vet and WebAssembly build all pass.
The exact test inventory remains 6,840; unresolved outcomes are never counted
as passing references.
