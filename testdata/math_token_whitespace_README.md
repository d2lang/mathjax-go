# Math token whitespace and fallback token kinds

The pinned MathJax 3.2.2 parser dispatches each ordinary math character through
its character maps. Its Base `Space` handler ignores exactly ASCII space, tab,
CR, LF, and NBSP. This is different from `GetNext`, whose command lookahead
uses JavaScript whitespace. Go's ordinary row and pending-script token loops
used Unicode whitespace, removing actual math tokens such as EM SPACE, OGHAM
SPACE MARK, LINE SEPARATOR, and PARAGRAPH SEPARATOR.

The three token-delivery loops now use the source's five-character map.
Initial script lookahead still uses `GetNext` semantics. Generic argument,
bracket, star, and control-sequence readers are unchanged.

Correctly delivering OGHAM also requires the original `BaseConfiguration.Other`
token factory: the final fallback uses the pinned Unicode range's token kind,
sets the creating font before the vector factory and range-variant override,
and marks only `mo` tokens for `fixStretchy`. Earlier explicit mappings, digit
and letter paths, and command-authored token factories remain separate. This
also corrects existing fallback characters such as U+180E and Braille `mtext`,
including authored fonts; generic font lowering is unchanged.

For characters with no source Unicode range, original `Other` throws while
accessing `range[4]`. Go returns a bounded conversion error naming the codepoint.
This guard belongs only to actual final `Other` delivery. It does not intercept
characters retained as literal text or command-authored token data. It prevents
new malformed SVG from VT/FF without pretending that the original renders them.

## Original inventory and validation

All references come from original commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`, D2's frozen package/options setup,
and the three hash-verified original assets. The inventory contains 3,440 unique
TeX/display pairs from 3,480 observations: 2,376 whitespace/caller observations,
240 token-owner observations, and 864 separately captured fallback-kind controls.
No original SVG, attribute, exception, or input is normalized.

The source is rebased onto actual main165
`5fd30978e6a93afce774885b2bf669caf8f4fe99`. The current short probe preserves the
GetCS change from that base and matches 3,032 complete original outputs. Of
these, **3,016 are well-formed SVGs** asserted strictly: 2,954 non-error renders
and 62 error renderings. There are 1,558 fixes and no exact regressions against
the merged165 baseline. The other 16 identical output strings are malformed
original XML and remain raw, never strict SVG assertions.

`math_token_whitespace_residuals.json` retains all 424 other observations:
194 original non-error SVGs, four original error SVGs, and 226 original runtime
failures. The 194 include 12 XML-invalid literal-text/lap/mmlToken VT/FF outputs;
the four error SVGs contain original unsafe ampersands. All 16 malformed originals
are retained byte-for-byte. The 204 actual null-range failures have a separate
bounded-error test; they are not original-SVG parity successes. The remaining
22 original runtime failures keep their previous Go outcomes.

The strict provenance audit scans all 430 prior testdata JSON files on actual
main165. Of the 3,016 strict pairs, 2,978 inputs first appear here, 22 have prior
complete references, 10 promote prior raw references, and six have prior input
metadata without a complete original SVG. These are per-fix input counts, not
a claim of globally unique MathJax coverage.

Controls cover ordinary rows, groups, initial and pending scripts, primes,
MathFont children, font declarations, styles, color, fractions/root indices,
arrays/CD, macros and supported declaration budgets, functions, Not/Dots,
positions, Braket/Quantity, Nonscript, and literal text/token readers. Whitespace
characters in text, comments, or arguments are retained as ownership controls;
they are not all claimed to dispatch as ordinary math tokens.

## Retained changed-render qualifications

Twenty own original-valid render observations change but remain nonexact:

- Ten Braket/Over cases have complete original fenced-numerator literal bindings.
  All literal controls are exact. Common a/b/c glyphs never move farther from
  their original affine positions; OGHAM/BOM text attributes and transforms are
  exact. The inherited numerator-owner fence omission remains.
- Two single-Braket BOM cases retain the old wrong operand ownership from its
  command lookahead. Root metrics, paths, and positions are unchanged, and the
  retained BOM's font becomes original-correct. Forced-token and original
  `GetNext` literal controls preserve that distinction.
- Six Bra/following-ket U+180E cases retain the separate-ket ownership gap.
  Every glyph/text position and viewport is unchanged; the retained U+180E
  text now has the original italic font. The failed-lookahead repair is separate.
- Two operatorname/BOM cases retain the handler's trim discrepancy. Complete
  rendered geometry is unchanged despite the corrected token/container kind.

These qualifications are bound to the complete preserved originals and both
binary outputs. None is used as a normalized passing SVG golden.

The merged165 broad replay covers 5,802 published residual inputs and 4,326
upstream inputs under the frozen D2 configuration. Published inputs have 136
genuine fixes and no genuine regressions. Two changed dfrac/U+180E observations
now have complete original primitive geometry, retaining only an inherited
mstyle closure difference. Other changed raw observations are original runtime
failures or four enclose controls unavailable in the original frozen bundle.
Upstream has no genuine changes. Four published and one upstream cancel-family
serialization controls are excluded from source fix/regression counts: 64
renders per binary each equal complete original parsed XML, including all
attributes, text, tails, and ordered children.

## Reproduction

```sh
node --jitless testdata/generate_math_token_whitespace.cjs /path/to/pinned-assets
go test -run 'TestMathToken(WhitespaceReferences|NullRangeReturnsBoundedError)' ./...
```

The generator verifies all three asset hashes, creates a fresh original VM per
input, and regenerates only original outputs. It does not run Go or choose the
strict/raw partition. Complete runtime stacks stay intact while regeneration
checks their first-line original exception. The inventory preserves the explicit
partition and source/binary binding. The strict test checks duplicate inputs,
parses the entire original XML through EOF, and compares complete output strings.
The guard test checks bounded errors and safe codepoint diagnostics for the 204
source runtime failures, not replacement renderings.

At fixture preparation, original regeneration is byte-identical and the current
short probe/replays pass. The focused/full/race/vet/WASM gates and final D2 visual
evidence remain pending; this document does not claim those checks have run.
