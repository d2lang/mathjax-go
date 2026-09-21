# Frozen mmlToken parser oracle

`mml_token_mathjax_3_2_2.json` contains 62 parser cases captured from unmodified
D2 v0.8.1 MathJax 3.2.2, commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. The generator verifies all three D2
asset hashes and uses a fresh VM for each case. It observes token trees and
errors without changing the parser. It also records pre-cleanup kept attributes
through the original `NodeUtil.setProperties` call.

Reproduce with a local copy of the pinned D2 assets:

```sh
node internal/tex/testdata/generate_mml_token.cjs /path/to/d2latex /tmp/mml-token-svg
go test ./internal/tex ./internal/svg -run TestMmlToken -count=1
```

The parser test checks complete kind/text/explicit-attribute/child trees, raw
error IDs/messages, kept attributes and the special movablelimits property.
The only kind representation conversion is Go's `mrow` with `Flags.Inferred`
to upstream `inferredMrow`. Numeric JSON representations are made consistent;
no tree nodes or explicit attributes are removed for comparison.

Cases cover every registered token kind, raw literal text, token/global default
attribute validation and the explicit allowlist, quoted/unquoted/duplicate/empty
values, boolean conversion, JS whitespace and case behavior, malformed inputs,
argument/error precedence, retained defaults, font contexts and script structure.
`mspace` consumes its text argument but has no children after upstream's
zero-arity inherited-attribute pass. A long-s boolean control prevents Go's wider
Unicode case folding from turning the literal `falſe` into boolean false.

Twenty-four cases also compare SHA-256 of the complete unmodified SVG output.
Non-SVG cases stop output after observing the compiled tree: in particular,
mglyph has no authored source and is a parser control, and explicit operator
spacing has its own independent renderer suite. No resource is fetched by this
generator. The existing Go `ms` renderer omits quotation glyphs; its token trees
are checked here, while that separate renderer discrepancy is not hidden by SVG
normalization or fixed in the parser.

`mjx-keep-attrs` protects explicitly authored token variants inside a TeX font
environment. The focused boundary test leaves the existing treatment of all
ordinary tokens and unrelated keep lists unchanged. The generator contains no
source overlays and does not depend on the separate brace-parser implementation.
