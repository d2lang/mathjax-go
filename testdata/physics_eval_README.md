# Physics Eval source parity

The complete fixture preserves **1,342 unique original SVGs**: 826 valid renders and 516 error SVGs. The current candidate matches every string exactly, fixing 1,008 cases against merged main172 and preserving 334 controls. Both `eval` and `evaluated`, stars, actual bar termination, caller recipients, fonts, direct Pop, overrides, error precedence and effective caller/child macro budgets are represented.

The earlier 78 focus references and 212 star references are retained unchanged. They are disjoint from each other and entirely contained in the 1,342-input complete fixture, so their repeated assertions do not add unique coverage. There are 1,632 full-SVG assertions across these three tests. The separate 181 original GetStar method observations check value, cursor and remaining input; two explicitly labeled legacy compatibility assertions are not source references.

A recursive scan of the 468 tracked JSON/gzip testdata files at exact merged main172 (`e866b51f45b67f8a21823871c32aecaf0e299474`) found all 1,342 inputs to be first-publication references, with no conflicting originals or ambiguous-only identities. The inventory records the measurement base and compiled source binding.

## Source and ownership

The frozen source is MathJax `ad8f5c21cb810236551da8c6512ba733e67357ee`: `PhysicsMethods.Eval`, `PhysicsItems.AutoOpen`, `TexParser.GetStar/GetNext`, and `ParseUtil.fenced`.

Braced Eval reinserts the exact expansion in the caller's input and macro budget. Raw parentheses/brackets consume their opening and deliver an AutoOpen item that closes at a bar. Pending recipients see the source non-MML event. Normal closing parses the one right-child MML in the closing environment before lexical restoration; direct SpreadLines Pop does so after restoration. Font continuations can contribute earlier content while unwinding, so only that right-child result is cached at the actual close. Direct smash-node construction after child parsing is an observed equivalence, not a claim that constructor call order is identical: passive observations found no child-time list queries, and complete source outputs bind the affected paths.

Eval uses the exact JavaScript whitespace predicate for GetStar and the following GetNext. Other existing star callers retain their current behavior through the staged generic wrapper. The one shared consumption helper owns cursor advancement. The held shared GetStar correction must consolidate the wrapper and Eval call and remove the two temporary compatibility assertions.

## Broader checks and retained limits

- All 802 runtime-end controls are unchanged. Forty-eight API responses differ only by the private harness's empty `svg` field; error text is identical.
- The current published replay has 6,800 complete-original TeX/display pairs, including all 32 newly published HLine residual inputs. Sixty metadata-only or direct-node observations are explicitly outside that public replay. There are no genuine changes after excluding two authored-attribute ordering variations and 914 response-shape-only API results. All 2,444 HLine controls (2,412 exact and 32 raw) are unchanged.
- The 4,326 frozen-D2 upstream inputs contain 12 additional genuine Eval fixes, no regressions and no changed nonexact outputs in this final composition.
- The two changed ordering controls in the current replay were rendered 64 times per binary (256 outputs). Every complete parsed XML tree equals the original, including all attributes, text, tails and child order. Originals are not normalized and these serialization changes do not count as source fixes. The earlier main171 three-control/384-render receipt remains preserved separately.
- All 5,621 shared AutoOpen derivative/operator/vector outputs are unchanged, with 5,514 exact. This includes 192 fresh original recipient/font/Over/Pop controls: 158 exact and 34 unchanged residuals. Fourteen augmented helper-registration observations are excluded from the default public probe configuration, rather than misclassified as ordinary D2 inputs.

The initial 20 NEL failures and complete comparison/repetition evidence are preserved in the audit receipts. Those failures motivated the local source GetStar selection; they are all exact in the final 1,342-input comparison.

## Regeneration and validation

Run `node --jitless testdata/generate_physics_eval.cjs /path/to/frozen/assets` for the complete fixture. The generator verifies the three frozen asset hashes and asks only the original oracle for full SVGs. LS/PS are escaped only in JSONL transport, preserving the TeX supplied to the original. Use the focus and star generators for their separately retained fixtures; no generator runs Go or selects expected output from the candidate.

All 1,632 public SVG assertions (1,342 distinct inputs), the direct methods, full frozen-oracle suite, race tests, vet and WASM build passed under pinned Go 1.27.0 with package parallelism one and the absolute pinned Node runtime in jitless mode. The four final gates took 36.068, 211.616, 0.565 and 0.866 seconds respectively.

The gates ran on composed preview `24214572ea5f6f232ac466be0d8d0ec003572df7`; rebasing onto actual main172 produced `7659d26a7e1b4757b730afbcf8165c18e91114d3` with the same complete source/fixture tree and internal tree `979e5acdd2caf4f85eb7562af0c84ef911927385`. The original per-group compressed replay receipts retain a stale baseline-source label from reused scaffolding; a separate explicit binding identifies the HLine baseline and hashes without modifying any original or response object.

All 81 D2 witnesses match the complete original SVGs. The new boxed evaluation witness in `parity/physics-eval` changes the closing-delimiter error into the source parenthesized evaluation of `x²/2` from 0 to 1; its after SVG is byte-identical to the original. The D2 block uses `||latex` delimiters so the authored evaluation bar remains part of the TeX.
