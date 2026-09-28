# Optional styles for cramped math

The frozen original is MathJax 3.2.2, source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`.
[MathtoolsMethods.Cramped](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/mathtools/MathtoolsMethods.ts#L229-L236)
reads an optional literal style, trims JavaScript whitespace, parses its required
argument, and creates an `mstyle` with `data-cramped=true`.
[MathtoolsUtil.setDisplayLevel](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/mathtools/MathtoolsUtil.ts#L47-L58)
recognizes only `\displaystyle`, `\textstyle`, `\scriptstyle`, and
`\scriptscriptstyle`, setting the corresponding display flag and absolute
script level. It ignores all other option strings without expanding macros.

Go previously reused the numeric/letter style lookup for generalized fractions.
It ignored the four literal commands and accepted unrelated options such as
`2` or `SS`. This change supplies the separate Mathtools lookup; generalized
fraction styles, the required-argument parser, and other commands are unchanged.
The existing JavaScript-space predicate includes BOM and excludes NEL.

A visible example is:

```tex
\Huge A=\cramped[\scriptscriptstyle]{x^2}+\cramped[\textstyle]{\sum_i^n i}
```

The original and candidate SVGs are identical, with view box
`0 -1966.1 14144.8 2696`. The baseline uses full-sized content and display limits,
with view box `0 -3890.7 15612.8 6859.2`.

## Original references

Against merged main145, `aa6200c004e8498321d821ac7edceb4040b441cb`, the inventory
contains **1,328 unique TeX/display inputs**: **1,236 exact complete original SVGs**
(1,184 valid expressions and 52 original error renderings), including **601 fixes
and 635 unchanged exact controls**. No formerly exact input regresses.

Coverage includes all four styles and default/empty options; unknown numeric,
letter, prototype-property and literal macro names; nested scripts, fractions,
roots, accents, arrays and CD; fonts, colors, sizes and positioning; pending
function/Not/Dots items; grouped, unbraced and missing operands; declaration
priority and literal options that contain definitions; JavaScript whitespace;
and unchanged generalized-fraction and overlap-command controls. Independent
source review and 136 fresh inputs are included in these counts. An additional
72 controls cover all Object-prototype property names, bare and wrapped in BOM
or NEL. The source [Options.lookup](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/util/Options.ts)
checks own properties, so these names are ignored rather than treated as styles.
The final composition adds 96 inputs involving the newly restored cramped stacks
and ten distinct explicit-style controls. The preceding 1,222 inputs have
identical baseline and candidate outputs on main143 and main145.

The affine SVG check composes every transform and compares outlined glyphs,
paint and root metrics. The complete 1,328-input inventory has **591 visible fixes**
and **zero visual regressions among 645 previously visually exact inputs**.
The ten additional byte-exact fixes affect structure only and are not counted
as visible fixes.

A replay of 3,479 published residual inputs finds 24 newly exact optional-style
cases in `cramped_substack_residuals.json`. A separate test now asserts their
complete original SVGs, bringing the assertion count for this change to 1,260.
That historical file is unchanged. The only other observed replay difference
was the already recorded nondeterministic ordering of two cancel attributes;
five repeat runs of both binaries bind all 16 affected cancel inputs to the
same two exact serializations. No rendering change is waived or normalized.

## Raw residuals and control proof

`cramped_style_residuals.json` retains **92 unmodified original/baseline/candidate
observations** separately from exact assertions. **67 remain byte-unchanged;
25 change without becoming exact**. They are:

- 32 expressions involving an explicitly authored grave `mo` with a superscript.
  Correcting the selected style exposes the existing reverse-prime script-layout
  discrepancy at the selected size.
- 32 corresponding explicit-style controls, all byte-unchanged. They use an
  ordinary style declaration around default `\cramped`, bypassing the changed
  optional-style lookup.
- Twelve MathFont/cramped-stack expressions and ten distinct explicit-style
  controls. The existing font scope across arrays differs from the original.
  All twelve expressions have exactly the same flattened glyphs, paint and
  root metrics as their respective controls, in both implementations. Every
  control is byte-unchanged by this patch.
- Two unchanged malformed `\cramped[{\scriptstyle]` inputs. The original error
  names `\cramped`; Go's existing shared GetBrackets error omits the command name.
- Four unchanged `\crampedllap` / `\mathrlap` optional-argument controls. Their
  missing optional-argument parsing belongs to the separate MathLap handler.

`cramped_style_geometry_proof.json` binds every grave and MathFont/array case
to its explicit-style control. All 32 grave pairs have identical original root metrics and identical Go
root metrics, and preserve the same glyphs and paint. For 29 pairs the entire
flattened geometry is identical. Three inline fraction controls add paired
three-decimal scale transforms; their full transform differences are identical
on the original and Go sides. The proof preserves these differences instead
of normalizing them away. No raw residual is accepted as an exact golden.

## Regeneration

```sh
node --jitless testdata/generate_cramped_style.cjs PINNED_ASSETS
go test ./... -run TestCrampedStyleOriginalReferences
```

The generator verifies all three frozen asset hashes and uses a fresh original
VM for each conversion in bounded batches of 24. It regenerates complete SVGs
with font cache disabled, em=16, ex=8 and the recorded display mode. Unicode
separators are escaped only on the JSONL wire. Historical Go observations and
the derived geometry proof are retained unchanged.
