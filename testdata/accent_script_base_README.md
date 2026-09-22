# Accent script-base references

These references use the unmodified MathJax 3.2.2 bundle pinned by D2 v0.8.1,
upstream commit `ad8f5c21cb810236551da8c6512ba733e67357ee`. Both generators
verify the three original bundle hashes before loading a fresh VM for each case.
The accepted comparison baseline is mathjax-go
`f1ae550cda1fb97e41bebf86799ffeae48cd3042`.

The public corpus has 52 cases in both inline and display modes. All 52 complete
SVGs match the untouched primary output. Ordinary decoration script permission
now fixes the two `nested-under` false DoubleSubscripts errors for
`\underline{\hat{x}}_i`. Against accepted `efd6d630`, those two SVGs change to
the primary output; the other 50 SVGs are byte-identical. The two `nested-line`
cases also gain the primary own `subsupOK: true` property. The other 48 complete
compiler-tree records are unchanged.

The explicit-tree projection retains kind, text, every explicit attribute, and
ordered children. The own-property projection additionally retains every own
property. Raw equality is 48/52 for explicit trees and 8/52 for own-property trees.
The boundary file records exact per-case paths and complete before/after field
maps for inherited metadata differences. Forty cases retain an extra `texClass: 0`
on the accent operator. The two `nested-line` cases additionally retain accepted
under/over `accent` metadata differences; their former missing-permission
qualification is removed. The two repaired `nested-under` cases bind four exact
inherited whole maps each to accepted `efd6d630` output for the identical
decoration prefix `\underline{\hat{x}}` before its later script. Each map records
the original primary map, expected complete map, exact node paths, source and
baseline digest. No field is globally dropped, and script permission itself is
never qualified.

The registered-MathML constructor corpus has 88 cases. It checks selected core
identity, first accent flags, and constructor-time `isMathAccent` before stretching,
then checks stability through measurement and repeated Go emission without input
tree or parent-identity mutation. It covers typed true/false, numeric and string
markers, missing/null values, scale, multi-character bases, operator sizing,
transparent wrappers, non-operator core returns, and selected action children.
The original 56 records are preserved exactly; 32 additions cover eight scalar
selection values in both complementary child-marker orders and both modes.
Fourteen additions fail with the former renderer lookup. The accepted parser's
exact scalar conversion is now shared through `mml.MactionSelectionNumber`;
the parser delegates to it, and the renderer uses it only for `maction` selection.
The separate generic numeric-attribute reader is unchanged.
All 34 action cases have valid primary constructor observations but subsequently
hit the original lite adaptor's missing `addEventListener` API. They are
constructor-only controls. No same-MathML paint-equality claim is made
for this private corpus.

The renderer change follows CommonScriptbase's constructor classification and
base traversal. The existing separate large-operator policy in `baseIsChar` is
unchanged. The coupled parser changes route `\bar` through the original Accent
handler and apply that handler's exact movable-limits rule. Neither renderer-only
nor bar-dispatch-only changes make all 20 original witnesses primary-exact; the
combined change does. The movable-limits rule is needed for the inline large
operator case once constructor classification is correct.

The full oracle exposed a cache prerequisite: constructor measurement can cache
the transparent parent of a subsequently stretched x-arrow. The primary
multi-part stretch branch invalidates that cache before assigning its final
stretched bounding box. Two dedicated raw-primary x-arrow cases and warmed/cold
ancestor, fixed-size, sibling, and identity controls bind the matching operation.
The earlier candidate failure is retained; no oracle expectation was changed.

Only the two inherited D049 `bar-control` wrapper qualifications are removed;
their expectations are now the original raw SVG hashes. Both ordinary-arrow
qualifications and all 48 original primary fixture records remain untouched.
The removed wrapper-normalization helper and its tests are obsolete because no
remaining case permits that transformation.

Run the public generator with the pinned `d2latex` directory and an evidence
directory; it writes `accent_script_base_mathjax_3_2_2.json` there. Run the
constructor generator with the same two arguments; it writes `constructor.json`
and available SVGs to the evidence directory. Neither generator rewrites the
primary bundle.
The x-arrow generator writes `accent_script_xarrow_mathjax_3_2_2.json` using the
same unchanged bundle and both display modes.
