# Frozen Mathtools bigtimes references

The unmodified D2 MathJax 3.2.2 macro dispatcher registers `bigtimes` as Base
`Macro` with the exact expansion
`\mathop{\Large\kern-.1em\boldsymbol{\times}\kern-.1em}` and no arguments.
The generator captures the actual registry entry, method identity, arguments,
priority, and function source before creating fresh runtimes for every render.

There are 358 complete valid-original SVGs in both display modes. Authored
documented Mathtools `DeclarePairedDelimiterX`/`DeclarePairedDelimitersX`
declarations provide the formatter dependency; contexts include spacing,
scripts, limits, styles, fractions, fonts, colors, matrices, fences, and runtime
overrides. Direct expansions and unrelated operators are independent controls.

Ten complete original diagnostic objects are retained separately with zero
valid-original/SVG parity credit. D2's frozen package set does not enable
`boldsymbol`; the original bare `bigtimes` expansion therefore reports undefined
`boldsymbol`. Go intentionally has a pre-existing `boldsymbol` extension, so the
new registration uses that formatter normally. This change preserves that Go
extension and does not claim that bare diagnostics now match the original.

Regenerate using the hash-verified frozen assets and pinned Node:

```sh
python3 testdata/generate_mathtools_bigtimes.py /path/to/oracle-assets /path/to/node
```

Expected SVGs are complete unchanged original API outputs, not Go-generated
references. The finite references do not establish universal MathJax parity.
