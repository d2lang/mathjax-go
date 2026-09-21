# Wide accent references

These 52 complete SVG references and display dimensions come from unmodified D2 v0.8.1 MathJax 3.2.2 assets, SHA-checked by `generate_wide_accents.cjs`. Run that script with Node, the pinned d2latex directory and an optional evidence directory. Each expression/mode uses a fresh VM. All serialized SVG bytes and dimensions are compared without normalization or tolerance.

The matrix covers fixed-size boundaries, long/very-long hats and tildes, inner font scopes, fractions, scripts, nested accents and unchanged ordinary accents/arrows. Thirty-two cases fail on the pre-fix implementation. The largest fixed delimiter must retain its alias character, and accent centering must include the prototype box's italic correction, as CommonMo.getStretchedVariant/protoBBox/getAccentOffset specify. The unrelated one-unit vector accent position is tracked separately as Studio D047.
