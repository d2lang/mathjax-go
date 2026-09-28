# Accent italic-correction references

`accent_italic_correction_mathjax_3_2_2.json` contains the complete SVG returned by
D2's unmodified MathJax 3.2.2 bundle. The fixture records the SHA-256 values for
all three oracle assets, which match `internal/oracle/runner.go`. Each input
uses a fresh runtime, font cache `none`, em 16, ex 8, and the listed display mode.
The references were generated with `testdata/differential/oracle.mjs`.

The 96 cases cover 12 accents or annotations on calligraphic T, calligraphic F,
and ordinary x in display and inline modes. Additional cases cover hat,
widehat, underline, and overline on calligraphic T in large text, scripts, and
nested accents. Overline, ordinary characters, and non-accent annotations
provide unchanged controls.

The source of the correction policy is `CommonScriptbaseMixin`'s constructor
in `ts/output/common/Wrappers/scriptbase.ts`, at MathJax commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. It sets `baseRemoveIc` to false for
line accents; otherwise, it removes italic correction when `useIC` is false
or `isMathAccent` is true. The SVG munder, mover, and munderover wrappers inherit
`useIC` from msub, msup, and msubsup respectively.

To regenerate, send one JSON line per case to the oracle using the case's `tex`
and `options: {"Display": case.display}`, and replace that case's `svg` with
the returned `svg`. Use oracle assets with the recorded hashes.
