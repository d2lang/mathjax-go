# Base strut references

`base_struts_mathjax_3_2_2.json` stores 374 complete SVG strings from D2's
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

There are 370 valid renderings and four unchanged unknown-command error
controls. Against baseline `b1695a0`, 356 references fail and 18 already
match. All 374 match the original after the fix, with no qualified or
excluded cases in this corpus.

A compact D2 witness is `\frac{\strut a}{\mathstrut b}`. The old renderer
reports an undefined control sequence; the corrected renderer produces the
original fraction with the requested vertical spacing. Separate labels
`\boxed{\strut x}` and `\sqrt{\mathstrut y}` demonstrate each command.

Regenerate with
`node --jitless testdata/generate_base_struts.cjs PINNED_ASSETS`.
Run `go test ./... -run TestBaseStrutReferences`.
