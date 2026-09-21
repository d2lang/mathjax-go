# Token font policy references

The 74 references come from the unmodified MathJax3.2.2 component embedded in
D2 v0.8.1, with three checked asset hashes and a fresh VM per formula and mode.
The generator observes the complete actual compiled tree before the unchanged
SVG output method. No parser/MML overlay or SVG normalization is used.

Every case compares the ordered token kinds and explicit font variants, and
forbids both internal font-provenance properties from escaping compilation.
Sixty cases additionally match both complete AST and complete SVG; the other
14 retain explicitly named, preexisting shape/renderer qualifications in the
test. Fourteen qualified cases still enforce every token font choice; they are
not claimed as complete rendering parity. Unicode fallback glyph remapping is
tracked separately as D042. Accent/prime AST shape and vector/overline rendering
remain separately recorded, without changes to their primary expectations.

Coverage includes ambient ASCII letters/digits/operators and family Gamma;
fixed Greek/normal/AMS symbols, functions/operators/delimiters, raw package
colon and closing characters; authored mi/mn/mo with absent or kept variants;
Unicode range override priority; and declaration, argument, group and nested
scope boundaries. All 56 separately accepted nested-font references remain
unchanged and pass. The clone test now explicitly creates ambient-eligible
inputs, preserving its same inner/outer/kept-attribute assertions.

D043 Physics VectorBold scope failures remain separately frozen as unchanged
baseline/candidate negative evidence, outside this ambient-token correction.
The unsupported old-oracle boldsymbol extension retains its existing whole-token
scope. Ten exact accepted-Go AST/SVG references in the separate extensions
fixture guard this compatibility boundary; they are not an old-MathJax oracle
or a general nested-extension correctness claim.
No generic vector/text transform, font data, symbol table or renderer is changed.

Regenerate with:

```sh
node internal/tex/testdata/generate_font_token_policy.cjs /path/to/d2latex /tmp/font-token-svg
```
