# Explicit MathML spacing oracle

`explicit_spacing_mathjax_3_2_2.json` records complete SVG hashes and numeric
operator defaults from D2's frozen MathJax 3.2.2 component. The generator checks
all three asset hashes documented in the repository's `PROVENANCE.md`, creates
a fresh runtime per case, overlays attributes on compiled MathML, and then
invokes the unmodified SVG output jax. No JavaScript asset is shipped here.

Regenerate with:

```sh
node internal/svg/testdata/generate_explicit_spacing.cjs /path/to/d2latex-assets
```

An optional second argument saves the actual oracle SVG files for inspection.
The normal Go tests require neither Node nor the external assets.

The 46 cases cover both display modes for explicit/inherited/zero/negative
spacing, missing dictionary defaults, adjacent embellished operators, scripts,
top embellished ownership, isolated/fixed-arity contexts, and ordinary controls.
Two further display cases prepend a pure-MathML empty operator before a sum,
with and without zero spacing on the sum's actual core `mo`. This does not
implement or assume support for the separate TeX `mmlToken` command.

The empty-before-sum cases use display mode because the current Go input stage
uses an `msubsup` node for the inline source where the pinned input stage retains
`munderover`; the resulting positions match, but that unrelated serialized node
kind is outside this renderer-spacing test. Unicode range defaults are tested
directly in `internal/tex/operator_spacing_test.go`, avoiding unrelated fallback
glyph/parser representation differences. Exact SVG comparisons do not normalize
these differences away.

The implementation follows pinned `CommonWrapper.getSpace/getMathMLSpacing`
(`ts/output/common/Wrapper.ts:526–570`), `MmlMo.hasSpacingAttributes/coreParent`
(`ts/core/MmlTree/MmlNodes/mo.ts:208–245`) and dictionary numeric defaults
(`mo.ts:364–389`). Explicit values remain distinct from those numeric defaults;
only missing sides use the script-level clamp, and adjacent embellished siblings
share the larger adjoining gap. No public mathml-spacing option is added.
