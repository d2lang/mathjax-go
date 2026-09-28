# Final items delivered to unbraced scripts

The pinned original source is MathJax 3.2.2 commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`.
[TexParser.Push/PushAll](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/TexParser.ts#L191)
delivers inferred-row children separately, and
[BaseItems.SubsupItem](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/base/BaseItems.ts#L235)
accepts the first completed MML item. The remaining items are then delivered
to the caller, after the script has completed.

Previously the Go script consumer grouped every returned node into one
inferred row and put all of them in the script. For example,
`\Huge x^\comm*{a}{b}+y` put the entire fixed commutator in the exponent.
The original scripts only its first delimiter; `a,b]` stays on the baseline.
Both the superscript witness and `\Huge x_\acomm*{a}{b}+y` now match the
complete original SVG.

The consumer now carries the first node and trailing nodes separately. Both
ordinary and prime script attachment paths finish the script and reduce any
pending position before delivering the trailing nodes through the ordinary
row append path. Any AutoOpen continuation starts afterward. An empty
PushAll still leaves the script pending, while an actual empty MML node can
supply the argument. Source-produced Fn/Not/Dots/Position items retain their
existing rejection rules. The separate Bqty registration is not part of
this prerequisite.

The 678-input inventory contains 646 complete exact original references:
622 valid renderings and 24 error renderings. Against main145
`aa6200c004e8498321d821ac7edceb4040b441cb`, there are 166 fixes and 480 exact
controls. Coverage includes starred and sized commutators, vector fonts,
empty arguments, primes, later scripts, parent functions, positions,
negation, dots, Braket, AutoOpen, font declarations, authored groups,
matrices, CD, and command overrides. No formerly exact input in this fresh
inventory regresses.

All 32 nonexact observations remain raw in `script_final_items_residuals.json`:

- 22 starred Bra cases retain the existing extra empty TeXAtom produced by
  Go's extra `{}` in the starred expansion. Ten script cases change; for
  each one, removing only that non-rendering empty group from the observed
  candidate string yields the complete original string, including every
  glyph transform and viewport. The other 12 outputs are unchanged. The
  pinned PhysicsMethods.Bra starred expansion has no such extra group.
- Two empty PushAll cases inside matrix cells retain an existing difference
  between MissingScript and MissingOpenForSup diagnostics.
- Eight Physics function-plus-AutoOpen cases remain wrongly accepted as MML
  instead of producing the original missing-open-brace diagnostic. Both
  baseline and candidate are wrong; this change moves their trailing MML to
  the caller but does not repair that separate Fn event classification.

The complete original, baseline, and candidate outputs are retained; none
of these cases is counted as exact or substituted into a golden. The empty
Bra group proof is in `script_final_items_geometry_proof.json`.

A replay of all 3,479 published main145 residual inputs changed one existing
Physics Fn case and two existing authored cancel attribute orders. The
latter are complete parsed-XML matches, with no geometry change. One
serialization ceased to be byte-exact in that run; it is preserved with
its attribute-order proof, rather than counted as a source regression.

Regenerate both original inventories with:

```sh
node --jitless testdata/generate_script_final_items.cjs /path/to/pinned-assets
```

The generator checks all three frozen D2 asset hashes and creates a fresh
original VM for each expression. It never runs Go. Both complete original
inventories regenerate byte-identically. Final base/gate and independent
review results will be recorded before handoff.
