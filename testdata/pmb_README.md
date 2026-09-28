# Poor-man’s bold macro

The pinned MathJax 3.2.2 [BaseMappings.ts](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMappings.ts)
registers `pmb` as the one-argument macro `\rlap{#1}\kern1px{#1}`. The first
copy is overlaid with zero width; the second is shifted by one pixel to make
the combined strokes heavier. Go had the required commands but omitted this
macro from its active registration map, producing an undefined-command error.
The production change adds exactly this source expansion to `simpleMacros`.

A clear render witness is:

```tex
\Huge\boxed{\pmb{\alpha+\beta=\gamma}\quad\pmb{\frac{a+b}{c+d}}}
```

The complete candidate SVG equals the frozen original for this witness. The
baseline reports an undefined `pmb` command. No font-weight approximation or
new geometry implementation is introduced.

## Original references

`pmb_mathjax_3_2_2.json` stores 1,038 complete, unmodified original SVG strings
from D2’s frozen MathJax bundle at source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. It records all three verified asset
hashes and baseline main125 commit `6a7adcd1c6cce9df99b3363d40f22c2f68857fd5`.
Every original render uses a fresh VM, font cache none, `em=16`, `ex=8`, and
the recorded display mode. There are 990 valid original renderings and 48
original error SVGs. Against that baseline, 62 were already exact and 976
are newly exact, with zero formerly exact regressions.

The full 1,048-case inventory includes grouped and unbraced operands, empty
and invalid arguments, scripts and primes, style/font/color changes, nested
macros, fractions, radicals, accents, positions, arrays, tables inside tables,
vector macros, dynamic operator and paired-delimiter overrides, and overridden
commands inside the expansion. Display and inline modes are both covered.
Explicit expansion and terminal control-space controls are retained. Two
independent reviews (186 and 68 fresh cases) were entirely exact; their 244
nonduplicate observations are included in these fixtures.

## Raw residuals

`pmb_residuals.json` preserves all ten other observations with raw original,
baseline and candidate SVGs. All ten originals render successfully, none was
exact at baseline, and all ten candidate outputs change from the former
undefined-command error. They are not asserted as correct.

Eight reproduce an existing difference in the already supported literal
expansion: four MathFont compositions, two pending-function/operator-token
compositions, and two nested-table source-splitter compositions. The file
also stores each expanded TeX input and both original and baseline expansion
SVGs, proving that the macro registration adds no separate behavior there.

The other two end with a terminal backslash. Original `GetArgument` returns
`"\\" + GetCS()`, and original `GetCS` returns a space at end of input. Go’s
shared argument helper retains the raw backslash instead. An explicit trailing
control space is exact and covered in the main fixture. This shared argument
normalization issue is retained separately rather than silently excluding the
valid original form or changing the shared parser in a macro registration fix.

Regenerate both files with
`node --jitless testdata/generate_pmb.cjs PINNED_ASSETS`. The generator verifies
the asset hashes and uses batches of 24 fresh VMs. It updates original SVGs
only, preserving the input inventory and baseline/candidate residual receipts.
Run `go test ./... -run TestPmbReferences` for the complete reference check.
