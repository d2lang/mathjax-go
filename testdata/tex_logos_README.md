# TeX and LaTeX logo macros

The pinned MathJax 3.2.2 [BaseMappings.ts](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMappings.ts)
registers the two zero-argument macros with these exact expansions:

```tex
\TeX   => T\kern-.14em\lower.5ex{E}\kern-.115em X
\LaTeX => L\kern-.325em\raise.21em{\scriptstyle{A}}\kern-.17em\TeX
```

Go’s generated source table included these registrations, but its active macro
map omitted them, so both logo commands produced an undefined-command error.
The production change adds exactly these two strings to `simpleMacros`. The
space before the final `X` is retained, and `LaTeX` still invokes `TeX` through
the ordinary macro dispatcher, preserving later user overrides. The existing
kern, raise/lower and scriptstyle implementations supply the original geometry.

A clear exact render witness is:

```tex
\Huge\boxed{\mathrm{\TeX\quad+\quad\LaTeX}}
```

The candidate SVG equals the complete original output; the baseline shows an
undefined-command error. A prerequisite audit on main126 also checked 40 logo contexts
against both their explicit original expansion and the existing Go expansion:
all 40 matched, establishing that no positioning repair was required.

## Original references

`tex_logos_mathjax_3_2_2.json` stores 818 complete, unmodified original SVGs
from D2’s frozen MathJax 3.2.2 bundle, source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. It records all three verified asset
hashes and baseline main130 commit `ed0afff52f6f50eb2e49daf7eada8cd9369a781a`.
Every original uses a fresh VM, font cache none, `em=16`, `ex=8`, and the
recorded display mode. There are 682 valid renderings and 136 original error
SVGs. Against this baseline, 288 cases were already exact and 530 are newly
exact, with zero formerly exact regressions. No residual was excluded from
the complete 818-case inventory.

The cases cover both logo commands, literal expansions, adjacency and control
sequence boundaries, scripts and primes, true unbraced operands, fonts/colors/
sizes/styles, fractions and radicals, accents, positioned contents, arrays,
HFill and CD contexts, internal math in text, and ordinary versus typed function
recipients. Dynamic operator and paired-delimiter overrides are exercised in
both registration orders for the logo names and for helpers used in their
expansions. Redefining `TeX` and then rendering `LaTeX` is explicitly covered.

The final inventory includes a separate 48-case unbraced declaration and name-
spacing sweep, an independently generated 272-case source review (all exact),
and 40 fresh compositions with the merged `pmb` macro. Duplicate expressions
are included only once. All earlier baseline and candidate outputs were
unchanged after rebasing from main127 onto main128. The final main130 baseline
recheck retains the same 530 failures and 288 exact controls.

Regenerate with
`node --jitless testdata/generate_tex_logos.cjs PINNED_ASSETS`. The generator
verifies the three original hashes and runs batches of 24 fresh VMs. Run
`go test ./... -run TestTeXLogoReferences` for complete raw-SVG comparisons.
