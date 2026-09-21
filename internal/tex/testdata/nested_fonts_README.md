# Nested font precedence references

The 56 cases use the unmodified MathJax 3.2.2 component embedded in D2 v0.8.1,
with three exact asset hashes and a fresh VM for each formula/display mode.
The renderer hook observes the actual compiled tree immediately before the
unchanged output method. Every case compares the complete kind/text/explicit
attribute/child tree and complete SVG bytes, without normalization. The Go
inferred-mrow representation is mapped to the corresponding primary node name.

Both bold/roman nesting directions, three levels, multiple-letter arguments,
fractions, scripts, radicals, surrounding unstyled text, scoped declarations,
sequential declarations, raw operators and explicit kept mmlToken styles are
covered. Ordinary expression/font/function controls remain unchanged. A separate
pure test verifies resolved inner choices survive the parser's existing node
clones while siblings receive the outer choice. Mathit, sans/monospace nesting and scoped italic controls retain the original
italic precedence witness after D040. Every compiled reference also
requires temporary font-scope bookkeeping to be absent from returned MathML.

The implementation preserves the existing parser's postorder font lowering but
protects tokens already resolved by an inner MathFont or SetFont scope. It does
not alter D040's separately corrected font names or claim universal ambient-font
parity. D041 tracks the preexisting incorrect font application to fixed symbols
and authored tokens without an explicit variant; its separate minimal original
and candidate outputs are retained in the audit. The original exploratory
48-case matrix, including those policy failures, remains historical evidence.

Regenerate with:

```sh
node internal/tex/testdata/generate_nested_fonts.cjs /path/to/d2latex /tmp/nested-font-svg
```
