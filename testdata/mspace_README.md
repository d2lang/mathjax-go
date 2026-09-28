# Original mspace command

The pinned MathJax 3.2.2 source registers `mspace` as another name for
`BaseMethods.Hskip`. The Go source table retained that registration, but the
active command dispatcher omitted it. Add the missing name to the existing
`horizontalSpace` route, alongside `hskip`, `hspace`, `kern`, `mskip` and
`mkern`. The dimension parser and spacing renderer remain shared.

Primary source at commit `ad8f5c21cb810236551da8c6512ba733e67357ee`:

- [BaseMappings horizontal-space registrations](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMappings.ts)
- [BaseMethods.Hskip](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseMethods.ts#L902-L907)

Hskip reads its width through the original `GetDimen(name)` helper and pushes
an `mspace` node. It does not add a separate braced-argument requirement or
restrict the command to math units. The source dimension parsing, unbraced
consumption, signed values, font-size behavior and diagnostics are retained.

A visible witness compares ordinary operator spacing with authored spaces:

```tex
\Huge\boxed{a+b\qquad a\mspace{1em}+\mspace{1em}b}
```

## Complete original references

`mspace_mathjax_3_2_2.json` contains **1,374 complete original SVGs**:
1,004 valid renderings and 370 original error renderings. The merged main134
baseline `6800b44f34e45feca1ea427785249915a1958716` fails 752 references;
622 are already-exact controls. All 1,402 exploratory inputs are retained,
including the 28 nonexact observations below. No formerly exact input
regresses.

The inventory covers all six names of the shared Hskip handler, braced and
unbraced dimensions, all supported unit families, signs, decimals, comma
forms, empty/malformed/missing operands, dimension whitespace, scripts,
fonts and sizes, fractions/radicals, pending function/operator items,
arrays and CD, authored empty rows, and operator/paired-delimiter
redefinitions. Compositions include the restored TeX/LaTeX logos and poor
man's bold macro. Both display modes use complete untouched original SVGs. Independent review
added 346 fresh comparisons: 322 exact, 240 fixed, and zero regressions; its
338 nonduplicate inputs are included in the explicit inventories.

## Retained inherited differences

`mspace_residuals.json` preserves 28 original/baseline/candidate results
for space within a later aligned or gathered cell. Eighteen existing
`hskip`/`hspace`/`mkern` controls remain byte-identical to baseline.
Every original output is valid; the ten newly reachable Go outputs retain the
separate missing initial-operator spacing. These are not passing references.

Ten freshly rendered `hskip` controls prove the shared boundary: each
original `mspace` SVG equals its original `hskip` SVG, and the new Go
`mspace` output equals the existing Go `hskip` output both before and after
this change. Every raw response is preserved in `inheritedControls`.
The source initial-operator correction remains separate from this command
registration.

Regenerate both exact and raw references, including the inherited controls:

```sh
python3 testdata/generate_mspace.py /path/to/pinned/assets /path/to/node
```

The generator verifies all three frozen D2 asset hashes, uses a fresh VM per
expression in batches of 24, and never runs Go or replaces an original with
a candidate SVG. Run `go test ./... -run TestMspaceReferences` for the complete
SVG assertions.
