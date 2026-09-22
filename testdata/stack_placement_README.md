# Direct Overset/Underset dispatch

These 50 references come from fresh VMs of unchanged MathJax 3.2.2 assets pinned by the generator to D2 v0.8.1 and MathJax-src ad8f5c21cb810236551da8c6512ba733e67357ee. The complete original SVG hashes, dimensions and explicit-attribute/content trees are retained without rewriting primary expectations.

The direct handlers follow BaseMethods.Overset/Underset: normalize the direct base through the accepted ParseUtil.checkMovableLimits port; force an immediate operator mark's accent false; force the Underset result's accentunder false. Removing the two shadowing simpleMacros makes these existing handlers reachable. Stackrel and Overunderset dispatch are unchanged. This parser currently constructs inline movable-limit scripts as side-script nodes; normalizeStackBase restores only an embellished, movesupsub-marked msub/msup/msubsup argument to its corresponding under/over family, preserving explicit attributes/properties and reparenting the same children. The 50 complete primary tree checks cover the resulting handler representation. No generic script parser or renderer policy is changed.

The reference baseline is 5836be2e73642e7ee36e6e514ab724c2e687362a (D051/D055, D057, D058 and D052 included). The final candidate also inherits accepted D059 merge 75fd483b027e3c24eea5c4aa0e30b8beca7b8a1d. Of 50 public cases, 48 now match the raw primary SVG, full projected tree and dimensions; 40 of those differed from the parent. Two unchanged subsequent-script cases remain the tracked D053 Double exponent error (413×18 versus primary 20×25). D054 makes the two stackrel controls match their unchanged primary SVGs and projected trees, removing their prior transparent-wrapper qualifications. The two subsequent-script cases are isolated in stack_placement_boundaries.json with exact parent SVG and projected tree, original primary hash, source/mode and dimensions. They are not counted as raw primary matches and are not accepted through a generic normalization rule.

The related D058 24-case matrix now requires raw primary output for its four formerly shadowed Overset/Underset entries. Its historical baseline hashes remain documented for traceability.

Regenerate the 50 primary SVGs and trees with:

```sh
node testdata/generate_stack_placement.cjs /path/to/d2renderers/d2latex /tmp/stack-primary
```

The generator also rewrites its adjacent primary fixture. Run a copied generator in a scratch directory when independently verifying the checked-in fixture.
