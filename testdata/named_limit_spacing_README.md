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

The fixture contains all 714 original inputs: 634 valid outputs and 80 original
error SVGs. All are complete-original exact after the change; 522 differ from
the PR 153 baseline and 192 are unchanged controls. There is no discarded or
qualified residual subset. Coverage includes both display modes, fonts, sizes,
styles, limits and scripts, fractions/roots/fences/arrays, adjacent spacing,
pending and unbraced operands, neighboring NamedOp commands, and paired/operator
overrides in both declaration orders. The 20 ordinary-limit controls retained
from the AMS variable-limit audit are included unchanged.

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
