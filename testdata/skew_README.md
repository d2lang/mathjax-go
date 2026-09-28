# Skewed accents

The pinned MathJax 3.2.2 [BaseMappings.ts](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMappings.ts)
registers `skew` as this three-argument macro:

```tex
{{#2{#3\mkern#1mu}\mkern-#1mu}{}}
```

Go’s generated source table contained that definition, but its active macro
map omitted it. The production change adds the exact expansion to
`simpleMacros`. The normal Macro argument reader obtains the shift, decorator
and body as three TeX arguments. An unbraced argument is one character or
control sequence; the macro does not add its own dimension parser or require
that the second argument name a particular accent command.

The added kern widens the accented body, shifting the accent. The opposite
kern restores the overall width. Both pairs of outer braces and the final
empty group are retained because they also affect ownership and script
attachment. The source substitutes the amount verbatim before `mu`, including
its original behavior for signed, fractional, empty and invalid values.

A clear render witness with room for the overhanging accents is:

```tex
\Huge\boxed{\hat{x}\qquad\skew{9}\hat{x}\qquad\skew{18}\hat{x}\qquad}
```

In the frozen original, `\skew{18}\hat{x}` moves the hat’s translation from
313.8 to 786 SVG units relative to unskewed `\hat{x}`, a 472.2-unit shift,
while both expression viewBoxes remain 572 units wide. A separate 64-case
prerequisite audit across amounts, accents, bodies and modes matched both the
original literal expansion and Go’s existing expansion. No geometry change
is needed to implement the missing macro.

## Original references

`skew_mathjax_3_2_2.json` contains 646 complete, unmodified original SVGs from
D2’s frozen MathJax 3.2.2 bundle, source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. It records all three verified asset
hashes and baseline main130 commit `ed0afff52f6f50eb2e49daf7eada8cd9369a781a`.
Every original uses a fresh VM, font cache none, `em=16`, `ex=8`, and the
recorded display mode. There are 562 valid original renderings and 84 original
error SVGs. Against the baseline, 54 were already exact and 592 are newly
exact, with zero formerly exact regressions. The original main128 baseline and
candidate receipts were retained; all 698 outputs from each side are unchanged
after rebasing the candidate and rechecking the merged main130 baseline.

The complete 698-input inventory covers 15 accent/decorator names, ordinary
and composite bases, shift forms, grouped and unbraced arguments, empty and
invalid operands, scripts/primes, fonts/styles/colors, fractions and radicals,
positions, matrices, HFill/CD contexts, nested skew, dynamic operator/paired
redefinitions of skew and its helpers, and literal expansions. An independent
96-case review was entirely exact; its nonduplicate inputs are included here.

## Retained expansion differences

`skew_residuals.json` preserves the other 52 raw observations, including each
original, baseline and candidate SVG and its explicit source expansion. All
52 originals are valid, none was exact at baseline, and all 52 candidate
outputs differ from the former undefined-command error. These are not asserted
as correct or silently replaced with Go SVGs.

For 50 cases, the candidate is byte-identical to baseline Go rendering of the
literal expansion, which already differs from the original. These cover
ordinary `@` token wrapping, grave-accent glyph remapping, and MathFont composite
outputs. Some cases overlap two of those boundaries. The original macro SVG
matches its original literal expansion for every retained case.

The other two cases end with a backslash as the third argument. Original
`GetArgument` normalizes that terminal control sequence to a control space;
Go’s shared argument helper retains the raw backslash. The normalized literal
expansion matches the original in Go. The explicit control-space form is an
exact control in the main fixture. This shared parser issue stays separate
from the missing macro registration.

Regenerate both fixture files with
`node --jitless testdata/generate_skew.cjs PINNED_ASSETS`. The generator checks
the asset hashes, uses batches of 24 fresh VMs and refreshes both original
macro and original literal-expansion SVGs. Run
`go test ./... -run TestSkewReferences` for complete SVG assertions.
