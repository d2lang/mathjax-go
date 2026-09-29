# Shared GetStar and Bra source references

The two fixtures preserve **5,170 distinct original TeX/display inputs**. The held GetStar/Tag/Physics-lookahead inventory has 3,336 inputs and the independently preserved Bra inventory has 1,884; their 56 duplicate pairs have identical complete originals. Six additional D2 witness originals have no overlap with that 5,164-input union.

The final partition is **4,394 complete byte-exact SVGs** (3,388 valid renders and 1,006 source error SVGs), **30 single-attribute safe-XML controls**, **682 original runtime outcomes with a bounded API-error assertion**, and **64 unchanged valid residuals without a passing SVG assertion**. The strict partition fixes 766 inputs against merged main174 and preserves 3,628 controls. Runtime substitutes and XML escaping are counted separately from exact SVGs.

## Source scope and inherited limits

The frozen source is MathJax `ad8f5c21cb810236551da8c6512ba733e67357ee`: `TexParser.GetStar/GetNext`, `PhysicsMethods.Commutator`, `PhysicsMethods.Bra` and the original `Other` null-range failure. Shared GetStar uses the source JavaScript whitespace predicate, then consumes at most one literal star. Eval and Quantity now call that shared reader; their already-correct local behavior is preserved. Only the two source GetNext peeks in Commutator change. Bra retains a recognized following ket even when its brace lookahead fails, restores the cursor after its command name, and propagates its argument errors. The separate starred-Bra correction removes the extra group only from the starred no-ket expansion; the unstarred expansion keeps its source group. Generic token whitespace, named space/Tilde and unrelated argument/optional readers are outside this change.

The existing `internal/tex/testdata/eval_getstar_mathjax_3_2_2.json` supplies 181 original method observations with unchanged bytes (SHA256 `6410ff6ac30bb5708ecafea10ed516b2902785ac9613eeb2f178dbf15c81ad05`). The shared-reader test checks boolean result, consumed UTF-8 bytes and remaining input. The held alternative method fixture is byte-identical and is not duplicated. Two temporary legacy Go compatibility assertions were removed; they were not original observations. The 28 passive Bra child observations (58 child-mml events) remain separately hash-bound provenance, not additional public-input assertions.

The 64 valid residuals are 34 existing Expectation `ev` handler differences,6 existing paired-declaration priority cases involving `tag`/`ref`/`eqref`, and 24 outer-product/dyad/ketbra/op optional-argument or script cases. Every complete baseline output equals the candidate output. Their original and comparison objects are retained in the residual fixture; no normalization or passing SVG assertion hides these gaps.

All 682 runtime originals contain U+0085 and throw `TypeError: Cannot read properties of null (reading '4')` in the frozen source. The existing Go bounded failure is an empty SVG and the exact API error `no Unicode range for character U+0085`. This is a bounded substitute, not source exception-message or successful-render parity. Of 346 newly bounded outcomes,306 previously returned valid Go SVGs and 40 returned error SVGs. The other336 retain the same error, differing only in the private probe's omitted versus empty `svg` field. Complete historical original exceptions remain untouched.

The 30 original `Misplaced &` error SVGs are invalid XML. The separate assertion permits exactly one replacement of `data-mjx-error="Misplaced &"` by `data-mjx-error="Misplaced &amp;"`, requires full-string equality otherwise and verifies XML validity. Original bytes remain in the fixture. These30 are unchanged controls and are outside the 4,394 byte-exact count.

## Provenance and current-base checks

The source checkpoint is `8d981db6ffdebfb083bec2a6e817d7a18a50ec04`, internal tree `a5a7d0ead6f1d144630234df12d0a86f885fa37d`, on actual main174 `24acaf75285fc5bd894316642fd4cf6becc923c7`. The inventory binds compiled candidate and baseline hashes, all retained source inventories, the separate Bra350 source observations and the six new D2 references. Old failed/intermediate captures remain audit evidence; expected SVGs come only from the original.

A recursive scan of 481 tracked JSON/JSON.gz testdata files at that exact merged base finds **3,801 first-publication strict inputs,536 prior complete-SVG references,46 promotions from retained raw originals, and 11 prior inputs represented by hashes or ASTs rather than a complete original SVG**. All matching complete originals agree, with no ambiguous-only input identities. These categories concern the 4,394 strict pairs, not runtime coverage. The separate30 safe-XML assertions comprise28 first-publication inputs and 2 already asserted by TestAMSOperatorDeclarationSafeXMLErrors; those2 are existing qualified assertions, not raw promotions. Per-fix inventories may overlap existing published controls.

The current6,800-input published replay has 222 genuine complete-SVG fixes, no genuine regression and no changed valid/error-SVG residual. Sixty-six original-runtime inputs become bounded errors (54 U+0085,6 U+000B,6 U+000C); these are reported separately. Four authored-attribute ordering controls were rendered64 times per binary (512 total); their complete parsed XML equals the original, so one apparent fix and three apparent regressions are excluded from semantic counts. Original strings are never normalized in assertions.

The 4,326 frozen-D2 upstream inputs yield 10 genuine SVG fixes and no regression or changed residual. All 5,621 shared AutoOpen inputs and 802 runtime-end inputs are unchanged after narrowly qualifying the private empty-SVG response shape. The protected Quantity corpus retains all 3,332 complete original SVGs and 18 runtime outcomes; all 1,342 Eval and 2,444 HLine outputs are unchanged. These broader controls are not additional freshly captured GetStar references.

## Reproduction

Use the pinned Node runtime in jitless mode and the three hash-verified frozen assets:

```sh
node --jitless testdata/generate_getstar_bra.cjs /path/to/frozen/assets /path/to/scratch-output
```

The generator reads the explicit inputs from both checked-in gzip fixtures, launches only the frozen original oracle in a fresh context per request, and writes regenerated fixtures into the separate scratch directory. It verifies every complete SVG byte, including retained raw and invalid-XML originals; LS/PS are escaped only in JSONL transport. Runtime rows retain their historical complete original objects while fresh complete original exceptions and the common source failure contract are recorded in `getstar_bra_regeneration.json.gz`. It never runs Go, derives an expectation from candidate output, or rewrites the checked-in files in place. Compare both regenerated gzip files byte-for-byte with the originals. The existing Eval star generator separately reproduces its 212 SVGs and the unchanged181 method observations.

Original-only regeneration passed for all 5,170 inputs: all 4,488 complete original SVGs are byte-identical, all 682 historical runtime objects remain unchanged, and 682 fresh original exceptions match the source contract. Both compressed fixture files regenerate byte-identically. Independent test, generator and fixture binding passed. Full frozen-oracle tests, race tests, vet and WebAssembly build passed on `fdf3c6524f045fc310f5bf1d203b490192394bec`; the final metadata change only records those results. The focused public test also passed.

Both D2 comparisons were inspected at one shared scale. The commutator changes from a missing-argument error to the original formula; the supported operator override makes the starred-Bra script move from lower right to beneath its operator. All 85 retained D2 witnesses equal the original SVGs byte-for-byte. See `parity/getstar-whitespace` and `parity/physics-starred-bra` for source, screenshots, raw SVGs and reproduction details.
