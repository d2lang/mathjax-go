# Overunderset direct-base movable limits

The primary reference is unmodified MathJax 3.2.2 from D2 v0.8.1. The generator verifies the three original asset hashes and runs a fresh VM per input/mode. Source contract: MathJax-src `ad8f5c21cb810236551da8c6512ba733e67357ee`, ParseUtil.ts400–406, NodeUtil.ts118–134/289–303, MmlMo.ts397–414, and BaseMethods.ts Overunderset.

Regenerate all 24 complete primary SVGs and trees without rewriting the checked-in fixture:

```sh
node testdata/generate_overunderset_direct_base.cjs /path/to/d2renderers/d2latex /tmp/direct-base-primary
```

Compare `primary.json` name, source, mode, dimensions, tree and SVG SHA256 to the corresponding fixture fields. The public test checks the complete SVG hash for every case, and the exact compiled explicit-attribute/content tree plus display Measure for all 24 cases (20 Overunderset plus four Overset/Underset). The extra primary property/flag observations remain available in the fixture, but the test does not claim whole-property AST equivalence.

Five inline cases fail on the original accepted base `0e8bf718f9b7d18c8d9a506313dc2d8f190cac75`: direct sum, product, explicit movable sum, explicit movable integral and explicit movable plus. They all match the raw primary after this correction. The final candidate inherits the separate immediate-base class-transfer fix at `b8a5a6f2a3da69fa0f506c7cf943c1e56c43b38f`.

The later D050 direct-dispatch correction upgrades the four Overset/Underset entries to raw primary SVG/tree/dimension checks. The following historical hashes document their former dispatcher-boundary outputs; they are no longer accepted expectations:

- `overset-product-inline`: `567c1b51c9bf9af571a0426df7e0bcd6e353ca7ab3c2b776580457a26b56d9c3`
- `overset-product-display`: `f1a23f589be3d68bfc9af425421d8267598218328873eebf94cce3252f6633cf`
- `underset-mathop-inline`: `45db42a276cacc18f1584a952ffe6cd7b16589b9ae48c47336e2bab96a07e6aa`
- `underset-mathop-display`: `b141d49b21ef66a9d3f08940d272c8ec5d7c74ff83cf12a933321de5d9d5674a`

The current `mathClass` constructor also omits the primary TeXAtom's own `movablelimits` property for authored `\mathop{x}`. Its whole SVG and explicit-attribute/content AST already match this matrix; this preexisting metadata omission remains qualified and is not repaired here. The private helper tests independently require exact normalization when a direct TeXAtom/mstyle/mo/mrow own property is actually present. They cover JavaScript truthiness, attribute-versus-property semantics, first-match contextual dictionary form policy, identity preservation and refusal to descend grouped/scripted bases.

The original D058 correction changed only the reachable Overunderset caller. D050 now reuses the same helper for the direct Overset/Underset handlers; their narrowly scoped inline script-family compatibility is documented in stack_placement_README.md. Brace construction, stackrel and generic core traversal remain outside the dispatcher change.
