# Physics Eval star boundary controls

These 212 complete original SVGs bind the source `GetStar` and subsequent `GetNext` calls in `PhysicsMethods.Eval`. They preserve the first candidate's 20 NEL failures, add 160 original-first star/cursor inputs, and add 32 square-opening counterparts. Both aliases, both display modes, BOM/NEL before and after a star, LS/PS, end of input and doubled stars are covered. No reference is selected from a Go output.

The separate `internal/tex/testdata/eval_getstar_mathjax_3_2_2.json` retains all 181 original method observations, including UTF-16 cursor positions and the corresponding UTF-8 consumed bytes. The direct source test checks value, cursor and unconsumed input. Two explicitly labeled compatibility assertions preserve the old generic Go wrapper while Eval uses the source predicate; those are not original-reference assertions and must be removed with the shared GetStar correction.

Regenerate both fixtures with:

```
node --jitless testdata/generate_physics_eval_star.cjs /path/to/frozen/assets
```

The generator verifies the three original asset hashes. It escapes LS/PS only in JSONL transport so each request remains one record; the original receives the unchanged characters. Full SVGs and method observations regenerate without normalization. The historical 78 Eval focus originals remain separate and unchanged.

Source: MathJax commit `ad8f5c21cb810236551da8c6512ba733e67357ee`, `TexParser.GetStar/GetNext` and `PhysicsMethods.Eval`. The staged implementation keeps one star-consumption loop, preserves generic caller behavior, and lets Eval select JavaScript whitespace. Once shared GetStar is corrected, route Eval through the generic wrapper and remove the temporary predicate split.
