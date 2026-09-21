# Math italic font references

These 38 complete SVG references come from the unmodified MathJax 3.2.2
component embedded in D2 v0.8.1. The generator verifies all three asset hashes,
uses a fresh VM for each case, and renders both display modes with em16/ex8.
Display measurements use the serialized dimensions multiplied by8 and rounded
up, matching the public Measure contract. No SVG normalization is applied.

Regenerate independently:

```sh
node testdata/generate_mathit.cjs /path/to/d2latex /tmp/mathit-reference
```

The pinned BaseMappings maps both MathFont `mathit` and SetFont `it` to
`TexConstant.Variant.MATHITALIC` (`-tex-mathit`). The generated source tables
already retain those values; these tests cover the public command dispatch.
The previous ordinary italic mapping uses different Latin glyphs and widths.

Cases cover single/multiple Latin characters, digits, Greek, scripts, fractions,
radicals, operators, empty/unbraced arguments and following unstyled text.
The declaration case also corrects the scoped `it` mapping. Plain expressions,
bold, roman, sans-serif and the sine operator remain unchanged controls.
The baseline fails22 cases, and the corrected mapping matches all38.

Nested font-command precedence is a separate parser concern. This correction
changes only the two variant names; it does not alter font-scoping algorithms,
input MathML, renderer metrics or font data.
