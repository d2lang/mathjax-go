# Enclosure baseline and percentage-width measurement

`\fbox{^}`, `\fbox{~}`, and `\fbox{_}` previously drew frames with incorrect height or depth. These glyphs have a negative depth or height. The original enclosure combines its child into a zero-initialized bounding box, preserving the mathematical baseline before adding the frame's padding. Clearing the box to negative sentinel dimensions lost that baseline.

A clear D2 witness is `\Huge\fbox{^}\quad\fbox{~}\quad\fbox{_}`. The complete original SVGs, including the single-command forms and font/script/composition controls, are in the public reference fixture. The renderer change covers the shared enclosure measurement; it does not add the optional `\enclose` TeX package to D2's frozen configuration.

## Primary source and lifecycle

All references use D2's frozen MathJax 3.2.2 assets and source commit `ad8f5c21cb810236551da8c6512ba733e67357ee`:

- [CommonWrapper constructor, getBBox, setChildPWidths, invalidateBBox](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/output/common/Wrapper.ts): initial and temporary boxes start at zero; invalidation clears only the cached flag and retains dimensions. A temporary request on an already cached wrapper returns the owned box.
- [CommonMenclose.computeBBox](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/output/common/Wrappers/menclose.ts): combines into the supplied box without `empty()` or `clean()`, adds extenders, then resolves child percentage widths using each child's width.
- [CommonMpadded.computeBBox](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/output/common/Wrappers/mpadded.ts): assigns dimensions, then resolves child percentages against its authored width.

Both local width passes clear the percentage marker. The existing root math pass retains its `clear=false` behavior. This timing matters: resolving a table after its enclosure has already been cached can invalidate the enclosure and add padding again on its retained box. No generic cache reset or table remeasurement was added.

`enclosure_measurement_counterexamples.json` preserves both intermediate experiments: removing only `Empty/Clean` produced 11 real percentage-table geometry regressions plus two historical attribute-order differences among 13 byte regressions; restoring the Menclose pass alone left 40 explicit-width FrameBox composition regressions. The final candidate matches the original for all 51 genuine geometry counterexamples. The original outputs were never replaced with Go output.

## References and checks

Baseline: merged PR142, `7c8e7b53f64dc952c411d11b6030974f77e91402`.

- `enclosure_bbox_mathjax_3_2_2.json`: 2,452 complete original public SVG assertions, including 2,418 valid renderings and 34 error renderings. The baseline matches 1,752; the change fixes 700. Coverage includes negative child metrics, cancellation notations and parameters, box fonts/colors/scales/scripts, explicit FrameBox widths, nested fixed and percentage tables, and uncoupled mathmakebox/phantom/smash/raise/lower/overlap consumers.
- `enclosure_notations_mathjax_3_2_2.json`: 176 authored MathML fixtures. The observer parses a public cancel expression, sets only its `notation` attribute before the original SVG output pass, then uses the unchanged original renderer. These test box/circle/rounded-box/sides/radical/long-division/strike/arrow and combined notations, not public availability of an unloaded extension.
- `enclosure_lifecycle_mathjax_3_2_2.json`: 58 original method observations. The observer first saves an unmodified complete SVG, then deliberately requests cached and temporary measurements, invalidates, and requests saved measurements. Tests compare owned/result dimensions, scaling/spacing, cache flags, and result identity. Retained boxes can accumulate extenders on deliberate repeated saved recomputation, exactly as the source does; normal rendering avoids the erroneous late invalidation through the restored local width passes.
- `enclosure_bbox_residuals.json`: 18 raw inputs retained separately. Sixteen expose pre-existing nondeterministic serialization order of authored `data-padding` and `data-thickness`; two are unchanged inherited unsupported-package diagnostics. The sixteen remain raw even when an individual run happens to match. No output is normalized in any SVG assertion.
- `enclosure_inherited_font_controls.json`: preserves the two historical141 font-context residuals independently fixed by merged PR142. They are now exact controls and are not counted as enclosure fixes.
- `enclosure_attribute_order_observations.json`: repeated separate-process original/baseline/candidate observations bind the attribute-order explanation to concrete inputs. Geometry and paint are unchanged by those order variations.
- `enclosure_geometry_report.json`: affine glyph/paint/root comparison over all 2,470 public inputs reports 708 visual fixes, zero visual regressions among 1,760 previously visual-exact inputs, and zero changed visible residuals. The two inherited visible residuals remain unchanged. Eight additional visual fixes belong to the raw dual-attribute family and are excluded from the byte-exact assertion count.

A separate replay of all 3,022 published residual inputs found 30 newly exact originals, zero exact regressions, and zero changed nonexact results. Historical original fixtures remain untouched.

## Reproduction

Run `node --jitless testdata/generate_enclosure_bbox.cjs /path/to/pinned/assets`. It verifies all three asset SHA-256 values, creates a fresh original runtime per input, and bounds each subprocess to 24 runtimes. The lifecycle observer also verifies that the SVG saved before its controlled method calls equals the uninstrumented oracle output. Historical candidate/baseline receipts and attribute-order observations are not rewritten by generation.

Tests require exact complete SVG equality for the public and notation fixtures. Lifecycle dimensions use a tolerance of `1e-12` for cross-language floating arithmetic; all cache flags, ownership, percentage markers, and the original SVG remain exact.
