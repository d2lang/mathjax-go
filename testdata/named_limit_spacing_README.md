# Spacing inside ordinary named limits

MathJax 3.2.2
[BaseMappings](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMappings.ts)
registers `liminf` and `limsup` as NamedOp with `lim&thinsp;inf` and
`lim&thinsp;sup`. The
[NamedOp handler](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts)
explicitly replaces `&thinsp;` with U+2006 before creating its movable-limits
operator. This is a handler-specific replacement, not general HTML entity
resolution.

The Go mapping previously used an ASCII space, widening the gap inside both
names and shifting their limits and following expressions. The fix changes
only those two values to U+2006. Both commands retain the existing NamedOp
construction and dynamic macro/paired-delimiter priority. The AMS `varlim*`
macro family is separate.

The fixture contains 812 complete original references: 724 valid outputs and
88 original error SVGs. All match exactly after the change; 578 differ from
the main155 baseline and 234 are unchanged controls. None was previously
recorded as a complete original SVG in main155's testdata. Coverage includes
both display modes, fonts, sizes, styles, limits and scripts, fractions, roots,
fences, arrays, adjacent spacing, pending and unbraced operands, neighboring
NamedOp commands, and paired/operator overrides in both declaration orders.
The 20 ordinary-limit controls retained from the AMS variable-limit audit are
included unchanged. An independent review contributed 98 additional exact
inputs, including nested scopes, definition sharing and script error precedence.

Four independent controls for AMS `varliminf` and `varlimsup` remain in
`named_limit_spacing_residuals.json`. Their full original and before/after
observations are preserved; this two-entry NamedOp change leaves them unchanged.
They are not counted as passing exact references.

A replay of 4,646 previously published residual inputs has no genuine changes.
Five apparent differences are the known authored cancel attribute order; 64
renders per binary per input preserve the complete original XML in both orders.
The 4,326 upstream TeX/display pairs, replayed under D2's fixed configuration,
show four exact fixes and no genuine regressions or changed nonexact outputs.
One circle-enclosure input changes only two-attribute serialization order and
retains identical existing nonexact geometry in both binaries; this was also
verified with 64 renders per binary. Raw SVG references are never normalized.

A visible witness is:

```tex
\Huge\boxed{\begin{gathered}\liminf_{n\to\infty}a_n\\\limsup_{n\to\infty}a_n\end{gathered}}
```

Regenerate with:

```sh
node --jitless testdata/generate_named_limit_spacing.cjs PINNED_ASSETS
go test ./... -run TestNamedLimitSpacingOriginalReferences
```

The generator verifies the three pinned D2 asset hashes and uses a fresh
original MathJax VM per conversion, with font cache none, em=16, ex=8 and the
stored display option. No Go output is used as an expected result.
