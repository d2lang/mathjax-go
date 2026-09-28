# AMS boxed macro and GetArgument values

`boxed_macro_mathjax_3_2_2.json` contains 2,089 complete, unmodified SVGs from D2's frozen MathJax 3.2.2 runtime, in display and inline modes. The fixed inventory was selected from 2,308 distinct original inputs; all 219 remaining receipts are retained in `boxed_macro_residuals.json` with their original, baseline, and candidate outputs. The raw file is not a passing golden fixture.

Baseline is merged PR 140, commit `1c1f88b4fb8ba8af348b38050f18ba5dc9ed7af1`. Of the passing inventory, 462 records differed on that baseline and 1,627 were already exact. No formerly exact original regressed, and no valid original became a changed nonexact output. An additional replay of every published raw fixture covered 2,383 distinct inputs: 24 became exact, with no regressions or changed nonexact results. All 1,344 previously published FrameBox references are covered by the full suite, and the new inventory includes 12 additional FrameBox argument controls.

## Source behavior

The Base and AMS mappings both define `boxed` as the ordinary one-argument macro `\fbox{$\displaystyle{#1}$}`. Register that definition instead of constructing a menclose directly. FBox and `ParseUtil.internalMath` already implement its independent parser with an empty lexical environment, a fresh macro counter, and shared command configuration. Outer font choices no longer leak into boxed contents; explicit inner fonts still apply. The literal expansion also honors overrides of `fbox` and `displaystyle`, while the higher-priority paired-delimiter map can override `boxed`.

The macro exposed an existing argument-reader defect at a terminal backslash. `TexParser.GetArgument` returns a backslash plus the value returned by `GetCS`, rather than the consumed source slice. GetCS returns a control space at EOF and removes the optional ASCII space following a controlword. Retaining the raw slice incorrectly turned the valid `\boxed\` into a missing-close-brace error after macro substitution. The separate source commit restores this value boundary without changing ordinary tokenization, bracket scanning, or GetCS itself.

Primary sources are MathJax's `base/BaseMappings.ts`, `ams/AmsMappings.ts`, `base/BaseMethods.ts` (Macro and FBox), `TexParser.ts` (GetArgument and GetCS), and `ParseUtil.ts` (internalMath and checkMaxMacros). The frozen runtime is identified by the commit and SHA-256 asset hashes in each fixture.

## Coverage

The corpus includes outer and inner fonts, explicit empty math fonts, colors and sizes; roots and root-index offsets; nested boxes and arrays; all pending-item and script argument positions; unbraced, malformed and missing arguments; escaped math delimiters and text; label/configuration sharing; paired-delimiter and operator overrides in both declaration orders; fbox/displaystyle overrides; and command-word delimiters, EOF backslashes, BOM, NEL, LF, CR, U+2028 and U+2029.

`TestBoxedMacroCounterBoundary` checks that boxed charges the caller's macro counter, that malformed arguments fail before that charge, and that successful or failing internal math children share definitions while restoring the caller's counter.

The retained residuals cover existing general GetCS newline fallback and whitespace-tokenization differences, original runtime exceptions, the frozen runtime's unavailable boldsymbol command, a sqrt diagnostic suffix, and two existing Mathtools row-interception compositions. In the raw set, only original-error boldsymbol controls change, because the contained boxed font reset is now correct. No original-valid residual changes. These raw original outputs are never normalized into candidate output.

## Regeneration

Run from the repository root with the three verified frozen D2 assets:

```sh
python3 testdata/generate_boxed_macro.py /path/to/assets /path/to/node
```

The generator verifies every asset hash, runs the pinned Node runtime with `--jitless`, creates an independent MathJax runtime per input via `differential/oracle.mjs`, and updates only original outputs in the fixed inventory. It parses JSONL on literal LF boundaries so U+2028 and U+2029 inside JSON strings remain intact. It does not invoke Go or filter inputs by a passing candidate result.
