# Mathtools text laps

The original Mathtools `MtLap` handler parses `\clap`, `\textclap`,
`\textllap`, and `\textrlap` as internal text with embedded math. The old
Go shortcut emitted raw text for `clap`, while the other three commands
used the mathematical lap path. As a result, embedded math delimiters could
appear literally, text could become math, and script-level/font behavior
could differ.

The implementation follows pinned
[MathtoolsMethods.MtLap](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/mathtools/MathtoolsMethods.ts)
and [ParseUtil.internalMath](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/ParseUtil.ts).
It consumes an argument with the original JavaScript whitespace rules,
uses the existing internal-text parser with explicit script level zero,
and pushes a zero-width `mpadded` directly. Left and centered laps apply
`-1width` and `-.5width` offsets respectively. Caller text fonts are copied;
embedded math uses an empty lexical environment, shared declarations, and a
fresh macro counter. There is no additional TeXAtom.

The separate Base `llap`/`rlap` and six MathLap/cramped commands keep their
own source handlers. Dynamic operator and paired-delimiter overrides still
precede the built-in dispatch.

## Original references

`text_lap_mathjax_3_2_2.json` contains 1,386 distinct TeX/display-mode pairs
captured from D2's frozen original MathJax 3.2.2, source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. All complete SVGs are compared
without normalization: 1,194 valid expressions and 192 original error
renderings. No original runtime exceptions or unresolved output differences
occur in this inventory.

Against main152 (`cbfa500ef050c7240493dfe2624798219c959143`), 1,202 references
are newly exact and 184 were already exact. The inventory includes 192
independently captured controls and 32 macro-budget boundaries. It covers
literal text and escapes, dollar/parenthesis math, whitespace and malformed
arguments, fonts, colors, roots, scripts, pending parser items, shared
registrations and declaration priority. Separate controls retain the other
lap handlers. All 128 Bqty intersections are exact after composition with main152.
Two assertions already appeared in a published complete-original
fixture; `text_lap_overlap_report.json` binds these to their source. Per-fix
counts are not a claim of globally unique coverage.

`text_lap_promotions_test.go` also promotes two previously published raw
`clap` observations directly from the unchanged MathLap residual file. They
are not counted among the 1,386 newly captured references. The existing
96-case vector-font test now compares every original AST directly: its two
`\vb{\clap{x}}` cases no longer remove the original style attributes. The
historical fixture and complete SVG hashes remain unchanged.

Regenerate with:

```sh
node --jitless testdata/generate_text_lap.cjs /path/to/pinned-assets
```

The generator verifies all three asset hashes and calls only the frozen
original renderer in independent VM contexts. It never calls Go or selects
assertions based on candidate output.

## Published controls and validation

Replaying all 4,638 published residual inputs at main152 yields two source
fixes (the promoted `clap` observations) and no rendered or source regressions.
Six further serialized differences only reorder authored cancel attributes.
Forty-eight repeated observations from each binary produce both orders, and
all complete parsed XML trees, including numeric geometry, equal the original.
Raw strings and this proof are retained in the replay receipt; no golden is
normalized.

Final validation runs the full frozen-original suite, race
suite, vet, and WebAssembly build, run sequentially.

The first full gate detected that obsolete vector-font qualification. Removing
only the qualification restores comparison against the unmodified primary
tree; no production change or replacement golden was needed.
