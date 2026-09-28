# Cramped stacks and primed array style

The original Mathtools `\crampedsubstack` macro expands one argument to
`\begin{crampedsubarray}{c}#1\end{crampedsubarray}`. Go lacked that registration.
It also treated the `S'` style of `crampedsubarray` as ordinary `S`, so powers
inside directly authored cramped arrays retained uncramped placement.

The patch restores the exact macro and sets `data-cramped=true` before MathML
inheritance. The existing script level, display style, and `useHeight=false`
array behavior remain in use. The relevant original sources, at commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`, are
[MathtoolsMappings](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/mathtools/MathtoolsMappings.ts)
and [BaseMethods.Array](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts).

## Complete original references

The source-derived corpus and two independent reviews contain 1,567 unique
public inputs. On baseline `d74de82316658af59f3d0e8a409b4bbeb005fb5d`, the
candidate matches 1,226 complete original SVGs: 1,108 valid renderings and 118
original error renderings. These include 1,006 fixes and 220 exact controls,
with no loss of a previously exact reference. The passing fixture contains
only unmodified original SVGs; tests compare the complete strings.

Coverage includes direct arrays and macro expansions, ordinary substack and
subarray controls, powers and nested scripts, surrounding math styles, fonts,
fractions, radicals, limits, boxes, array columns, pending parser items,
argument and row boundaries, and dynamic command overrides.

`cramped_substack_residuals.json` preserves all 341 other original, baseline,
and candidate outputs, including the explicit controls used to establish
inheritance. Of these, 105 outputs are unchanged and 236 change. Thirty-six
are original errors. They are not normalized or asserted as passing SVGs.
`cramped_substack_review.json` records both independent source reviews,
literal-expansion checks, glyph/style checks, and geometry observations.

The remaining valid families concern existing array font leakage, framed or
dashed column specifications, row rounding, nested row ownership, and the
optional-style behavior of `\cramped[\scriptscriptstyle]`. Literal expansions
reproduce the macro cases in both original and Go. Ordinary subarray/substack
controls independently reproduce the surrounding font, frame, and scale
defects. For framed arrays, the candidate's complete power subtrees match the
original even though the inherited exterior frame and spacing remain wrong.
For the optional-style family, the complete inner cramped-style subtree
matches its unchanged ordinary-substack control. This evidence establishes
which behavior belongs to the new macro without accepting inherited failures
as new golden output. Original-error families retain their diagnostic and
closing-scope differences as raw observations.

The second independent review's 354 distinct inputs include 244 exact matches,
218 fixes, and no byte or previously visual-exact regressions. All 140 stored
macro/literal pairs agree in both original and candidate. A replay of 3,040
published historical residual inputs adds one exact match, with no exact
regressions or changed nonexact results.

## Reproduction

Run `python3 testdata/generate_cramped_substack.py /path/to/pinned/assets NODE`.
The generator verifies the three recorded SHA-256 asset hashes and invokes
only the unmodified original oracle, with a fresh runtime for every input and
at most 24 runtimes per subprocess. It regenerates original strings while
preserving historical baseline/candidate observations. It never invokes Go.
