# Alignedat arguments and array state

`alignedat` uses the original AMS `AlignAt` handler. It consumes an optional
vertical alignment followed by a mandatory pair count; neither argument is
table content. For example, `\begin{alignedat}{2}a&b\end{alignedat}` previously
rendered an extra digit 2. `[t]` and `[b]` align the first and last baselines,
`[c]` selects the axis, and other nonempty alignment strings pass through the
source `setArrayAlign` behavior.

Primary source is MathJax 3.2.2 commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`:

- [AMS environment registrations](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/ams/AmsMappings.ts)
- [AmsMethods.AlignAt](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/ams/AmsMethods.ts)
- [BaseMethods.EqnArray](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts)
- [EqnArrayItem row and tag lifetime](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseItems.ts)
- [ParseUtil.trimSpaces and setArrayAlign](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/ParseUtil.ts)

The source rejects nondigits but accepts empty and all-zero counts. Positive
counts supply repeating right/left, zero-gap definitions; they do not impose
a maximum authored column count. Empty or zero counts supply an empty
alignment with EqnArray's 1em fallback spacing. The source `GetNext` whitespace
boundary includes BOM and excludes NEL. This is applied to the shared
equation-count reader without changing unrelated argument consumers.
Missing-argument diagnostics name the active `\begin` control sequence.

The ordinary AMS route and the Mathtools special-row route both retain
alignedat's untaggable, row-local state. Explicit tags are rejected and labels
and no-tag commands do not leak into the surrounding equation. This does not
change the separate Mathtools arrow layout implementation.

`alignedat_mathjax_3_2_2.json` contains **984 complete original SVGs**, including
748 valid formulas and 236 original error SVGs. Against main135
`587d2d1e03afd952a7813c53f06bcfe39445c75b`, 876 fail and 108 already match.
The inventory covers both display modes, empty/zero/positive/unbraced counts,
extra authored columns, optional alignment and trimming, missing arguments,
JS whitespace, empty/final rows, nested arrays and caller scopes, tags and
labels, pending items, HFill, macro/paired overrides, Mathtools rows and
unchanged AMS environment controls.

The complete exploratory inventory is 1,140 inputs. The other **156** raw
references are retained in `alignedat_residuals.json` with original, baseline
and candidate outputs: 140 original-valid formulas and 16 original errors.
They are not passing expectations. There are no formerly exact regressions
against the baseline or against the intermediate argument-only candidate.

Fifty-eight residual outputs change. Eleven representative changed visual
families were rendered and inspected. The residual initial-operator spacing,
array-font reset, explicit row-spacing and paired-delimiter structure were
also reproduced byte for byte by unchanged baseline `alignat` controls whose
original SVGs equal the corresponding alignedat originals. Existing raw row
ownership (`\cr`, `\newline`, unbraced text ampersands) and malformed-end
diagnostics remain. The XalignAt BOM cases become valid tables and exactly
reproduce the existing ordinary-space XalignAt layout gap. Starred
ArrowBetweenLines still lacks its source column padding; these and its
continuation controls are retained separately. The initial-operator,
XalignAt/Flalign layout and Mathtools arrow corrections remain separate work.

Regenerate both retained inventories directly from D2's original runtime:

```sh
node --jitless testdata/generate_alignedat.cjs /path/to/pinned/assets
```

The generator verifies SHA-256 for all three frozen D2 assets, creates a fresh
VM for each recorded input, and regenerates the complete fixed inventory,
including residual originals. It never invokes Go, filters by candidate
results, or substitutes a Go-generated SVG for an original reference.
`TestAlignedatReferences` compares the complete SVG strings without
normalization.

Independent review rerendered 668 unique environment/argument controls:
598 are exact and the remaining controls reproduce existing differences.
A further 132 fresh Mathtools lifetime controls contain 106 exact results;
their changed residuals are the separately retained arrow-layout family.
Neither review found a formerly exact regression. The source review and
the full frozen-oracle, race, vet, and WASM checks pass.
