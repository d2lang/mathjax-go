# Physics braket child-environment original references

The active Bra, Ket, BraKet and MatrixElement handlers parse their generated expressions in a fresh child with the caller's complete font and identifier environment. This includes multi-letter identifier grouping in MathFont commands, which affects spacing and glyph selection in expressions such as `\mathcal{\braket{TT}{TT}}`. The same child shares command registrations with its caller, so declarations made inside an operand remain visible after the braket.

The five local call sites use `parseChild` plus `unwrapInferred`. Generated expansion strings, semantic braces, star handling, followed-ket lookahead/rollback and BraKet bar cleanup are unchanged. The generic expansion helper and inactive handlers are unchanged. Each complete generated expression uses one fresh macro counter shared by its operands, while the caller's budget is preserved.

BraKet's separate optional-second-argument reader also follows the original `GetNext`: it commits JavaScript whitespace, then reads an argument only at a literal `{`. U+FEFF is consumed before that brace; U+0085 is retained for the ordinary parser. The explicit source `noneOK=true` argument contract is preserved.

## Public fixture and source failures

`physics_braket_env_mathjax_3_2_2.json.gz` contains 1,760 distinct TeX/display pairs with every original response preserved:

- 1,748 complete, unmodified SVG assertions: 1,544 valid renderings and 204 error renderings.
- 12 original NEL runtime failures, separately asserted as `no Unicode range for character U+0085` with empty SVG. Each retained full original exception starts with `TypeError: Cannot read properties of null (reading '4')` and belongs to an input containing literal NEL.

The candidate fixes 286 complete SVGs and preserves 1,462 exact SVG controls, with no regression or changed nonexact SVG. Of those fixes, 272 exercise the copied child environment or shared registrations; 14 exercise the BOM optional-argument lookahead. The 12 source runtime cases previously rendered successfully and now return the bounded error. Those changes are separate from SVG fixes and do not claim original exception-message parity. There is no omitted raw SVG partition.

The audit reuses 1,104 earlier complete originals and captures 656 more. Family tags overlap; aggregate counts deduplicate TeX/display pairs. All 180 explicit BraKet whitespace/star controls remain in the fixture, including the distinct BOM and NEL outcomes. The fixture also retains active alias overrides, child declarations, scripts, delimiter redefinitions, raw-bar cleanup and followed-ket rollback cases.

## Passive source observations

The three evidence containers preserve 220 distinct inputs already present in the public fixture:

- `physics_braket_env_font_observations.json.gz`: 144 complete parser-entry/return observations, including the copied font, multi-letter identifier pattern, `noAutoOP` and the child's zero initial macro count.
- `physics_braket_env_budget_observations.json.gz`: 12 complete observations of Bra, Ket, BraKet and MatrixElement. Bra/Ket render at 999 and 1000 substitutions and error at 1001. Two BraKet operands share totals of 998 and 1000 before the over-limit case; three MatrixElement operands likewise share one counter. Successful returns retain the caller's original counter.
- `physics_braket_env_independent_observations.json.gz`: 64 independent font intersections, comprising 48 raw-bar cases and 16 followed-ket rollback cases.

The passive observer calls each original `Parse` and `mml` method once, preserving the returned tree. Every observed output equals its uninstrumented original. The 144 font target-to-literal bindings use 80 distinct literal pairs already in the public union. Four additional calligraphic witness/literal bindings and 16 registration-sharing controls also belong to that same union. These observations and bindings add no public inputs or duplicate assertion counts.

## Original-only regeneration

The source is the frozen MathJax commit `ad8f5c21cb810236551da8c6512ba733e67357ee`. The three pinned D2 asset hashes are embedded in the public fixture.

```sh
node --jitless testdata/generate_physics_braket_env.cjs /absolute/path/to/oracle-assets /absolute/path/to/scratch-output
```

Use a separate scratch output directory. Only the hash-verified original supplies outputs. The generator reproduces all 1,748 full SVG strings, retains all 12 complete historical runtime objects, and records fresh full exceptions while checking their null-range first line and literal NEL input. JSONL transport escapes line/paragraph separators. It repeats all 220 passive records, requiring unchanged complete event objects, complete original outputs, and equality to the uninstrumented render. All four gzip files regenerate byte-identically; their hashes and the regeneration receipt are recorded in the inventory.

The public Go test requires unique names and input pairs, the source pin, complete SVG strings, well-formed original XML and the exact bounded-runtime contract. No expected SVG comes from Go and no SVG assertion is normalized.

## Publication overlap and validation

The fresh overlap scan uses actual merged main177 `19657cbd098093fcedef0d4b8ceda9e667ea7b83`, whose complete tree equals published KetBra head `1994824901e61b4b933d86dc3264ef0177034d72`. Across all 498 tracked testdata JSON/JSON.gz paths, it finds 644 first-publication SVG inputs and 1,104 existing strict SVG inputs, with no raw promotions. All 12 runtime inputs are first publication. The 24 Expectation matches and one handler-oracle match are subsets of those same 1,104 GetStar strict inputs and are counted once. Every prior full original agrees; there are no conflicts or ambiguous display modes. Per-fix inventories may overlap earlier inventories and do not claim repository-wide unique coverage.

The reviewed source preview is `3eb7a1bdf671a43bd08a6a63c905dcf85d7a839e`, internal tree `be28906e9b072dac5cf0a69b7873a12e0327ad89`, compiled on the KetBra preview `33add2a28b08e59b188ef6974ad09131bb7e9592`. The focused 1,760-input public test passed. The actual177 rebase preserves the source and fixture patches exactly; rebased source is `e6db02d9df926f556508d1cf37265a597684195d` with the same internal tree. Full frozen-oracle tests, race tests, vet and WebAssembly build passed on checkpoint `41623db5f6d3c6840f193e69586865f5127989fe`. Production, tests, fixtures and evidence are unchanged by the final validation metadata commit.

Independent broad replay preserves all 17,549 original objects. The 6,800 published and 4,326 upstream inputs show no genuine SVG change or regression. Six attribute-order controls were repeated 768 times: 640 outputs from five cancel controls match complete original XML, while 128 outputs from the inherited unsupported-enclose control only match baseline and candidate XML. That control remains nonexact to the original undefined-command error and is not counted as parity. The 5,621 shared and 802 runtime complete candidate responses remain unchanged.

All 15,898 protected observations remain unchanged: 4,394 GetStar strict, 776 GetStar residual, 3,350 Quantity, 1,342 Eval, 2,444 HLine, 1,796 Expectation and 1,796 KetBra responses. These replays are not added to the unique fixture count.

The two D2 witnesses reuse public-union inputs. `physics-braket-font` restores calligraphic identifier grouping and narrows the box to the original width; `physics-braket-lookahead` places the second argument after U+FEFF inside the inner product. Both before/after comparisons were inspected at a shared scale, and both complete after-SVGs equal the original. All 91 D2 checks passed on the preview in round 63. The evidence was rebuilt on rebased public ancestor `f297e06be377952b480fac10a3a423232f72e23c` with actual main177 as the before revision. All three SVGs in each witness are unchanged, and all 91 complete original-SVG checks passed again. Both regenerated screenshots were visually inspected again; their PNG hashes differ from the earlier capture.
