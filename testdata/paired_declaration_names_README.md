# Paired-delimiter declaration names

The three Mathtools declaration forms now accept source-valid bare names and reject source-invalid names before reading delimiter/body arguments. For example, `\DeclarePairedDelimiter{foo}{(}{)}\foo*{\frac{x}{y}}` renders the same fenced fraction as its canonical `\foo` declaration. Previously the bare name was rejected. Supplementary-scalar names and braced escaped-line names now produce the original name-validation error.

The implementation follows `NewcommandUtil.GetCsNameArgument` in MathJax commit `ad8f5c21cb810236551da8c6512ba733e67357ee`: GetArgument, ParseUtil.trimSpaces, removal of one optional leading backslash, then the non-Unicode JavaScript `/^(.|[a-z]+)$/i` validator. The validator accepts one BMP code unit other than LF/CR/U+2028/U+2029, or an ASCII-letter word. Source trimming includes BOM, excludes NEL, and preserves one terminal ASCII control-space. This reader validates the complete name; it does not tokenize it with GetCS.

The AMS declaration prerequisite must precede this change: HandleDeclareOp has a different, permissive name contract and now uses its own reader. The final publication base is actual merged main162, `bba94fc9e5a95e80ca2cb71cd845f82f3350fc6f`; its production tree is identical to the reviewed AMS preview. The isolated GetCS escaped-line fix is not included. The shared strict helper also serves augmented newcommand definitions; frozen D2 does not enable newcommand, so those callers are not claimed as public oracle coverage.

## Complete original references

The deduplicated inventory contains 2,168 public TeX/display pairs: 1,688 source and Unicode controls, 144 optional-star composition controls, and 336 independent review inputs. Every original output is retained unchanged.

- 2,052 complete original SVG assertions: 1,094 valid renderings and 958 error renderings. Against the reviewed AMS baseline, 1,092 become exact; no formerly exact result regresses.
- 116 raw outcomes, unasserted as passing SVGs: 70 original-valid SVGs, 8 original error SVGs, and 38 original runtime exceptions. Of these, 28 change: 6 valid BOM/star compositions, 4 invalid parameter-count diagnostics, and 18 NEL runtime cases. All 28 have full literal proofs: original bare-name target equals original canonical slash-name control, and candidate target equals the unchanged baseline control.
- The other raw controls include 44 complete literal-bound inherited MathOp wrapper cases, 12 escaped-LF invocations awaiting GetCS, and 2 unchanged text-control-symbol inputs. Original runtime exceptions do not count as valid representations.
- 149 direct original name-handler observations test returned names and errors after argument extraction, including the original one-versus-three argument-read boundary. Twenty-eight prior observations are retained unchanged; 121 are new. Twelve non-BMP names are rejected. No lone-surrogate input is asserted as equivalent Unicode.

The 70 earlier malformed UTF-16 transport controls and all 60 earlier mixed AMS/paired observations remain untouched in the external source audit. They are distinct from the valid-scalar rendering contract here.

## Prior coverage and replays

The inventory scans all 413 tracked testdata JSON files, including internal directories, at actual merged main162. Among the 2,052 strict input pairs, 42 already had full references, 12 promote previously raw original inputs, and 1,998 are first-publication pairs. Across strict and raw rows, 2,102 of 2,168 are first-publication pairs. No stored original mismatch or ambiguous-only prior reference was found. These per-fix counts are not globally unique MathJax coverage.

The 5,634-input published-residual replay finds 12 genuine fixes and no changed nonexact output or regression. Four apparent cancellation serialization fixes are only attribute order: 32 renders per binary per input preserve complete parsed XML, with all raw variants retained. The 4,326 upstream TeX input replay is unchanged. Exact SVG assertions do not normalize attributes, geometry, diagnostics, or output strings.

## Reproduction

Both generators run only hash-verified frozen original assets. The public generator starts a fresh oracle VM for each input in bounded 24-input subprocess batches. It refreshes only original SVG/output fields, including literal controls; it does not select the passing subset or update baseline/candidate receipts. The private generator observes the original paired handler with supplied arguments.

```sh
node --jitless testdata/generate_paired_declaration_names.cjs /path/to/pinned-assets
node --jitless internal/tex/testdata/generate_paired_declaration_names.cjs /path/to/pinned-assets
go test -run 'TestPairedDeclarationNameReferences' .
go test -run 'TestPairedDeclarationNamesPrimaryMethod' ./internal/tex
```

The exact/raw partition and publication-base metadata are rebound to actual merged main162 without changing original outputs. The 149-method short test and preview probe passed. Final current-base gates are recorded in the handoff receipt.
