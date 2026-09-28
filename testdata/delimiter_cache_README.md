# Delimiter measurement caches

The Go renderer invalidated measured child boxes whenever it selected a fixed
stretchable glyph. Original MathJax preserves those measurements: choosing a
fixed variant updates the glyph, while assembled stretching invalidates only
previously computed boxes along the ancestor chain. The extra invalidation
changed advances and placements in nested, empty stretchable fences.

For example, these now reproduce the complete original SVG:

```tex
\left[\Biggl[\Biggr]\right]
\left(\Bigl[\Bigr]\right)
```

This is separate from the preceding row-stretch correction, which determines
which outer children contribute to a row's stretch target. It restores the
subsequent cache policy and preserves all 1,020 merged row-stretch references.

The implementation follows pinned MathJax 3.2.2 source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`:

- [CommonMo.getStretchedVariant](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/output/common/Wrappers/mo.ts#L247)
  selects fixed and largest fallback variants without invalidation. Its
  assembled-stretch branch invalidates before computing the stretched box.
- [CommonWrapper.invalidateBBox](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/output/common/Wrapper.ts#L382)
  clears an already-computed box and recursively invalidates its parent.
  It does not clear children or propagate through an uncached box.

A diagnostic source overlay reproduced all 24 initial discrepancies among
96 fresh controls by applying Go's old invalidation behavior to the original.
Those overlays were used only to isolate the cause. Every reference below
comes from the untouched frozen runtime.

## Complete original references

`delimiter_cache_mathjax_3_2_2.json` contains **1,414 complete original SVGs**:
1,356 valid renderings and 58 original error renderings. It records the pinned
source, verified frozen D2 asset hashes, and merged main133 baseline
`a161cb04f8bd20d459604c283ffc9a4dcf190f2e`. The baseline fails 510 references.
The full 1,534-input corpus has zero formerly exact regressions.

Coverage includes fixed and assembled size boundaries, six fence families,
four explicit delimiter sizes, empty and ordinary content, nested fences,
embellished wrappers, scripts, accents, arrays, CD, Physics applications,
authored min/max sizes, font scales, spacing, padding and error controls.
Every previous row-stretch residual is included. Independent review added
518 fresh original comparisons: 502 exact, 164 newly exact, zero regressions;
its 16 residuals remain byte-identical to baseline. Rendering uses a fresh VM
per expression, font cache none, `em=16`, `ex=8` and the recorded display mode.

## Preserved raw observations

`delimiter_cache_residuals.json` preserves **120** complete original,
baseline and candidate responses that remain nonexact. None was exact at
baseline. In 98 cases the candidate response is byte-identical to baseline;
22 change but retain another difference. These are not passing references,
and no output normalization is used.

The remaining observations include authored scale/spacing and padding,
scripted embellished operators, Physics quantity fences and horizontal
brace measurements. The exact raw responses remain available for subsequent
source audits rather than being rewritten or removed.

All 11 changed expressions were inspected visually at the same scale and
compared structurally in both modes. Five quantity forms keep their prior
incorrect fence type, but their new measurements match ten additional exact
original parenthesis controls. Three scripted-token forms retain identical
glyph paths and move their widths toward the original while preserving a
separate sizing/depth difference. Three border-style forms now have the
original glyph placements and viewport; their remaining differences are
CSS declaration order and 0.1 SVG-unit border-corner rounding. No new visible
regression was found. The unchanged and changed residuals remain raw.

A further composition check renders all 1,656 distinct inputs in the 15
published residual inventories against merged main133 and this candidate.
Of 288 changed results, 266 become exact and the other 22 are precisely the
changed residuals inspected above. No formerly exact case regresses.

Regenerate both files with
`python3 testdata/generate_delimiter_cache.py PINNED_ASSETS NODE`. The generator
verifies every asset hash, uses batches of 24 fresh VMs, and updates only
original responses, preserving historical Go receipts. Run
`go test ./... -run TestDelimiterCacheReferences` for the full exact-SVG check.

The full frozen-oracle suite, race detector, `go vet`, and WebAssembly build
pass on the merged main132 base. The first full run exposed an older internal
assertion requiring invalidation for fixed-width zero. That assertion now
requires source cache retention for the fixed case while keeping assembled
widths four and six, warm/cold geometry, sibling and input-identity checks.
No original reference was changed to address the assertion.
