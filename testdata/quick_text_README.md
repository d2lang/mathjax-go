# Physics quick text, comma aliases, and escaped control space

`quick_text_mathjax_3_2_2.json` contains 2,386 complete unmodified SVGs from D2's frozen MathJax 3.2.2: 2,242 valid renders and 144 original error SVGs. Against main153 (`b32dec184f1f67981470a9542017efd35a3f95a1`), 1,252 are fixes and 1,134 are already exact controls. These assertions were selected from 2,394 distinct original inputs. All eight remaining inputs caused an original JavaScript runtime exception; their complete original stack, baseline output and candidate output are retained separately in `quick_text_residuals.json`. They are not passing golden references.

Ten exact references also appear in published fixtures with the same complete original SVG. `quick_text_overlap_report.json` records the recursive, case-insensitive schema scan. Per-fix inventory totals are not globally unique-input coverage.

## Source behavior

Physics defines both `qc` and `qcomma` as the ordinary macro `\qqtext*{,}`. They therefore charge one caller macro expansion and honor an overridden `qqtext`. The existing Qqtext helper must also use its source semantics: consume its optional star, read either its raw argument or its registered fixed text, then insert `(star ? '' : '\quad') + '\text{' + argument + '}\quad '` into the calling parser. Returning prebuilt text and spaces incorrectly ignores the star, embedded text/math handling, late `text` and `quad` definitions, caller pending items, and those commands' macro charges. Qqtext itself adds no macro charge and creates no child parser.

The source's terminal-backslash behavior requires a narrow cursor adjustment. `GetArgument` consumes the backslash, and `GetCS` returns a control space while advancing the JavaScript cursor once past EOF. Qqtext inserts at the clamped string end but resumes at that retained cursor. The Go helper records this virtual advance only when the actual unbraced argument begins at the final backslash, keeps the physical cursor in bounds while inserting, then resumes one byte into the generated ASCII prefix. Preset text and braced arguments do not take this path. Original controls distinguish the returned argument value from this cursor behavior, including leading JavaScript whitespace, surrounding source, overrides and other argument-reinsertion callers.

Finally, Base maps escaped control space (`\ `) to the ordinary macro `\text{ }`. Its former direct NBSP token happened to agree in common cases but lost the macro charge and late `text` binding, including source-valid Qqtext EOF compositions. This registration uses the existing macro path. Named `\space` and `~` remain the separate Tilde handler, as in the source.

Primary sources at MathJax commit `ad8f5c21cb810236551da8c6512ba733e67357ee` are `physics/PhysicsMappings.ts` (quick comma registrations), `physics/PhysicsMethods.ts` (Qqtext), `TexParser.ts` (GetStar, GetArgument, GetCS), and `base/BaseMappings.ts` / `BaseMethods.ts` (escaped control-space Macro versus Tilde). The fixture binds the exact three frozen D2 asset hashes.

## Coverage and residuals

The inventory covers both aliases, `qqtext`/`qq`, all 20 existing fixed-text commands, star behavior, raw/embedded math arguments, caller text fonts and math styles, colors, arrays, all relevant pending items, unbraced and following scripts/primes, macro and paired-delimiter overrides in both declaration orders, late helper overrides, actual near-limit caller macro budgets, and fresh internal-text math child budgets. It includes direct escaped-space/literal/named-space/Tilde controls across fonts, styles, scripts and scopes; EOF, control-word delimiters, BOM, NEL, LF, CR, U+2028 and U+2029; and unchanged controls for other argument/reinsertion consumers.

There are no changed valid nonexact SVGs in this inventory. The eight raw inputs use U+0085 NEL around Qqtext arguments; the frozen JavaScript runtime raises a TypeError. Go returns a bounded SVG instead. Those original failures remain unmodified evidence, not normalized or silently omitted.

A separate replay of all 4,638 published raw inputs found only two serialized cancel-attribute-order differences. Complete parsed XML equals original, baseline and candidate for both, and 48 repeated renders per binary reproduce both attribute orders. Thus that replay has no rendered/source regression; it does have two nondeterministic byte differences. `quick_text_published_replay.json` preserves the complete qualification and serialized observations.

## Regeneration

```sh
python3 testdata/generate_quick_text.py /path/to/frozen/assets /path/to/node
```

The generator verifies all asset hashes, runs Node with `--jitless`, creates a fresh original runtime for every input, and rewrites only original results in the fixed inventories. It neither invokes Go nor filters cases by candidate success. JSONL is split only on literal LF, preserving U+2028 and U+2029 inside strings. The SVG tests compare every complete original SVG without normalization.
