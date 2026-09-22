# Current inherited operator script level (D059)

Pinned MathJax 3.2.2 `MmlMo.adjustTeXclass` reads the current operator's
`attributes.getInherited('scriptlevel')` when a previous node exists.
`TeXAtom` uses exactly that same method. The inherited lookup includes its
per-node default/global prototypes but excludes explicit attributes. With
no previous node, primary updates the previous class while retaining the
stored previous level. D059 implements these two branches without changing
the generic `getPrevClass` rule, class demotion, inheritance, or parsing.

Six public outputs have missing spacing after a script-style change:
`\scriptstyle x\textstyle +y`, its scriptscriptstyle equivalent, and the
relation form, each in both modes. Sixteen fresh public references also cover
grouped styles, current-small, sum scripts and ordinary operators. All16 complete public SVGs are raw-primary exact after D056; the old inline tag qualification is preserved historically and removed from current acceptance.
All sixteen same-primary-MathML SVGs and every post-output class/previous/level
state must match raw primary, including repeated rendering.

Eighteen direct registered controls observe both `mo` and `TeXAtom`:
inherited zero/two, explicit ignored, default/global fallback, fresh previous
null, an actual previous-node call followed by null, and explicit NONE class.
They preserve source attributes, node/parent identity and returned identity.
The recorded before/after snapshots and layer observations distinguish
inherited from resolved values and prove the previous-null retention branch.
JSON fixtures use the primary's numeric integral script levels; implementation
supports the int/int64/float64 representations used by Go state and JSON.

Both generators require all three SHA-verified unmodified D2 v0.8.1 MathJax
assets. The public generator captures before and after the original output
call without changing it. Direct controls invoke the original primary method;
no oracle source is patched. Tests reconstruct its exact registered trees.

D057 immediate-base transfer remains inherited and all existing tests remain
unchanged. D052 nested movable limits and D058 direct-base normalization are
separate. No dispatch, explicit-limits or font policy change is included.
