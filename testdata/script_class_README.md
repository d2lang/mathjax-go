# Immediate script-base TeX-class transfer (D057)

The reference is the unmodified MathJax 3.2.2 bundle embedded in D2 v0.8.1,
verified by three SHA-256 values in each generator. Run the public generator
with that asset directory and an optional directory for complete SVG/AST
artifacts. The registered generator observes twelve `setTeXclass` calls on
fresh factory-created trees, before and after, including returned node identity.
Neither generator changes the primary parser, inheritance or renderer.

`AbstractMmlBaseNode.core()` returns its immediate first child. It differs
from `coreMO()`, which descends recursively. `updateTeXclass` transfers and
clears previous-class/level state. An inner script already consumed the
operator's state, so another transfer from that operator loses the outer
script's spacing. D057 changes only that transfer's source to the immediate
base; all recursive layout/operator helpers remain unchanged.

Eighteen public cases cover sum/product display rows, integral nested scripts
and under/over stacks in both modes, transparent groups and ordinary/script
controls. All18 complete public SVGs equal primary bytes after D056. The former simple-sum-inline literal tag qualification is preserved historically and removed from current acceptance. All18 same-primary-MathML renderings, post-output class/previous/level state and repeat checks remain unchanged and exact.


Twelve direct primary-derived controls cover nested embellished scripts,
transparent rows/styles, direct operators/identifiers and nonembellished
bases at matched current/previous levels 0 and 2. Every node's class and
previous-class/level, the returned node, source attributes and object/parent
identity are checked. They detect consumption of the immediate child's state
and preserve the nonembellished fresh-chain path.

An initial independent mixed-level control (current inherited level 0,
previous explicit level 2) exposed an existing separate operator rule:
primary `MmlMo.setTeXclass` uses current inherited level; Go reads the previous
resolved level. Those raw observations are retained outside the repository
as an unresolved registered-MML boundary, not accepted by these tests or
changed by D057. The direct controls here intentionally use matched levels
so they isolate transfer ownership. The initial public oracle capture was
also retained: pre-output state precedes primary TeX-class preparation, so
final expected state is captured after the unmodified output call.

Nested movable-limit rendering (D052), direct Overunderset normalization
(D058), dispatch/script attachment and explicit-limits policies are separate.
No assertion or fixture from another accepted fix is modified.
