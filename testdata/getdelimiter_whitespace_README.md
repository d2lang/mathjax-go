# Original delimiter whitespace references

The frozen MathJax 3.2.2 `GetDelimiter` method calls `GetNext`, whose `nextIsSpace` uses JavaScript `\s`. When a fixed-size delimiter is braced, it calls `GetArgument(...).trim()`. Both include U+FEFF BOM and exclude U+0085 NEL. Go used `unicode.IsSpace` at the leading boundary and `strings.TrimSpace` for the braced delimiter. The bounded fix uses the existing exact JavaScript whitespace predicate at those two locations.

The fixture contains complete, unmodified original APIs from a fresh runtime per input. Its 1,288 SVG assertions comprise 902 valid formulas and 386 rendered diagnostics. Against main `ddd8efd95eb38c9eae53dbf42262c217af654658`, 150 valid BOM cases and 114 diagnostic cases differ; the other 1,024 are controls. All 1,288 match after the fix. Two original TypeErrors are separately retained with their full stacks and excluded from SVG parity credit; Go's specific `no Unicode range for character U+0085` runtime error is tested separately.

Coverage includes all 16 fixed-size delimiter commands, raw/command/braced delimiters, left/right/middle and Mathtools mleft/mright, the complete JavaScript whitespace set, NEL/MVS/ZWSP controls, and nested styles/scripts/text/arrays in both display modes. The source fixture also captures the frozen original method bodies. The D2 witness contains actual U+FEFF immediately after `\left` and `\right`; its escaped spelling is `F(x) = \\left\uFEFF(\\frac{a}{b}\\right\uFEFF)`.

Regenerate only from hash-verified original assets:

```sh
python3 testdata/generate_getdelimiter_whitespace.py /absolute/path/to/pinned-assets /absolute/path/to/node
```

The generator does not read Go-generated SVGs. These finite checks do not establish universal MathJax parity.
