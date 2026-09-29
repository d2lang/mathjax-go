# Physics Eval star boundary controls

These 212 complete original SVGs bind the source `GetStar` and subsequent `GetNext` calls in `PhysicsMethods.Eval`. They preserve the first candidate's 20 NEL failures, add 160 original-first star/cursor inputs, and add 32 square-opening counterparts. Both aliases, both display modes, BOM/NEL before and after a star, LS/PS, end of input and doubled stars are covered. No reference is selected from a Go output.

The separate `internal/tex/testdata/eval_getstar_mathjax_3_2_2.json` retains all 181 original method observations, including UTF-16 cursor positions and the corresponding UTF-8 consumed bytes. The direct source test checks value, cursor and unconsumed input. The shared GetStar reader now consumes this unchanged fixture directly. The two temporary compatibility assertions for the old Go whitespace behavior have been removed; they were never original-reference assertions.

Regenerate both fixtures with:

```
node --jitless testdata/generate_physics_eval_star.cjs /path/to/frozen/assets
```

The generator verifies the three original asset hashes. It escapes LS/PS only in JSONL transport so each request remains one record; the original receives the unchanged characters. Full SVGs and method observations regenerate without normalization. The historical 78 Eval focus originals remain separate and unchanged.

Source: MathJax commit `ad8f5c21cb810236551da8c6512ba733e67357ee`, `TexParser.GetStar/GetNext` and `PhysicsMethods.Eval`. Eval now calls the shared source GetStar reader, using one star-consumption loop with JavaScript whitespace. The temporary predicate split is gone; Eval's following GetNext behavior and the original fixtures are unchanged.
