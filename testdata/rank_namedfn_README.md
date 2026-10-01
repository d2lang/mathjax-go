# Physics rank uses the original NamedFn handler

Physics maps `rank` to `BaseMethods.NamedFn` without an explicit ID. The original
handler defaults to the invoked command name and creates an `mi` with an explicit
OP TeX class, then pushes an FnItem. The Go source-map registration previously
accepted only NamedFn entries with an explicit ID, leaving `rank` on a separate
handler that set `fnOP` instead of the explicit TeX-class property. Its later
automatic operator classification changes spacing after the bare or scripted
function.

For example, `\rank+Z` has original width `8.244ex` and previous Go width
`8.872ex`. The source-map loader now accepts NamedFn's default command-name ID,
and `rank` follows the same generic handler as the other NamedFn mappings.
No argument readers, function-item lookahead, or symbol tables change.

The display and inline D2 witness inputs are included:

```tex
\rank+\rank+\rank+\rank = 4\rank
```

## Frozen references and limits

`rank_namedfn_mathjax_3_2_2.json.gz` preserves all 264 complete original API
objects from fresh frozen MathJax 3.2.2 runtimes. There are 256 valid original
SVGs and 8 rendered-error SVGs; there are no original runtime failures.

The strict equality corpus contains 262 cases: 254 valid SVGs and all 8
rendered-error SVGs. The previous Go source differs in 102 valid cases, all fixed
byte for byte by this change. The other 160 cases are unchanged controls. The
corpus covers binary operators, scripts/primes, explicit limits, lexical fonts
and colors, nested roots/fractions, child text/box parsers, matrices, pending
Not/Dots/Fn recipients, other NamedFn and Physics functions, and paired/operator
registration overrides.

The remaining two original controls (`\rank\times Z`, both display modes) are
retained in the archive's `deferred` section. They show an existing FnItem/GetForm
lookahead difference: Go emits one extra invisible U+2061 ApplyFunction node;
geometry is identical after this rank fix. They receive no SVG equality or
parity credit. `rank_namedfn_residuals.json` preserves the complete original,
previous Go, and candidate Go API objects for these two cases.

The references bind source commit `ad8f5c21cb810236551da8c6512ba733e67357ee` and the
exact frozen D2 asset hashes. The original-only regenerator checks those hashes
and retains every rendered-error result; unexpected runtime failures fail the
capture rather than dropping an input:

```sh
python3 testdata/generate_rank_namedfn.py PINNED_ASSETS NODE
```
