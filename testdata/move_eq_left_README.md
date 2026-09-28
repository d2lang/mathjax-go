# MoveEqLeft optional macro

The frozen D2 runtime is MathJax 3.2.2 source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`.
[MathtoolsMappings](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/mathtools/MathtoolsMappings.ts)
registers `MoveEqLeft` as the ordinary one-argument Macro
`\hspace{#1em}&\hspace{-#1em}`, with optional default `2`.
The Go registration uses the existing optional macro machinery: the expansion
is reinserted into the caller, has an ordinary caller macro charge, and delivers
its ampersand through the parser-owned array entry path. An authored star is
not a separate variant and remains caller input.

The accompanying shared Macro diagnostic change adds the invoked command name
only to `MissingCloseBracket`, matching original `GetBrackets` currentCS
ownership. Other bracket errors retain their existing behavior. User operators
and paired-delimiter declarations keep their existing priority; the six raw
paired-delimiter diagnostics described below are a separate handler boundary.

## Complete originals and qualified errors

The explicit inventory contains 1,246 distinct TeX/display pairs: 814 initial
observations, 156 supported declaration/budget controls, 128 independent SVG
review inputs, 140 independent parser inputs, and 20 frozen optional-macro
controls, with 12 overlaps. Every duplicate original agrees. There are 488 valid
original SVGs, 750 original error SVGs, and eight original runtime exceptions.
Every observation is retained; the generator does not choose cases based on Go.

The strict fixture contains 1,064 complete original SVGs, including all 488
valid original renders. Against merged PR157, these are 574 fixes and 490
unchanged exact controls. The separate residual file also has 168 qualified
safe-XML controls: 102 newly reachable and 66 unchanged. Original MathJax emits
one bare ampersand in `data-mjx-error="Misplaced &"`; Go preserves the same
message and rendered content while escaping that attribute. The test retains
the entire original SVG, permits exactly that one attribute replacement, and
parses the entire result as XML through EOF. These are not byte-exact original
assertions.

Fourteen remaining observations are not passing SVG references:

- Six inherited paired-delimiter optional-argument errors omit the active
  command suffix. Baseline and candidate outputs are unchanged.
- Eight NEL inputs raise original runtime exceptions. Their complete original
  stacks and baseline/candidate outputs are preserved without treating them as
  valid original renders.

Coverage includes default, explicit-empty and signed optional values, star and
unconsumed tokens, correct array entry and row ownership, whitespace and bracket
errors, caller/child scope, repeated uses, and dynamic command precedence.
The initial `def`/`newcommand` inputs are unavailable-command controls because
those commands are absent from the frozen D2 configuration. Effective budget
and override coverage comes from the separate supported `DeclareMathOperator`
and `DeclarePairedDelimiter` controls. The 22 official-newcommand-augmented
method observations remain outside the repository's frozen-D2 fixture counts.

The actual merged-158 fixture-tree scan found all 1,064 strict inputs and all
168 safe-XML inputs to be first complete-original publications, with no prior
full or raw overlap. This provenance base is separate from the merged-157
rendering baseline used for the saved comparison. Final current-base release
gates are recorded separately after composition.

The 4,326 upstream-input replay on the merged-157 baseline has no output
changes. The 5,086 published-residual replay has no genuine source changes or
regressions; two apparent cancel fixes are attribute-order variation, bound by
64 repeated renders per binary and complete parsed XML comparison. Their
original strings remain unchanged and they are excluded from source fix counts.

## Regeneration

```sh
node --jitless testdata/generate_move_eq_left.cjs PINNED_ASSETS
go test ./... -run TestMoveEqLeft
```

The generator verifies the three frozen asset SHA256 hashes and invokes only
the original component in a fresh runtime for each conversion. Exact SVGs and
raw original SVG/error records regenerate independently of Go, preserving LF
JSONL framing and Unicode line separators. Neither raw originals nor qualified
error strings are replaced with Go output.
