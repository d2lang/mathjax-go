# Base strut references

`base_struts_mathjax_3_2_2.json` stores 982 complete SVG strings from D2's
frozen MathJax 3.2.2 bundle, upstream commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. The fixture records the three
original asset hashes. Every case uses a fresh VM, no font cache, `em=16`,
`ex=8`, and the recorded display mode. No output is normalized or generated
by the Go implementation.

Both commands already appear in the captured
[`BaseMappings`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMappings.ts),
but their runtime handlers were missing:

- [`BaseMethods.Strut`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts)
  creates `mpadded` around an explicit empty `mrow`, with `height="8.6pt"`,
  `depth="3pt"`, and numeric `width=0`.
- `mathstrut` expands exactly to `\vphantom{(}` through the ordinary macro
  path and uses the existing phantom implementation.

The base strut is distinct from the
[`AmsCdMethods.cell`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/amscd/AmsCdMethods.ts)
strut: CD uses `height="8.5pt"`, `depth="2pt"`, and no explicit empty `mrow`
or width attribute. The production change does not reuse or modify that
helper. The original SVG controls include CD diagrams both with and without
explicit base struts, so that distinction remains visible in the references.

The corpus covers both commands alone and in ordinary groups, ten math
sizes, four math styles, colors and fonts, roots and root indices, fractions,
scripts, accents, braces, fences, phantoms and smashes, matrices, arrays,
aligned equations, cases, internal text math, and CD diagrams. It also tests
unbraced arguments and pending function, dots, negation, prime, and positioning
items, including `raise`, `lower`, `moveleft`, and `moveright`.

The initial 374-case corpus contains 370 valid renderings and four unchanged
unknown-command controls; baseline `b1695a0` fails 356 of those references.
The final corpus adds 608 references for dispatch priority and independent
composition checks, including intentional syntax-error controls. It contains
958 valid renderings and 24 original errors. Against the latest main baseline
`ec073699cbca740b24abf49e9f223fb02d95bceb`, 878 of the 982 full SVG comparisons
fail. All 982 match the original after the fix; no outputs are normalized,
qualified or excluded from this corpus.

A compact D2 witness is `\frac{\strut a}{\mathstrut b}`. The old renderer
reports an undefined control sequence; the corrected renderer produces the
original fraction with the requested vertical spacing. Separate labels
`\boxed{\strut x}` and `\sqrt{\mathstrut y}` demonstrate each command.

Regenerate with
`node --jitless testdata/generate_base_struts.cjs PINNED_ASSETS`.
Run `go test ./... -run TestBaseStrutReferences`.


## Paired-delimiter command-map priority

Adding the ordinary `mathstrut` macro exposed an existing distinction between
MathJax's separate dynamic command maps. A declaration such as
`\DeclarePairedDelimiter{\mathstrut}{[}{]}\mathstrut{\Gamma}` must use the
paired-delimiter registration, even though the builtin macro still exists.

[`MathtoolsConfiguration`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/mathtools/MathtoolsConfiguration.ts)
registers the paired-delimiter map at priority -5, while
[`AmsConfiguration`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/ams/AmsConfiguration.ts)
registers declared operators at priority -1.
[`MathtoolsUtil.addPairedDelims`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/mathtools/MathtoolsUtil.ts)
adds the definition to its own map. Therefore a paired delimiter wins in either
order of `DeclarePairedDelimiter` and `DeclareMathOperator`; a later operator
does not displace it. Repeated declarations within the paired-delimiter map
replace the previous entry in that map.

The Go dispatcher now bypasses its lower-priority macro lookup when the name
has a paired-delimiter registration, then reaches the existing Mathtools
handler. It does not delete either definition or reorder the other package
handlers. The guards are validated with both struts and existing macro names
`dfrac`, `tfrac`, `stackrel`, `tripledash` and `Bra`, plus a new control name.
Both definition orders, alternating definitions, repeated paired definitions,
`DeclarePairedDelimiterX`, and `DeclarePairedDelimiterXPP` are covered. Contexts
include groups, scripts, color, functions, negation and vertical positioning.

Additional controls redefine `vphantom` before using `mathstrut`, showing that
its expansion remains an ordinary macro with dynamic command lookup. Operator
overrides of each strut also retain their original behavior. Full raw original
SVGs for all these controls are in the same fixture; none is a Go golden.
