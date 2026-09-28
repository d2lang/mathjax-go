# AMS operator declaration names

The primary reference is D2's frozen MathJax 3.2.2 at
`ad8f5c21cb810236551da8c6512ba733e67357ee`.
[AmsMethods.HandleDeclareOp](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/ams/AmsMethods.ts#L206)
reads an ordinary argument, applies ParseUtil.trimSpaces, removes one optional
leading backslash, and registers the whole resulting name. It does not invoke
the stricter NewcommandUtil.GetCsNameArgument validator.

Previously the Go AMS handler reused that strict name reader. It rejected
valid bare names such as `\DeclareMathOperator{foo}{Q}\foo x`, as well as names
that the original permits to be registered without immediate invocation. The
caller-local change preserves the source's JavaScript whitespace rules and
terminal control-space rule. It leaves body extraction order, same-parser
macro registration, macro charges and existing command priority unchanged.
No shared reader or paired-delimiter name validation changes in this fix.

The explicit inventory contains 1,124 distinct input/display pairs: 884 author
controls, 224 nonoverlapping independent controls, 20 canonical-name controls
with six duplicates, and two upstream AMS controls. All original observations
are retained. There are 1,104 complete original SVG assertions: 682 fixes and
422 unchanged controls against merged main160. Four additional error controls
match after escaping only the original invalid `data-mjx-error="Misplaced &"`
attribute. Their test checks the exact single replacement and parses the entire
result as XML. The original attribute strings are preserved unchanged.

The other 16 observations are raw, not passing references: four valid-original
BOM-before-star inputs, eight NEL-before-name/star diagnostics, and four
original NEL runtime exceptions. The BOM/star and NEL/name cases expose the
separate shared GetStar whitespace behavior. Target and canonical slash-name
originals and candidate outputs agree; none of the changed raw targets had a
valid baseline rendering. Runtime exceptions are retained without claiming a
usable original SVG. Lone-surrogate JavaScript strings are outside the valid
UTF-8 input inventory; their separate transport audit is not asserted here.

Coverage includes bare/slashed ASCII, BMP and supplementary Unicode names;
unused names that are not a single callable token; stars, limits and scripts;
scoped fonts/styles; arrays and fractions; pending function, position and
script owners; active operator and paired definitions in both orders; and
missing/unbraced argument diagnostics. Effective budgets use supported
DeclareMathOperator definitions with explicit lexical separators. No unavailable
`def` or `newcommand` input is counted as macro-budget evidence.

An overlap scan of 405 published JSON files on merged main160
`2f39558e522906da2ab6fac6e4060b90736d3a99` found no prior input/display pairs
among these 1,124 observations. Passing first-publication counts are therefore
1,104 strict SVG references and four separately qualified safe-XML controls;
the 16 raw observations are not included in those passing counts.

The wider replay covers 5,634 published raw inputs and 4,326 upstream inputs.
It finds no genuine published changes or regressions and two exact upstream
fixes, both included in the strict fixture. Two apparent cancel changes are
only attribute serialization order: 64 renders from each binary per input
retain identical complete parsed XML and are excluded from fix/regression
counts. A later GetCS line-ending fix must additionally compose the separate
paired-name validator repair; this AMS reader alone does not claim that scope.

## Regeneration

```sh
node --jitless testdata/generate_ams_declaration.cjs PINNED_ASSETS
go test ./... -run TestAMSOperatorDeclaration
```

The generator verifies all three frozen asset hashes and invokes only the
original runtime. It preserves strict SVGs, raw errors/exceptions and canonical
original controls. No reference SVG is derived from Go. The input inventory is
explicit; regeneration does not filter observations by candidate results.
