# Operator-name whitespace references

This change restores `AmsMethods.HandleOperatorName` at frozen MathJax source
`ad8f5c21cb810236551da8c6512ba733e67357ee`. Its argument uses
`ParseUtil.trimSpaces`, including the terminal ASCII control-space exception.
The unstarred following-`limits` check uses `GetNext`: consume JavaScript
whitespace before saving the cursor, then rewind only a different command.

JavaScript whitespace includes BOM and excludes NEL. Ordinary math tokens,
optional-star recognition, macro dispatch priority and child-parser scope keep
their existing separate rules.

## Original references

The original capture contains 1,098 distinct input/display pairs. All 962
complete original SVGs now match: 838 valid formulas and 124 original TeX error
renders, with 164 fixes and 798 unchanged controls against main169
`dcdb525519b6a36b83b0eb2af303174e7bae6021`. A scan of 448 published JSON files
finds 960 first-publication strict inputs and two promotions of existing raw
references. Original SVG strings remain unmodified.

The other 136 original observations throw in the frozen `Other` handler because
NEL, VT or FF has no Unicode range. Go now returns its existing bounded API
error for every one; 36 previously produced a rendering. Six historical
published declaration inputs expose the same null-range behavior through
operator-name expansion and are retained with their provenance, giving 142 raw
runtime references and 1,104 total reference inputs. These are separate from
passing SVG counts. Tests bind the original exception and input codepoint,
then require an error and no SVG; they do not accept a Go rendering as a golden.

The inventory covers JavaScript whitespace and near-whitespace, both stars,
terminal control-space, followed limits/non-limits commands, nested fonts,
fractions, scripts and arrays, active overrides, declarations and real
998/999/1000 invocation boundaries. All inputs were specified before the
candidate comparison; the six historical observations already existed in
published original fixtures.

The old handler-classification tests pinned Go cursor positions that retained
leading whitespace. Their eight original inputs now bind to 32 freshly captured
original cursor observations, including following-command and Unicode controls.
The observer preserves the complete uninstrumented original output and records
both UTF-16 and UTF-8 positions; all existing class/property assertions remain.

## Broader replay and regeneration

Current replay covers 6,438 saved published inputs and 4,326 upstream
MathJax-Tests TeX/display inputs in D2's frozen configuration. It finds two
complete SVG fixes, six newly bounded original-runtime failures and no genuine
regressions. The upstream outputs are unchanged. Four cancel serialization
controls are excluded from both apparent fixes and regressions after 512
repeated complete XML comparisons, including attributes, text, tails and
ordered children. Raw strings remain preserved.

Run `node --jitless testdata/generate_operatorname_whitespace.cjs /path/to/assets`.
The generator verifies all three pinned asset hashes, then renders the saved
inputs with the original bundle in fresh VMs. It invokes no Go code and retains
complete original SVGs and exceptions. Both fixtures regenerate byte-for-byte.

The D2 witness contains a literal BOM between `\operatorname{lim}` and
`\limits`. Its before/after/original SVGs, shared-scale screenshots, hashes and
reproduction instructions accompany the completed validation record.

## Validation

Focused tests, the full original-oracle suite, race checks, `go vet`, and the
WebAssembly build passed. The first full run exposed six historical cursor
assertions that expected Go's old whitespace retention; the 32 original cursor
observations above replace those assumptions. Production code was unchanged
for that test correction and the subsequent successful full run.

The D2 screenshot shows the original failure (an error and a black bar) becoming
the boxed limit expression. Its fixed SVG is byte-identical to the original
MathJax reference. The same fixed D2 binary also reproduces all 77 committed D2
witnesses byte-for-byte against their original references.
