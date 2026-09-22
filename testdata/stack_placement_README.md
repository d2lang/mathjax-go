# Direct Overset/Underset dispatch

These 50 references come from fresh VMs of unchanged MathJax 3.2.2 assets pinned by the generator to D2 v0.8.1 and MathJax-src ad8f5c21cb810236551da8c6512ba733e67357ee. The complete original SVG hashes, dimensions and explicit-attribute/content trees are retained without rewriting primary expectations.

The direct handlers follow BaseMethods.Overset/Underset: normalize the direct base through the accepted ParseUtil.checkMovableLimits port; force an immediate operator mark's accent false; force the Underset result's accentunder false. Removing the two shadowing simpleMacros makes these existing handlers reachable. Stackrel and Overunderset dispatch are unchanged. This parser currently constructs inline movable-limit scripts as side-script nodes; normalizeStackBase restores only an embellished, movesupsub-marked msub/msup/msubsup argument to its corresponding under/over family, preserving explicit attributes/properties and reparenting the same children. The 50 complete primary tree checks cover the resulting handler representation. No generic script parser or renderer policy is changed.

The reference baseline is 5836be2e73642e7ee36e6e514ab724c2e687362a (D051/D055, D057, D058 and D052 included). The original D050 change made 46 of 50 public cases raw primary. Accepted D054 merge 33983b0fb1d9f4e2a4d940e8f4d45a709d02012a makes the two stackrel cases raw primary. D053 then makes the two subsequent-script cases raw primary (20×25 instead of the former false 413×18 error). All 50 cases now require the complete raw primary SVG, projected explicit tree and dimensions; no qualified cases or normalization remain in this test. Both historical qualification sets and their exact accepted-parent receipts are preserved in the D053 review checkpoints.

The related D058 24-case matrix now requires raw primary output for its four formerly shadowed Overset/Underset entries. Its historical baseline hashes remain documented for traceability.

Regenerate the 50 primary SVGs and trees with:

```sh
node testdata/generate_stack_placement.cjs /path/to/d2renderers/d2latex /tmp/stack-primary
```

The generator also rewrites its adjacent primary fixture. Run a copied generator in a scratch directory when independently verifying the checked-in fixture.
