# Physics Eval source parity

The complete fixture preserves **1,342 unique original SVGs**: 826 valid renders and 516 error SVGs. The current candidate matches every string exactly, fixing 1,008 cases against merged main171 and preserving 334 controls. Both `eval` and `evaluated`, stars, actual bar termination, caller recipients, fonts, direct Pop, overrides, error precedence and effective caller/child macro budgets are represented.

The earlier 78 focus references and 212 star references are retained unchanged. They are disjoint from each other and entirely contained in the 1,342-input complete fixture, so their repeated assertions do not add unique coverage. There are 1,632 full-SVG assertions across these three tests. The separate 181 original GetStar method observations check value, cursor and remaining input; two explicitly labeled legacy compatibility assertions are not source references.

A recursive scan of the 461 tracked JSON/gzip testdata files at exact merged main171 (`79c65b0fde0dc49fae7240e2ea743371db5837b6`) found all 1,342 inputs to be first-publication references, with no conflicting originals or ambiguous-only identities. The inventory records the measurement base and compiled source binding.

## Source and ownership

The frozen source is MathJax `ad8f5c21cb810236551da8c6512ba733e67357ee`: `PhysicsMethods.Eval`, `PhysicsItems.AutoOpen`, `TexParser.GetStar/GetNext`, and `ParseUtil.fenced`.

Braced Eval reinserts the exact expansion in the caller's input and macro budget. Raw parentheses/brackets consume their opening and deliver an AutoOpen item that closes at a bar. Pending recipients see the source non-MML event. Normal closing parses the one right-child MML in the closing environment before lexical restoration; direct SpreadLines Pop does so after restoration. Font continuations can contribute earlier content while unwinding, so only that right-child result is cached at the actual close. Direct smash-node construction after child parsing is an observed equivalence, not a claim that constructor call order is identical: passive observations found no child-time list queries, and complete source outputs bind the affected paths.

Eval uses the exact JavaScript whitespace predicate for GetStar and the following GetNext. Other existing star callers retain their current behavior through the staged generic wrapper. The one shared consumption helper owns cursor advancement. The held shared GetStar correction must consolidate the wrapper and Eval call and remove the two temporary compatibility assertions.

## Broader checks and retained limits

- All 802 runtime-end controls are unchanged. Forty-eight API responses differ only by the private harness's empty `svg` field; error text is identical.
- The refreshed published inventory has 6,768 complete-original TeX/display pairs, including the new runtime raw/proof records. Sixty metadata-only or direct-node observations are explicitly outside that public replay. There are no genuine changes after excluding two authored-attribute ordering variations and 914 response-shape-only API results.
- The 4,326 frozen-D2 upstream inputs contain 12 additional genuine Eval fixes and no semantic regressions. One apparent regression is the same inherited authored-attribute ordering variation.
- All three ordering controls were rendered 64 times per binary. Every complete parsed XML tree equals the original, including all attributes, text, tails and child order. Each raw string difference is restricted to its named single adjacent attribute swap; originals are not normalized and these cases do not count as source fixes.
- All 5,621 shared AutoOpen derivative/operator/vector outputs are unchanged, with 5,514 exact. This includes 192 fresh original recipient/font/Over/Pop controls: 158 exact and 34 unchanged residuals. Fourteen augmented helper-registration observations are excluded from the default public probe configuration, rather than misclassified as ordinary D2 inputs.

The initial 20 NEL failures and complete comparison/repetition evidence are preserved in the audit receipts. Those failures motivated the local source GetStar selection; they are all exact in the final 1,342-input comparison.

## Regeneration and validation

Run `node --jitless testdata/generate_physics_eval.cjs /path/to/frozen/assets` for the complete fixture. The generator verifies the three frozen asset hashes and asks only the original oracle for full SVGs. LS/PS are escaped only in JSONL transport, preserving the TeX supplied to the original. Use the focus and star generators for their separately retained fixtures; no generator runs Go or selects expected output from the candidate.

The short direct-method and 78+212 public checks passed on the bound candidate. Full-suite, race, vet and WASM validation is pending; the complete 1,342-reference test has not yet been run in Go.
