# Authored font-family references

These 78 complete SVG references use the unmodified D2 v0.8.1 MathJax 3.2.2
assets, checked against exact hashes in the generator, in a fresh VM for each
formula and display mode. No output normalization is applied.

The family-selection contract follows CommonWrapper.getVariant and
explicitVariant in MathJax-src ad8f5c21cb810236551da8c6512ba733e67357ee.
MathML font attributes override CSS font declarations. An explicit mathvariant
wins over both; numeric weights above 600 select bold. Font family selects
explicit text and the reference adaptor's corresponding text metrics, without
changing the original MathML. CommonMo's large-operator and later stretched
variants remain authoritative.

Cases cover all five token kinds, CSS/attribute precedence, variant overrides,
normal/bold/italic and numeric weights, empty values, font lists, CJK and
supplementary text, scripts and scaling, string quotes, padding, large operators
and fixed/assembled delimiters. Ordinary math and genuine error output remain
controls. Renderer tests also retain input node/parent/attribute identity and
require repeated and cloned rendering to produce the same complete SVG.

This is the pinned renderer/adaptor contract, not a promise that an arbitrary
font is installed in a browser or native host. Other unsupported font-only
style behavior is outside this family-selection correction.

Regenerate with:

```sh
node testdata/generate_authored_font_family.cjs /path/to/d2latex /tmp/authored-font-svg
```
