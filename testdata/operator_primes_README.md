# Operator primes and ordinary-symbol cleanup

The pinned MathJax 3.2.2 source is commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. Its
[MmlMo inheritance](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/core/MmlTree/MmlNodes/mo.ts)
checks the complete operator text for two distinct character families:

* `checkPseudoScripts` recognizes the source's quote, prime, superscript and
  subscript ranges. It records `pseudoscript`, including explicit `false` in
  superscript-family parents, and gives eligible tokens zero inherited space.
* `checkPrimes` recognizes the eleven straight/curly quote characters and
  records their exact forward/reversed prime substitutions. It leaves the
  token's raw text intact. Existing SVG font selection can therefore preserve
  the authored text in an explicit CSS font while selecting prime glyphs in
  the normal math font.

These checks run after the operator dictionary and before accent metadata.
Go previously implemented only the left-backtick subset. This omitted metadata
on ordinary prime scripts and rendered grave accents with the wrong glyph.
A clear visible example is:

```tex
\Huge\boxed{\grave{x}\quad\grave{AB}\quad\grave{\mathcal{F}}}
```

The original and corrected result use U+2035 reverse-prime glyphs, with the
same original accent placement and box height. The baseline uses U+0060.

## Required cleanup boundary

The metadata change also requires the original
[FilterUtil.cleanStretchy](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/input/tex/FilterUtil.ts)
ORD wrapper. Without it, the new zero-space metadata changes the visible
spacing in `x°+y`. The fixture asserts that counterexample against the complete
original SVG in both modes.

The cleanup processes only registered `fixStretchy` construction events.
Those events are captured before the existing inheritance pass removes its
flag, deduplicated in their source order, and filtered for attachment to the
compiled tree. After movable-limit conversion, the positional operator table
and actual TeX class decide whether to add an ORD TeXAtom. This does not change
character dispatch or rewrite arbitrary explicit MathML operators.

The new wrapper inherits only the source's size, display style, script level
and prime style. Font and color remain on the token. At this particular
boundary, source
[AbstractNode.replaceChild](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/core/Tree/Node.ts)
clears the old operator's parent after the factory has inserted it into the
wrapper. Go reproduces that state locally before reinheritance. Its general
node-replacement ownership contract is unchanged. This matters for `x^{²}+y`:
retaining the parent incorrectly recomputes `pseudoscript=false` and selects a
different variant. Original parent observations are asserted separately from
SVG geometry.

A fraction's numerator participates in that logical-parent walk as well:
[MmlMfrac](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/core/MmlTree/MmlNodes/mfrac.ts)
inherits its first child's embellished/core identity from AbstractMmlBaseNode.
Go's cached dynamic flags omitted that case. Restoring it requires preserving
two separate output rules: CommonMfrac.canStretch always returns false, and
[CommonScriptbase.stretchChildren](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/output/common/Wrappers/scriptbase.ts)
asks the outer child whether it stretches before resizing its core operator.
Its temporary width measurement is also retained when every child stretches.
The MathML fraction can therefore be embellished without stretching its
numerator as though the fraction were a bare delimiter.

All 320 fresh prime/fraction contexts and 384 fresh fraction-as-annotation and
all-stretch controls match the original. Another 640 non-prime fraction
controls exercise function/operator classes, authored spaces, limits, accents,
scripts, tables and fences. The two preexisting fraction-accent reference SVGs
remain unchanged and exact. Every recorded raw/core/logical parent and prime
property assertion is strict; there is no fraction-parent qualification.

## Original references and remaining differences

`operator_primes_mathjax_3_2_2.json` contains **5,870 complete, unmodified
original SVGs**: 5,696 valid expressions and 174 original error renderings.
The three frozen D2 asset hashes are recorded in the fixture. Every conversion
uses a fresh VM, font cache none, `em=16`, `ex=8`, and the recorded display mode.
For 3,172 inputs it also records the original raw text, prime properties,
inherited pseudo-script spacing, raw parent, core parent and logical parent,
observed immediately before the original output pass.

Against merged main136, `61adeef0965dba380c43ab0fcdb5bbeb3d7bdd28`, **3,473
inputs were already exact and 2,397 become exact**, with zero formerly exact
regressions. The complete inventory has 6,126 inputs and covers every source
character family and neighboring excluded code point, mixed tokens, line
terminators, explicit attributes and fonts, logical parents, accents, scripts,
limits, radicals/fractions, nested stretchy fences, mapped character controls
(`: ) ] | ~`), factory-font provenance, and unbraced argument ownership. It
also includes independent review inputs and the prior skew residual inventory;
44 of those 52 skew residuals now match their untouched original SVGs.

`operator_primes_residuals.json` preserves the other **256 raw original,
baseline and candidate observations**. None was exact at baseline. Of these,
145 remain byte-unchanged and 111 change without reaching complete parity.
There are 254 valid original SVGs and two original unsupported-mathvariant
runtime errors. The file is an explicit unresolved inventory, not an assertion
that candidate outputs are correct. Forty-two quote cases include fresh original
literal-prime controls: the original literal equals the original quote result,
and candidate output equals the baseline rendering of the literal. Those
remaining differences therefore predate the remapping, including the newly
reviewed explicit-font grave numerator inside a fraction script, and include script
bounding boxes, CSS sizing/font selection and prescript placeholders.

The two historical BOM controls retain their original `mi` versus Go's `mo`
lexical mismatch. Cleanup adds only an ORD wrapper to the existing Go token;
the test binds the complete preexisting glyphs, positions and dimensions by
removing exactly that wrapper for its historical hash comparison. Original,
baseline and candidate SVGs remain in the unresolved inventory. Exact original
`\mmlToken{mo}{BOM}` and authored `\mathord` controls distinguish explicit MO
construction from registered fallback cleanup; explicit MO does not itself
register a cleanup event.

A separate affine SVG audit found no visual regressions among 4,786 previously
visually exact inputs and 1,142 newly visually exact inputs. It compares root
geometry and visible leaf geometry/paint after composing group transforms;
it does not call structural-only changes visual fixes. The remaining 35
changed visible residuals remain unresolved, including the recorded literal
controls, combining-mark layout and a mixed-prime Physics fence case. Four new raw
fraction/function observations (two formulas in both modes) have exact visible geometry and differ only by an
extra empty U+2061 ApplyFunction group; that separate dispatch issue remains
unpublished and is not bundled into this change.

Historical original JSON files remain untouched. Tests that formerly deleted
prime properties or accepted the missing ordinary wrapper now require the
complete original where the fix resolves it. Other historical qualifications
remain scoped to their existing fields and cases.

## Regeneration and verification

Run:

```sh
node --jitless testdata/generate_operator_primes.cjs PINNED_ASSETS
go test ./... -run 'TestOperatorPrime(References|OriginalMetadata)'
```

The generator checks the pinned asset hashes and processes batches of 24 fresh
VMs. Its observer wraps only the original output entry point and does not
mutate MathML. It refreshes original SVGs, metadata and literal controls while
leaving historical Go receipts unchanged. Unicode line separators are escaped
only on the JSONL wire so they remain part of their original TeX strings.
