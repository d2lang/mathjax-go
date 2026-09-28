# AMS variable limit macros

The original is D2's frozen MathJax 3.2.2 component at
`ad8f5c21cb810236551da8c6512ba733e67357ee`.
[AmsMappings](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/ams/AmsMappings.ts#L73)
registers four zero-argument macros. Each creates a math operator around an
explicit `mi` token containing `lim`, decorated with an underline, overline,
right arrow, or left arrow. They are distinct from ordinary `liminf` and
`limsup`, whose named-operator thin spaces already match the original.

The Go implementation previously treated `varliminf` and `varlimsup` as named
functions and did not register `varinjlim` or `varprojlim`. Restoring the source
macro bodies preserves ordinary macro charges, same-parser reinsertion, and
late lookup of the helper commands. Dynamic operator and paired-delimiter
definitions retain their existing precedence. This change removes the two
incorrect named-function fallbacks.

The shared UnderOver prerequisite supplies the source script-style lookup and
stretch arithmetic. Before that correction, the arrow variants retained
inherited decoration differences. The final composed candidate compares all
1,536 preserved inputs directly with their complete original SVGs: 1,432 valid
formulas and 104 original error renderings. There are 764 fixes, 772 unchanged
exact controls, and no unresolved outputs in this inventory.

Coverage includes all four commands, scripts and explicit limits, display and
inline modes, scoped fonts and styles, fractions and arrays, ordinary named
limits, source-body literal controls, active command replacement in both
declaration orders, and effective macro-budget boundaries using the supported
`DeclareMathOperator` command. The original input strings and SVGs are preserved.

The overlap scan inspected all 396 JSON files beneath `testdata` directories
on merged main159 `c19a00c2202542652ffa46ce0a93d17c0bdd8108`. It found 978
previously unpublished input/mode pairs, 554 existing complete-original
references, and four previous raw observations now promoted to strict tests.
All previous original strings agree. These are per-fix counts, not a globally
unique or exhaustive MathJax coverage claim.

The broader replay covers 5,270 published raw inputs and 4,326 upstream
TeX/display inputs under the frozen D2 configuration. It finds four genuine
AMS fixes in the former and eight in the latter, with no genuine regressions
or changed unresolved renderings. All twelve have been added to the strict
fixture without changing their original SVGs. Four apparent cancel fixes and
one enclose observation differ only in attribute serialization order. Both
binaries emit both orders over 64 repeated renders per affected input, with
identical complete parsed XML. The frozen bundle does not register enclose;
its original error remains an inherited configuration difference.

These rendering counts were measured against merged main158
`f72d955662476fe3daedc3875accdf7315a03168`. The source has since been rebased
onto main159 with an unchanged production delta; final composed build and
full checks are pending.

## Regeneration

```sh
node --jitless testdata/generate_ams_variable_limits.cjs PINNED_ASSETS
go test ./... -run TestAMSVariableLimitReferences
```

The generator verifies all three original asset hashes, invokes only the
original component in a fresh runtime for every input, and preserves JSONL
framing across Unicode line separators. The test compares the complete SVG
without normalization. No expected SVG is produced by Go.
