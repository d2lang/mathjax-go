# Physics brace quantities

The references come from the original D2 MathJax 3.2.2 bundle, pinned to
`ad8f5c21cb810236551da8c6512ba733e67357ee`. Exact tests compare the entire
unmodified original SVG. The separate residual file preserves original,
baseline, and candidate SVGs without normalizing attributes or geometry.

The source registration is
[`Bqty: ['Quantity', '\\{', '\\}', true]`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsMappings.ts).
[`PhysicsMethods.Quantity`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/physics/PhysicsMethods.ts#L72)
requires a braced operand and builds TeX using the registered delimiter
commands. Ordinary arguments use `\left\{ ... \right\}`; stars use fixed
`\{ ... \}`; the four literal size commands `big`, `Big`, `bigg`, and `Bigg`
select their corresponding left/right forms. A star still consumes an
optional size command, then uses the fixed fences.

The expansion runs in a new child parser, preserving the original copied
environment and shared configuration. Its inferred-row children are pushed
individually. This matters for a following script and for an unbraced script
operand: the first final item satisfies the script, while remaining items
return to the caller through the already implemented script-delivery path.

Unsupported control sequences use the existing source-defined Quantity
fallback. That branch creates empty fences from the literal registration
strings and restores the token cursor. It does not parse the delimiter
strings as TeX. Missing arguments and malformed input retain the original
error-rendering controls.

The inventory includes both display modes; all four sizes and stars;
fractions, roots, scripts, primes, pending functions and positions; font and
color scopes; matrices, CD, Eqnarray, and MathLap compositions; JavaScript
whitespace and cursor fallback; and operator/paired-delimiter overrides of
the command and expansion helpers. Existing Quantity registrations and
explicit literal expansions are retained as independent controls.

## Reproduce the original references

```sh
node --jitless testdata/generate_physics_brace_quantity.cjs /path/to/pinned-assets
```

The generator checks all three asset hashes and uses a fresh original VM
for every expression, in bounded subprocess batches. It regenerates the
original SVG fields in both exact and raw inventories. It never executes Go
or selects which references are asserted. The accompanying public-overlap
receipt counts distinct input/display pairs already published at the tested
baseline, rather than claiming that per-PR fixture totals are globally unique
MathJax coverage.

## Current baseline and residuals

Against main151 (`6ae444150070ff7dbe59174bbb1e7a49624bd16b`), the 1,698
distinct input/display pairs produce 1,626 exact references: 1,418 valid
expressions and 208 original error renderings. The change fixes 1,534 baseline
differences and preserves 92 exact controls, with no previously exact
regressions. None of these 1,626 complete original input references appeared
in the baseline's published JSON inventories; the overlap receipt records
the scan. The independent 212-input review and its six explicit controls are
included, rather than added again to these totals.

All 66 candidate-output changes caused by composing the intervening Eqnarray
and MathLap fixes become byte-exact originals. Every other output remains
byte-identical to the reviewed main149 candidate.

All 72 raw references have valid original renderings. Of these, 52 controls
are byte-unchanged from the current baseline. The 20 newly reachable
differences are qualified individually in
`physics_brace_quantity_qualification.json`:

- Sixteen expressions retain existing MathFont leakage into matrix cells.
  Each original equals its explicit literal expansion byte-for-byte, and
  each candidate equals the unchanged baseline literal output byte-for-byte.
- Two escaped line-feed expressions lose an existing nonbreaking space.
  The final three glyphs have the same 250-unit horizontal error as the
  unchanged `pqty` control, with identical glyphs and other affine values.
- Two escaped line-separator expressions retain an existing tokenization
  error. Their complete candidate error SVG equals the unchanged `pqty`
  control. They are retained as valid-original residuals, not error controls.

The published-residual replay covers 4,566 inputs and finds no source-output
changes. Four serialized observations change only the order of authored
cancel attributes. Both binaries emit both serializations over 48 repetitions
per input, and all complete parsed XML trees equal the original, including
every numeric geometry value. These are neither fixes nor source regressions;
the compact receipt retains the original strings and observations. No golden
SVG is normalized.

Validation runs the complete suite with the pinned original oracle enabled,
then `go test -race -p 1 ./...`, `go vet -p 1 ./...`, and
`GOOS=js GOARCH=wasm go build -p 1 ./...`. The exact and raw original files
also regenerate byte-identically from the three verified assets.
