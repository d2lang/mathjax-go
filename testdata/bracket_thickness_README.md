# Mathtools bracket thickness dimensions

`UnderOverBracket` in the pinned MathJax 3.2.2 source converts the optional
thickness with `length2em(thickness, .1)` and formats it with `em`. The Go
handler previously accepted only em and unitless numbers, silently ignoring
point and other units. Unitless numbers also use .1 as their reference size.

The original SVG fixture covers both bracket directions and display modes,
all supported unit families, named spaces, empty/unknown values, permissive
prefix parsing, zero and negative lengths, Unicode whitespace, and rounding.
All 116 cases are complete SVG output captured from the unmodified D2 bundle.
The generator verifies its three asset hashes and uses a fresh runtime per case.

Regenerate with:

```sh
NODE_OPTIONS=--jitless node testdata/generate_bracket_thickness.cjs PINNED_ASSETS
```

The shared length parser now recognizes the same leading whitespace as
JavaScript. Border shorthand preserves its authored form until individual
components change, as `Styles.splitWSC` does; this matters for zero and negative
bracket widths, which are not recognized as ordinary CSS width tokens.
