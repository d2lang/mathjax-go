# Styled fraction scope

`styled_fraction_scope_mathjax_3_2_2.json.gz` contains 170 complete original
SVGs from D2's frozen MathJax 3.2.2 runtime: 154 valid formulas and 16 TeX
diagnostics. Both display modes are represented. Expectations are generated
only from the original, with a fresh JavaScript VM for each conversion.

AMS maps `dfrac` and `tfrac` to `AmsMethods.Genfrac`, with fixed style values
`0` and `1`. The method wraps only its completed fraction in `mstyle` before
pushing it. The Go implementation already has equivalent direct handlers, but
two old macro entries shadowed them. Their expansions inserted `displaystyle`
or `textstyle` into the caller, so the style persisted into following formulas
and scripts were attached inside that style. Removing those entries restores
the selected source handler and leaves dynamic paired-delimiter and operator
overrides in their existing order.

The corpus exercises following sums and fractions, subscript/superscript
attachment, outer scripts and script styles, ordinary fraction and binomial
controls, both supported dynamic override maps, and malformed or missing
arguments. The D2 witness is included in both display modes. On base commit
`b3ce847`, the first 168 references contain 48 changed outputs and 120 exact
controls; the two witness inputs also change. The corrected renderer matches
every complete original SVG. An additional 160 basic and 882 nested fraction
conversions also match their complete originals after the correction.

Regenerate with the exact three original assets and the bundled Node runtime:

```sh
python3 testdata/generate_styled_fraction_scope.py /path/to/assets /path/to/node
```

The generator verifies all asset hashes, disables JIT to avoid an intermittent
Node VM batch crash, and emits deterministic gzip bytes. It neither invokes Go
nor selects expectations by candidate output.
