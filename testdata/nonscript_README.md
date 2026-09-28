# Nonscript spacing

References use the frozen D2 MathJax 3.2.2 assets, whose source commit is
`ad8f5c21cb810236551da8c6512ba733e67357ee`. All SVG goldens are complete,
unmodified original outputs. Runtime failures and other mismatches remain in
`nonscript_residuals.json` with original, baseline, and candidate observations;
they are not counted as exact references.

The implementation follows
[BaseMethods.Nonscript](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts),
[BaseItems.NonscriptItem](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseItems.ts), and
[BaseConfiguration.filterNonscript](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseConfiguration.ts).
The item examines its next actual stack item. It marks an immediately
following mspace, including a space inside a not-parent mstyle. Commands
which emit no item leave it pending; groups, styles, functions, positions,
other stack items and closing events consume it. User command registrations
still precede the builtin.

The filter runs after inheritance and before moveLimits/cleanStretchy. It
removes marked spaces only when the inherited scriptlevel is positive. A
fixed-size space's explicit level-zero mstyle receives a temporary outer row
to remember the surrounding level; a retained space loses only that temporary
row. Numeric comparison uses the existing ECMAScript scalar conversion.

`\Huge a\nonscript\qquad b+x^{a\nonscript\qquad b}` is the visible witness:
the ordinary gap remains, while the gap in the exponent disappears.

The current corpus has 2,282 distinct inputs: 2,032 complete exact originals
(1,660 valid expressions and 372 error renderings), 1,876 fixes and 156 exact
controls against main142 plus the separate Quantity fallback prerequisite
`41eef3f1e2c5afc7db16f8f9c32d75e2780871cf`. There are no previously exact
regressions. Both modes are covered, including ordinary and fixed spaces,
Rule/Space, nesting, roots/fractions, scripts/primes, pending items, positions,
fonts, overrides, arrays/CD, Mathtools row commands, and sequences of 1,001
non-expanding Nonscript commands.

The 250 raw references are retained by category:

- 118 original runtime failures where removal empties a required fixed-arity
  slot. Go returns a bounded nonnil internal-conversion error with no SVG,
  rather than propagating a process panic or drawing invented content. The
  raw TypeError receipts are preserved and the API behavior has a dedicated
  test. A fresh 404-case arity sweep retained SVG output for every one of its
  272 original SVG outcomes; none was rejected by the guard.
- 24 other original runtime failures: prescript slot removal and VDotsWithin
  receiving a pending Nonscript item. Existing nonpanic Go outcomes remain
  raw, without claiming parity.
- 40 inherited underbrace clipping-rounding observations, including literal
  controls. Outer geometry remains unchanged.
- 28 inherited supported Quantity argument/style-boundary observations and
  controls. The separate Quantity fallback prerequisite fixes the unrelated
  caller-command consumption cases.
- 20 authored nonfinite scriptlevel observations and literal controls; the
  original itself emits NaN geometry.
- 20 other inherited observations: false scriptlevel scaling and unavailable
  boldsymbol, with corresponding controls.

Independent source reviews cover item eligibility, temporary rows, filter
ordering, dynamic priority, and the bounded arity failure. Independent public
sets contribute 330 and 178 inputs; the combined branch removes the eight
Quantity fallback differences from the first set and avoids all six process
panics found in its initial candidate. All original observations are retained. The main143 rebase leaves every
corpus output unchanged. A 3,040-input published-residual replay against the
Quantity prerequisite changes only two already documented authored cancel
attribute orders; the complete parsed SVGs and all geometry are identical.
Both serialized observations and the attribute-order proof are retained.

Regenerate both original-reference files with:

```sh
node --jitless testdata/generate_nonscript.cjs /path/to/pinned-assets
```

The generator verifies all three asset SHA-256 hashes and creates a fresh
original VM per expression. It never runs Go or substitutes candidate output.
Both reference files regenerate byte-identically from the frozen bundle.
Full repository gates are recorded after the prerequisite is merged and the
final base is known.
