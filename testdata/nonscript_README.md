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

The original pre-composition corpus has 2,282 distinct inputs: 2,032 complete exact originals
(1,660 valid expressions and 372 error renderings), 1,876 fixes and 156 exact
controls against main144 (`a9fcab295ac962d46f533c5d3e4993a50f1e7bf4`),
which includes the separately merged Quantity fallback prerequisite. There are no previously exact
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
panics found in its initial candidate. All original observations are retained. The main143 rebase left every
corpus output unchanged; the main144 branch drops the now-merged Quantity
dependency without changing the composed production source. A 3,040-input published-residual replay against the
Quantity prerequisite changes only two already documented authored cancel
attribute orders; the complete parsed SVGs and all geometry are identical.
Both serialized observations and the attribute-order proof are retained.

## Composition with array entry repair

The Nonscript source is rebased without modification onto merged main146,
`4aeeb29d500693c2b476818d502e205892aaf446`. All 2,282 preceding candidate
outcomes remain byte-identical to the independently reviewed main144 candidate.
A separate 486-input original inventory exercises array entry repair before
Nonscript's inherited-space filter: plain and fixed-size spaces, leading
operators, relation atoms, fractions, fonts, grouped spaces, pending functions,
Not/Dots/Prime/Position/Braket/AutoOpen/infix items, optional vertical alignment,
left/right gathered layouts, and actual-array-top command guards.

Of these, 476 match complete originals (420 valid and 56 error renderings).
There are 160 improvements from the previous Nonscript candidate, with no
formerly exact regressions. The ten residuals preserve two original VDotsWithin
runtime failures and eight existing shoveleft/shoveright diagnostic-wording
differences; every residual is byte-unchanged from the reviewed Nonscript source.
The original runtime outputs remain raw, and Go stays nonpanicking.

Together the two inventories contain **2,864 unique inputs, 2,604 complete
original SVG assertions, and 260 raw observations**. Against main147 there are
**2,448 fixes and 156 exact controls**. The 118 dedicated required-child failure
checks remain unchanged. No original reference was replaced by Go output.

The candidate is finally composed with merged main147,
`8c7516e9747a72ef7ecae4ac8d6d6cfc4bfd3332`, which adds the original optional
cramped styles. All 2,768 preceding baseline and candidate outcomes remain
byte-identical. Another 96 fresh original inputs exercise all four optional
styles with ordinary/fixed spaces, font declarations, grouped spaces and
aligned entries, both alone and in superscripts. All 96 are exact and all 96
fix baseline failures. They are included in the array-composition fixture;
its earlier 486-case metadata remains historical, with the additional 96
explicitly recorded as `crampedStyleComposition`.

The main147 published-input replay covers 4,011 unique inputs. Its only three
byte differences are known cancel attribute-order changes, with no source
change. Forty-eight repeat runs per input and binary bind outputs to the same
two exact serializations, with identical full parsed XML trees.
`nonscript_array_published_replay.json` preserves those outputs and proof.
No rendering or source regression is hidden by a qualifier.

Regenerate all four original-reference files with:

```sh
node --jitless testdata/generate_nonscript.cjs /path/to/pinned-assets
```

The generator verifies all three asset SHA-256 hashes and creates a fresh
original VM per expression. It never runs Go or substitutes candidate output.
The reference files regenerate byte-identically from the frozen bundle.
On merged main144, the full frozen-oracle suite, race tests, vet, and
WebAssembly build all pass. The rebased production source is identical to
the independently reviewed composed main143 candidate; no gate repair or
reference substitution was needed. Both reference files were regenerated
again after recording the merged baseline and remained byte-identical.

## Final validation on main147

The full frozen-original suite, race tests, vet, and WebAssembly build pass
after composing with merged array alignment and optional cramped styles. All
four original files regenerate byte-identically. Independent review confirms
unchanged prior outcomes, exact fresh style controls, and zero source-related
changes in the historical replay. All 52 D2 witnesses match their original
SVGs. The new `parity/nonscript-spacing` witness preserves the ordinary gap and
removes the gap inside the exponent; its before SVG records the formerly
undefined command, and fixed/original SVGs are byte-identical.
