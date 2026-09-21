# Frozen quoted-string SVG oracle

The 60 cases in `ms_quotes_mathjax_3_2_2.json` use unmodified CommonMs/SVGms from
the D2 v0.8.1 MathJax 3.2.2 component. The generator verifies the three asset
hashes and constructs registered MathML nodes directly. It does not invoke
`mmlToken`, synthesize quote nodes in the input, or patch output code. Thirty
specifications run in both inline and display modes, each in a fresh VM.

```sh
node internal/svg/testdata/generate_ms_quotes.cjs /path/to/pinned/d2latex /tmp/ms-svg
go test ./internal/svg -run TestMsQuotes -count=1
```

The frozen input includes ordered explicit/inherited/default/global attributes,
node flags, text, properties and children after the primary inherited-attribute
and TeX-class passes. Go reconstructs that same semantic MathML input, including
the established `inferredMrow` → flagged `mrow` representation. Every case checks
the complete SVG SHA-256 without normalization. Repeat/clone rendering must be
identical, and the original text, attributes, child list, node identities and
parent identities must remain unchanged.

Cases distinguish unset curly defaults from explicitly authored straight quotes,
each unset/empty side, monospace, inherited and overriding quotes, multi-character
quotes, literal XML/TeX text, empty bodies, font variants, scaling, scripts,
fractions, inherited style, color, multiple text children and adjacent strings.
Non-string `mi`/`mtext` controls remain exact. This is a renderer test: D036 parser
support and D039/D040 ordinary font-command behavior are independent.

Primary source at MathJax-src commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`:

- `ts/output/common/Wrappers/ms.ts`: quote selection and wrapper-only insertion.
- `ts/output/svg/Wrappers/ms.ts`: CommonMs mixin for SVG output.
- `ts/output/svg/Wrappers/TextNode.ts`: positionable groups for multiple text
  wrappers, required to place opening/body/closing glyphs separately.

The Go grouping change is deliberately limited to `ms` parents. No parser or
other text-output behavior is changed by this discrepancy fix.
