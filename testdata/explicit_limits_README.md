# Explicit operator limits and script-family completion

These references use the unmodified MathJax 3.2.2 assets bundled with D2 v0.8.1, checked by SHA256 before every generator invocation. The source contract is MathJax-src `ad8f5c21cb810236551da8c6512ba733e67357ee`: BaseMethods.Limits, Superscript/Subscript and TeXAtom; FilterUtil.moveLimits; NodeUtil.copyChildren/copyAttributes/setProperties; and the node-specific coreMO methods.

The accepted baseline is `0daaf607d985dfdd636e1ff9c353355541425961`. This correction keeps the public Limits handler, default under/over-family creation, OP MathClass own movablelimits property, and post-inheritance moveLimits filter together. A standalone filter/property prerequisite changes six previously correct nolimits cases. An explicit-command-only family adapter fixes those cases but fails grouped scripted operators. The composed source policy retains correct default mathop rendering and supplies the required explicit Limits behavior.

## Public references

Run generators into a copied scratch directory to preserve the checked-in fixtures:

```
node generate_explicit_limits.cjs /path/to/d2latex /tmp/limits-primary
node generate_mathop_limits_filter.cjs /path/to/d2latex /tmp/mathop-primary
```

The 98-case explicit-Limits corpus covers sum/product/integral/mathop, repeated limits/nolimits, single and paired scripts before and after commands, grouped bases, authored annotations, braces, invalid ordinary operands, and pending/consumed prime lifetime. All 98 complete SVGs and compiled projected trees are checked. The projection retains every node kind, text, ordered child, explicit attribute and own property; it does not claim identical inherited/default attribute tables or every renderer flag. Of these cases, 94 have raw primary SVGs and 86 have raw primary projected trees. The accepted parent has only 33 raw primary SVGs.

The four unchanged raw SVG/tree boundaries are the two stackrel modes and the two consumed-superscript prime modes (the existing DoubleExponent behavior). Eight cases have raw primary SVGs but exact, path-specific inherited metadata boundaries: brace/underbrace accent and core class metadata, and consumed-subscript/closed-group prime variantForm, class and pseudoscript metadata. The fixture `explicit_limits_boundaries.json` binds their complete accepted-parent trees, exact allowed paths/values, source, mode, primary SVG hash and parent SVG hash. Every other field is compared against the primary. No broad subtree or attribute exclusions are used.

The separate 28-case MathClass corpus requires raw complete primary SVGs and projected trees for default, grouped, nested, styled, annotated and repeatedly switched mathop cases, including exact primary error outputs. It guards the prior exploratory regression where nested inline mathop changed from primary 20×43 to 11×58. The coupled inherited 40 nested-movable, 16 operator-level and 18 script-class cases now all require raw primary SVGs: their 14 former single-tag qualifications are removed. The D058 direct-base private assertion now also requires primary own movablelimits=false for normalized mathop; its previous missing-property qualification is removed.

## Private method and ownership contracts

The private fixtures execute the actual registered pinned methods. Twelve Limits conversions bind node-family semantics, child identities/reparenting and intentional wrapper attribute/property discard. Eighty maction-selection controls bind scalar JavaScript numeric coercion, clamping and fresh-empty-row behavior. The supported scalar contract includes null, booleans, numeric values and strings; it does not emulate arbitrary JavaScript object coercion. Authored annotation subtypes retain primary raw child order. Only eager parser scripts carry a private transient origin, copied through cloning and consumed as needed by the adapter; Compile removes it before inheritance/rendering. Pending-prime rejection follows the actual stack contract without changing prime grouping/counting.

Twelve registered filter controls exercise nested same-kind and mixed-kind conversions in both creation orders. They bind exact shared Attributes identity, copied property ownership (including unknown-property handling through source setProperties), current-child identities/reparenting, detached old-wrapper state, list removal and the complete live primary topology. Twenty additional controls bind display inheritance/explicit precedence, own-property versus attribute lookup, core explicit versus inherited movablelimits, wrapping, truthy scalar values and detached registrations.

The ordered private helper follows registered per-kind list order and fixed family pass order. The compiler collects live nodes rather than retaining a parser-wide creation registry. The tests prove live topology and dynamic flags agree for the tested same-kind orderings, and mixed-kind pass order remains fixed. Detached old-wrapper traces can differ with collection order; identical historical creation traces are deliberately not claimed. Replacements preserve the predicate-relevant core delegation and shared Attributes/copy-all-properties policy, so ancestor and descendant replacements retain the same live result.

Prime attachment behavior beyond the precise Limits eligibility lifetime, stackrel construction, duplicate-script policy and generic core helpers remain outside this correction.
