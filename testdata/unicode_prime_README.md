# ASCII and U+2019 prime runs

D061 is based on accepted D060 merge `3dcb09030b4404c4c5982a9cebfbe5c1b6e51cd9`. Pinned MathJax 3.2.2 BaseMappings dispatches exactly ASCII apostrophe and U+2019 RIGHT SINGLE QUOTATION MARK to Prime. Its Prime handler collects either spelling with GetNext, then pushes the same pending PrimeItem. The implementation changes only entry and run recognition/UTF-8 advancement; D060 ownership, finalization, scripts, fonts and errors remain unchanged.

`generate_unicode_prime.cjs` checks the three original D2 v0.8.1 asset hashes and captures118 complete SVGs plus explicit and own-property trees, with a fresh unmodified primary VM for every case. All16 required primary errors are preserved. Of these,106 SVGs and explicit trees match raw primary. Twelve literal/non-prime controls retain exact accepted-parent SVG and complete explicit/own trees: mbox, operatorname, U+2018, U+2032, U+2033 and en dash, each in both modes. Their differences are explicitly stored in `unicode_prime_boundaries.json`; none is claimed fixed. Ordinary text and explicit mtext literal controls match raw primary. The same file records86 exact node-path/value/text qualifications for the inherited missing pseudoscript property; all other own properties remain strict. Twenty of106 strict cases have raw own-property equality. The public tests also require deterministic repeated rendering, primary display measurements and no escaped transient limits-origin state.

`generate_unicode_prime_collector.cjs` directly calls the registered Prime handler and actual TexParser GetNext/nextIsSpace/getCodePoint methods. Its40 records cover both spellings, mixed runs, lengths1–5, BMP/astral following characters, ordinary whitespace and stop boundaries, plus required occupied-superscript rejection. Tests bind original base identity/unchanged rejection, token text and attributes/properties, remaining text and separate UTF-16/Go byte cursor units. The existing Go-only factory provenance marker is asserted true and removed solely for this private method comparison, as in D060; public compiler trees contain no such marker.

The collector now follows the pinned JavaScript whitespace set, including BOM and excluding U+0085/NEL. All40 records are strict primary collector comparisons, including rejection before whitespace lookahead. The separate prime-whitespace fixture covers every source whitespace character; its README distinguishes successful public references from the unmodified primary's NEL TypeErrors. Unbraced `x^'`, `x^’`, `x_'` and `x_’` remain unchanged literal-script behavior in Go while primary reports missing-open-brace errors; they are separately reported, not silently fixed or treated as parity.

The original96 pending-prime primary fixture is byte-identical; only its four D061 qualifications are removed in favor of raw SVG/explicit AST equality. The other92 raw compiler trees are unchanged. Existing66 lifetime/font controls,162 private ownership cases,48 annotation,50 stack,72 stackrel and98 limits references remain unchanged and must pass. No generic Unicode replacement, parser argument rewrite, renderer change, or D062–D064 correction is included.

D103 composition retains both original primary and historical boundary JSON files
byte for byte. The current 118-case test has 108 primary-SVG references, eight
unchanged literal/non-prime residuals, and exactly two composed D106 diagnostics
(`operator-literal-inline` and `operator-literal-display`). The latter retain the
exact old nonprimary SVG and literal U+2019 structure. Only the inner normal mi(x)
class is corrected: its complete explicit/own subtree equals the untouched
primary msup-base mi(x), with no own texClass. A comparison copy of the immutable
old boundary changes only that one source- and primary-validated property; every
other field remains strict. Neither whole candidate trees nor normalized SVGs
become expected outputs, and the separate full child-parser repair remains D106.
