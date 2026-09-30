# Display style in dcases and drcases

MathJax 3.2.2 registers `dcases` and `drcases` with `BaseMethods.Array` and style `D`. `BaseMethods.Array` writes `displaystyle: true` on that table. The Go ordinary-array handler already selected style `D`, but its cases-table completion unconditionally wrote `displaystyle: false`. The repair derives that existing assignment from `style == "D"`. It preserves the early direct-Pop return and the separate starred Cases handler.

This matters visibly: fractions use display size and sums place their limits above and below. The D2 examples are in `parity/dcases-display-style/` and `parity/drcases-display-style/`, with complete before, after, and frozen-original SVGs and screenshots.

## Original references and coverage

The three compressed containers preserve complete outputs from the frozen D2 MathJax 3.2.2 bundle at `ad8f5c21cb810236551da8c6512ba733e67357ee`:

- `dcases_displaystyle_mathjax_3_2_2.json.gz`: 500 strict full-SVG references, comprising 230 repairs and 270 exact controls. There are 468 valid SVGs and 32 complete error SVGs.
- `dcases_displaystyle_residuals.json.gz`: eight complete valid `displaylines` originals with the separately inherited missing explicit center-alignment gap. The test preserves and validates these originals without requiring the mismatch to persist.
- `dcases_displaystyle_observations.json.gz`: 144 existing inputs, each with its complete plain output, observed output, and ordered `ArrayItem.EndTable` snapshots. These are a subset of the public corpus, not additional inputs. This file is byte-identical to the original passive capture.

The full source union has 508 distinct `(tex, display)` inputs and 512 source memberships: 476 valid SVGs and 32 error SVGs, with no original runtime exceptions or unsafe XML. Its original 504 rows remain unchanged. Four additional already-published `dcases` inputs cover empty fractions and empty superscripts in both modes; they were discovered as real improvements in the independent broad replay and were freshly reproduced from the original.

Coverage includes fractions, infix `over`, sums, integrals, binomials, limits, local style overrides, fonts, fences, scripts, numerators, nested tables, ordinary/starred sibling environments, malformed calls, and all 13 live registrations carrying a `D` argument. The passive trace observes source table definitions without changing output.

A fresh scan of all 523 tracked JSON/JSON.gz files at actual merged PR180 `5985de47cd782beeee58344eb1ddaa0bbd5cc51e` found no original conflicts or ambiguous display modes. Of the 500 matched inputs, **488 are first full-SVG assertions, eight promote preserved raw originals, and four already have strict assertions**. The eight residual displaylines inputs are first source observations only. Four promoted originals were in the ordinary-array residual fixture; the other four were in `array_rows_residuals.json`, which had no Go test consumer. Repeated metadata/source occurrences are counted once per input.

## Corrected historical baseline accounting

The initial 504-case Go capture sent only nested `options.Display`. Its bound probe read the top-level `display` field, so requests intended for display mode were rendered inline. The original JavaScript oracle correctly read nested Display, and all original SVGs remain unchanged. Independent reproduction retained and reproduced all historical results, then replayed both Go binaries with explicit top-level mode.

Sixteen display-mode numerator outputs differ under the corrected wire. Fourteen remain real style fixes; the two plain-cell numerator cases already matched when the baseline was rendered in the correct mode. Thus the initial apparent 228 fixes/268 controls become **226 genuine fixes/270 controls** for the unchanged 504 inputs. Adding the four independently found raw-original cases gives the final **230 fixes/270 controls**. The historical capture and its 16 differences remain retained as audit evidence; no original or historical response was rewritten.

## Regression qualification

The committed source at `b25453663d1e4b6f12be0943d367bbb2aab03ae8` was compared with the production-equivalent merged-base binary using full API objects. All eight held focused outputs remain unchanged. The protected replay retains 39,071 complete source observations (38,223 unique inputs), plus 284 published SVG-hash observations (282 unique inputs) across 51 fixture files. Twelve seeded method-profile records are retained separately and excluded from public rendering. Protected results have eight strict improvements, including the four focused raw promotions and the four additional inputs above, with no regressions or runtime changes. Hash controls are unchanged; their assertions concern SVG hashes, not a replay of their retained AST fields.

The four separate historical broad suites contain 17,549 observations. Their four genuine improvements are the same additional protected inputs, not four more unique cases. Three changed cancellation strings and two historical baseline differences are existing attribute-order variation: independent repeated runs proved identical complete parsed SVG trees and both orders in both binaries. Raw outputs remain retained. Source-invalid XML and original runtime exceptions stay separate from strict SVG parity; no normalization was applied to the focused references.

## Regeneration and validation

Run the original-only generator with the three pinned asset files and a new output directory:

```sh
node --jitless testdata/generate_dcases_displaystyle.cjs /path/to/pinned-assets /path/to/new-output
```

The generator verifies asset hashes, creates fresh VM state for every conversion, reproduces all 508 complete originals including the eight held outputs, and repeats both plain and instrumented conversions for all 144 passive inputs. It compares complete event/output objects and regenerates all three gzip containers with the same deterministic serializer used for export. No Go renderer provides expected SVGs.

`dcases_displaystyle_inventory.json` records source, overlap, qualification, and file hashes. Original-only regeneration passed: all 508 complete originals and 288 plain/observed passive conversions were reproduced in 796 fresh VM contexts, and all three gzip containers were byte-identical. Independent static review also verified every original, source membership, publication category, generator, and test. Formatting and Go test execution, full gates, final D2 publication binding, and publication remain pending with the parent task; this assembly receipt does not claim those checks have run.
