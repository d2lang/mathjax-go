# Mathtools pairedDelimiters TeX setter

The MathJax 3.2 package documentation lists `pairedDelimiters` as a predefined
delimiter configuration object. D2 initializes its fixed package set before
conversion; neither the frozen D2 renderer nor the Go conversion options expose
custom JavaScript package initialization.

The active original `MathtoolsMethods.SetOptions` additionally accepts this key
through `\mathtoolsset`. Its exclusion list contains the misspelled
`pariedDelimiters`, together with `tagforms` and `allow-mathtoolsset`. Consequently,
`\mathtoolsset{pairedDelimiters={}}a=b` is valid original input. The value is stored
without rebuilding the dynamic delimiter registration map. Existing and later
TeX delimiter declarations continue working independently.

`generate_mathtools_paired_option.cjs` reads only hash-verified, unmodified D2
MathJax 3.2.2 assets at original commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. Every render and direct method observation
starts in a fresh runtime. Regenerate with:

```sh
node --jitless testdata/generate_mathtools_paired_option.cjs /path/to/pinned-assets
go test ./... -run TestMathtoolsPairedOption -count=1
```

The fixture retains 284 complete SVGs: 274 valid originals and ten rendered
diagnostics. Against the canonical pre-fix revision `fe29d3a`, 266 valid cases
fail through `UnknownKeyVal`; eight valid controls and eight diagnostics are
unchanged. Two diagnostic cases now reach the later unknown option, matching the
original diagnostic; these receive no valid-input parity credit. The 19 direct
original method observations bind the exact exclusion
typo and verify boolean/string values, brace handling, duplicate-key replacement,
and the three forbidden/unknown-key boundaries. No original runtime failures are
counted as SVG evidence. These finite checks do not establish universal parity.

Primary documentation:
[MathJax 3.2 Mathtools options](https://docs.mathjax.org/en/v3.2/input/tex/extensions/mathtools.html#tex-mathtools-pairedDelimiters),
pinned documentation revision `a39fc2d1f7ef82477636c1aeb9dd8d8fc3eb3fe4`,
`input/tex/extensions/mathtools.rst` lines 153-161. The active method source is
captured verbatim in `internal/tex/testdata/mathtools_paired_option_observations.json`.
