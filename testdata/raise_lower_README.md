# Raise/lower references

`raise_lower_mathjax_3_2_2.json` contains 258 complete SVGs from D2's unmodified
MathJax 3.2.2 bundle. The fixture records the three oracle asset hashes checked
by `internal/oracle/runner.go`. Each case uses a fresh runtime, font cache
`none`, em 16, ex 8, and the listed display mode. The fixture was produced with
`testdata/differential/oracle.mjs`, using Node with `--jitless`.

The cases cover both commands, positive/negative/explicit-plus/zero/braced
lengths, groups, complete unbraced commands, macros, font and style scopes,
functions, scripts, primes, nested positions, row-closing errors, linebreak
argument precedence, a chemical reaction with a lower label, and unchanged
controls. Existing dimension and mu-dimension tests now compare their six
previously qualified raise/lower cases directly to the original tree and SVG.

The source behavior comes from MathJax commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`:

- `ts/input/tex/base/BaseMethods.ts`: `RaiseLower` reads a dimension, reverses
  direction for a negative length, and pushes a PositionItem. `CrLaTeX` reads
  its optional dimension before pushing a closing cell item.
- `ts/input/tex/base/BaseItems.ts`: `PositionItem.checkItem` wraps the next final
  MML item in mpadded; a closing item instead reports a missing box.

Regenerate by sending each case's `tex` and `options: {"Display": case.display}`
to the oracle and replacing only its `svg` with the response. Use assets with
the recorded hashes.
