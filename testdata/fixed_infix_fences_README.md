# Fixed infix fraction fences

The reference is the unmodified D2-pinned MathJax 3.2.2 asset set at source commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. The generator verifies all three asset
SHA-256 values and uses a fresh VM for every formula/mode. It observes the compiler
tree immediately before the unchanged SVG renderer; no parser or output overlays
are installed. Regenerate with:

```
node testdata/generate_fixed_infix_fences.cjs /path/to/pinned/d2latex [evidence-directory]
```

The 48 public records contain 30 `choose`/`brace`/`brack` cases and 18 controls.
Every fixed-fence case requires complete raw-primary SVG bytes (via SHA-256), all
explicit attributes and own properties, plus display measurement and repeat
render stability. Inherited display/text/script styles, nested scripts, tall
fractions, surrounding spacing and nested/outer fences are included. Private
checks preserve the original numerator identity/attributes/properties, fraction
ownership and numeric zero, the ORD/open/close row properties, the unresolved
four-way palette, and inherited selection of equal min/max delimiter sizes.

The accepted reference baseline is `06fcb3e2df8c5800d3b26cc7e1586c2ff59bd041`.
All 30 fixed-fence cases fail against that baseline and match the primary after
the repair. At the D063 acceptance point all 18 control SVGs and raw compiler trees remained
byte-identical to that baseline. D064 now upgrades the six binomial records to
untouched complete primary SVG and full trees. Across all 48 records, 48 complete
SVGs, 46 explicit-attribute trees and 40 own-property trees are primary exact;
all 42 non-binomial actual outputs remain byte-identical to accepted D063.

The exact remaining boundaries are recorded in `fixed_infix_fences_boundaries.json`:

- Two `atop` records retain the accepted explicit string `"0"` rather than the
  primary numeric `0` at the single recorded fraction attribute path. Their
  complete SVGs are raw-primary exact.
- Six `genfrac` records retain missing outer-row `open`, `close` and `texClass`
  own properties at the recorded exact paths. Their complete explicit-attribute
  trees and SVGs are raw-primary exact. No renderer behavior is inferred from a
  property-only distinction.

Metadata qualification is limited to those exact path/key/value receipts. Tests
reconstruct the complete accepted tree from the primary tree and require all
remaining fields. No broad attribute/property omission or SVG normalization is
used.

The optional evidence directory additionally receives four separate error probes:
repeated `choose` and invalid `genfrac` style, both modes. Bad style already
matches the primary. Repeated `choose` is a preexisting false success while the
primary produces `AmbiguousUseOf`; the candidate still lacks that parser error
and its newly fixed delimiters change the wrong nested-fraction paint. This is
recorded as an unresolved separate error-contract observation, not included
among successful parity assertions or fixed by this patch.

Scope is only the existing supplied-delimiter infix fraction path: reuse the
existing fixed palette, set `withDelims`, and preserve fixedFence's ORD row and
numeric zero. General `fenced`, generalized-fraction behavior, the renderer,
prime handling and unsupported delimiter commands are unchanged by D063. D064
separately repairs the three binomial registrations and their fixed-fence handler,
as documented in binomial_commands_README.md. The original D063 baseline
receipts remain archived; no primary fixtures were rewritten for the upgrade.
