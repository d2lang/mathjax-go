# GetCS escaped line endings

The frozen MathJax 3.2.2 `TexParser.GetCS` uses a non-dotall regular expression.
After a backslash, LF (U+000A), CR (U+000D), line separator (U+2028), and paragraph
separator (U+2029) fail its ordinary-character arm. Its fallback consumes one
UTF-16 character and returns control-space. Go consumed the same character but
returned that line-ending character as a command name, producing undefined
commands or incorrect declaration calls.

The source change maps only those four consumed characters to control-space.
It preserves the ordinary command-map dispatch (including late supported
operator/paired declarations and the `\\text{ }` control-space macro), ASCII word
scanning, supplementary Unicode consumption, bounded terminal-backslash cursor,
and raw braced argument/text/token readers. No general whitespace skip or
Unicode command-name validation is changed.

The source base is actual merged main164
`0b179b832657fc7f97acd353d0c1f642ed8df3ae`, which includes the AMS and paired
name-reader prerequisites from main162 and main163. AMS reads a permissive trimmed name;
paired declarations validate their raw name without interpreting it through
GetCS. Before those fixes, changing GetCS alone exposed incorrect active names.
The new 752 declaration/caller references exercise both readers, all four line
endings, braced versus unbraced names, ordinary/position/function recipients,
late control-space/text overrides, and effective macro-count boundaries. All
752 now match the complete original SVG.

## Original references

All originals come from commit `ad8f5c21cb810236551da8c6512ba733e67357ee` through
D2's frozen package configuration and three hash-verified assets. The public
inventory preserves 1,600 unique TeX/display pairs: 716 existing source-audit
inputs, 752 new declaration compositions, four witness/literal controls, and
128 distinct promotions from published raw inventories. The original outputs
are retained unchanged.

`getcs_line_endings_mathjax_3_2_2.json` asserts 1,552 complete, unmodified SVGs,
including 272 original error renderings. Against merged main164, 996 of these
are fixes; there are no exact regressions. The strict fixture includes 1,382
first-publication input pairs, 36 earlier complete references, and 134 promoted
raw references. These are per-fix input counts, not globally unique MathJax
coverage. The inventory scans all 425 tracked JSON files beneath any testdata
directory, including private fixtures, and records the prior-source split.

`getcs_line_endings_residuals.json` retains 48 raw observations: 18 valid SVGs
and 30 original runtime exceptions. Of the valid outputs, 16 are byte-unchanged
from baseline: unescaped LS/PS row/text handling and escaped NBSP behavior.
The other two consume an escaped LS followed by an unescaped PS. Each has an
explicit unchanged control-space-plus-PS binding: both complete original SVGs
match, and the candidate target equals the unchanged baseline control. The
shared reader fix does not claim to repair subsequent raw-character dispatch.
Twenty-four NEL runtime observations change their Go output; the original
runtime failures remain raw and are not valid-rendering parity claims.

The 74 private GetCS observations compare returned name, cursor and remaining
source. The 308 original GetArgument observations are all strict after promoting
the six previously isolated escaped-line controls; no original value/cursor
record was altered. UTF-16 metadata is retained. Public controls distinguish
true whitespace, escaped versus raw lines, CRLF cursor ownership, braced bypass,
macro arguments, scripts/primes, pending items, cells/fences and actual supported
override/budget behavior. Unknown joined control-word controls remain diagnostic
controls, not claimed effective macros.

The prior main163 broad replay also checked all 5,758 published residual inputs and 4,326
upstream TeX inputs under D2's fixed configuration. The published replay yields
134 genuine fixes and no other semantic changes; upstream has no semantic
changes. Six cancel-attribute serialization controls were checked with 384
repeated full XML-tree comparisons (all attributes, text, tails and ordered
children); their raw variants and originals are preserved in the audit receipt.
They are not normalized into the strict fixture or counted as fixes.

## Reproduction

Run the original-only generator with the verified frozen asset directory:

```sh
node --jitless testdata/generate_getcs_line_endings.cjs /path/to/pinned-assets
node --jitless internal/tex/testdata/generate_getcs_line_endings.cjs /path/to/pinned-assets
node --jitless internal/tex/testdata/generate_getargument_whitespace.cjs /path/to/pinned-assets
```

Each public input receives a fresh VM. Generation does not run Go, select a
passing subset, or normalize an SVG. Raw original runtime stacks stay intact
while regeneration verifies their first-line exception. Run the public test
`TestGetCSLineEndingReferences` and the two private method tests to check the
complete original results.

The D2 witness contains a **literal LF after a single backslash** between two
fractions. The before output is an undefined-command error, while the candidate
is byte-identical to the complete original SVG. A backslash followed by the
letter `n` is a different TeX command and must not replace that literal LF.
