# Radical size references

`radical_size_mathjax_3_2_2.json` contains the complete SVG returned by the
unmodified D2 MathJax 3.2.2 bundle for each listed TeX input. The bundle's three
SHA-256 values are recorded in the fixture and match `internal/oracle/runner.go`.
Each input uses a fresh runtime, font cache `none`, em 16, ex 8, and the listed
display mode. The references were generated with `testdata/differential/oracle.mjs`.

The 64 cases cover square and indexed roots under all ten size declarations,
normal-size controls, tall radicands, scripts, nested roots, independently sized
nested roots, a fraction as the index, size scope, and a colored root.

The constructor behavior is from these sources at MathJax commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`:

- `ts/output/common/Wrappers/msqrt.ts`: the constructor calls `createMo`.
- `ts/output/common/Wrapper.ts`: `createMo` inherits attributes before wrapping.
- `ts/core/MmlTree/MmlNode.ts`: `inheritAttributesFrom` copies effective display
  style, script level, math size when set, and prime style.

To regenerate, send one JSON line per case to the oracle, using the case's `tex`
and `options: {"Display": case.display}`, and replace only that case's `svg` with
the returned `svg`. `MATHJAX_GO_ORACLE_DIR` must refer to assets with the recorded
hashes.
