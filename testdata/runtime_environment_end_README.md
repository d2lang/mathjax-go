# Runtime environment-end delivery

The fixed D2 MathJax package set gives `spreadlines`, `numcases`, and `subnumcases` executable end handlers. The original reads their physical `\end` token through the active command map, charges its BeginEnd macro budget, and delivers the resulting stack operation before pending items reduce. Capturing their raw bodies in advance lost that ordering: overridden or argument-consumed `\end` tokens could be mistaken for an environment boundary, while executing a true end missed the original macro charge.

This change keeps those three environment bodies on the executing parser input. Ordinary EndItem and SpreadLines' direct Pop have separate delivery rules. Real argument parsers start without an outer close owner; same-input groups, styles, scripts, and array continuations retain it. Braket and Physics AutoOpen still run their fenced `toMml` conversion when popped. SpreadLines only changes a direct table or immediate inferred-row children, preserving the spacing of fenced and scripted tables.

A popped unfinished script is registered only when actually published as MML. Its clones follow the existing `copyNode` preorder registration. After successful parsing, the original parent-chain live-list predicate excludes discarded nodes; remaining one-sided generic scripts become the proper compact subtype before inheritance. An unrepaired live script returns a bounded conversion error at that phase. Missing SpreadLines dimensions fail when the first actual Pop selects a table with truthy rowspacing, before later tokens can replace the error. Neither check is a renderer recovery or an EOF macro charge.

## Frozen evidence and counts

The source version is MathJax 3.2.2, commit `ad8f5c21cb810236551da8c6512ba733e67357ee`, under the exact D2 package setup. The generator verifies the three bundled asset hashes recorded in each fixture. It executes the original JavaScript in fresh VMs; it never uses Go output to generate SVG goldens.

The 806 captured observations deduplicate to 802 explicit TeX/display pairs. Duplicate complete originals must agree. The initial candidate comparison used main169 `dcdb525519b6a36b83b0eb2af303174e7bae6021`. Replaying all 802 originals against merged main170 `8393014a3f2dd05d0b2881141c85a8153ddd333e` leaves every baseline output unchanged. The same counts apply:

- 670 strict complete SVG references: 286 valid renderings and 384 original TeX-error renderings.
- 506 baseline failures become exact: 206 valid renderings and 300 error renderings. The other 164 strict cases are unchanged exact controls.
- 132 raw references: 16 valid SVGs, 62 error SVGs, and 54 original runtime exceptions.
- Zero formerly exact regressions and zero candidate process failures across all 802 inputs.

The recursive overlap scan covers all 453 tracked JSON files under every testdata directory at the exact main170 published base. It found all 670 strict inputs to be first publications and no ambiguous display identities. Per-fix inventories can overlap future publications; these figures are bound to this base, not a cumulative unique-coverage claim.

`runtime_environment_end_inventory.json` retains original inventory hashes and the origin index of every observation. `runtime_environment_end_repair_proof.json` preserves all 12 old prototype process failures alongside their complete original and repaired outcomes. Eight of those inputs are original-valid and now strict SVG references; the four others are original runtime failures and now bounded API errors. The initial failed batch and complete intermediate revisions remain in the task audit, never substituted for primary references.

## Raw outcomes and API boundaries

All 52 changed nonexact outputs are original runtime exceptions. Eighty raw outputs are byte-unchanged from main169: the 78 nonexact SVG/error-SVG cases and two runtime substitutes. There are no changed original-valid residuals in this corpus.

The 54 source runtime outcomes deliberately remain raw, with unmodified original stacks. Forty-eight now return bounded API errors: 34 at first-Pop spread conversion, 12 at successful-parse script cleanup, and two for the inherited empty NumCases table. Six retain bounded error SVGs for orphan fresh-child SpreadLines. These are explicit bounded-error deviations from JavaScript exceptions, not exact-render claims. The public test checks the documented 48 API contracts and the six error-SVG boundaries; it does not turn observed Go SVGs into goldens.

The 78 unchanged nonexact SVG cases document the separate ordinary-environment raw-capture, command-Matrix, and Physics delimiter-scanner interception seams. They include swallowed/overridden close tokens and existing wrong acceptance or diagnostic ordering. This implementation does not claim to make all environment parsers streaming.

`runtime_environment_end_cleanup_proof.json` records 90 source phase observations. They establish cleanSubSup before inheritance, later-script repairs, ignored-AutoOpen discarded views, one-sided replacement identity, and copied-node ownership. `runtime_environment_end_spread_proof.json` records 36 direct source-method controls: a falsy rowspacing assigns undefined without calling the dimension parser; a selected table with truthy rowspacing fails on a missing dimension. Explicit nil models the absent effective value in that narrow Go helper branch; this is not a general claim about authored undefined-value rendering.

## Broader comparison

The current main170 published residual replay covers 6,574 inputs: 82 genuine exact fixes, no genuine regression, and no genuine changed nonexact output. Exactly 866 API responses differ only by protocol shape: baseline has the exact keys `{error, svg}` with empty SVG; the private candidate probe has only `{error}`, with identical error text. The raw objects are preserved; no nonempty SVG is admitted to that qualification.

The union of both final-base comparison rounds contains five changed cancel inputs, verified again on the final candidate with 640 repeated complete renders. Only the single adjacent data-padding/data-thickness attribute order changes; full XML trees agree, including attributes, text, tails, and ordered children. The 4,326-input frozen-D2 upstream replay has two genuine fixes and no genuine regression/changed residual. One inherited `\enclose{circle}[data-thickness=10,data-padding=5]{x}` output varies only in those adjacent authored attributes, verified with 128 repeated complete renders and full baseline/candidate XML equality. Both outputs remain explicitly nonexact against the original undefined-command error. This replay uses the frozen D2 configuration, not upstream Jest configuration. Serialization variations are excluded from source-fix counts and are not normalized in regression assertions.

## Regeneration and verification

From the repository checkout:

```sh
node --jitless testdata/generate_runtime_environment_end.cjs /absolute/path/to/pinned-d2-assets
```

The strict, raw, old-panic, cleanup, and direct-method original files regenerated byte-identically. Runtime primary stacks must reproduce exactly. For the direct-method observer, the historical raw stack is retained while its exception first line and every other result field are verified; observer filenames are not treated as source behavior. The generator contains no broad SVG/error normalization.

The source-focused cleanup and spread-selection tests, all 670 complete SVG tests, and all 54 bounded runtime checks passed on main170. All 802 complete original/baseline/candidate objects are unchanged from the earlier main169 comparison. The independent 124-input phase/output repeats also passed. The initial complete source/fixture patch rebased onto main170 without conflict or content change. The first full gate then exposed three direct NumCases test callers missing the active Begin owner and a transposed source diagnostic identifier. The tests now supply the original owner context while preserving all prior body, exact-shape, cursor, and copy-registration assertions. Production now uses the exact `MissingBeginExtraEnd` identifier. Independent original captures verify zero counter increment for the built-in extra end at counts 0 and 1000 and the Cases opening/closing ownership. The original full failure and those primary observations remain preserved.

The final source checkpoint `621a31b10748f5161cfbdac50249e23f3fe41d22` passed the complete frozen-oracle suite, race, vet, and WASM build sequentially. Oracle execution used the pinned assets, full mode, and an absolute Node runtime with `--jitless`. All 802 public outcomes are unchanged after the diagnostic correction; all five original/proof fixture files retain their byte-identical regeneration hashes. `runtime_environment_end_inventory.json` binds the exact source, binary, comparison, serialized-output qualifications, initial failure, and final gate receipts.

## D2 visual checks

The two shared-scale comparisons in `parity/runtime-environment-ends` and
`parity/runtime-braket-pop` show corrected nested matrix spacing and a formerly
rejected environment-end/Braket expression. Each includes before, fixed, and
original MathJax SVGs and Chromium screenshots. Both fixed D2 SVGs equal their
original references byte-for-byte, and the same candidate D2 binary reproduces
all 79 committed D2 witnesses exactly. Both screenshots were visually inspected.
