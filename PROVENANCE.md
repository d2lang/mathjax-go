# Source provenance

mathjax-go is a source-shaped Go translation of the TeX-to-SVG subset in
MathJax 3.2.2 that D2 executes.

| Source | Version / commit | License | Relevant surface |
| --- | --- | --- | --- |
| MathJax source | 3.2.2 / `ad8f5c21cb810236551da8c6512ba733e67357ee` | Apache-2.0 | `ts/core`, `ts/input/tex`, `ts/output/common`, `ts/output/svg`, and TeX SVG fonts |
| MathJax components | 3.2.2 / `600692ad9d3552cc25f85510d5797bc942ecc9f7` | Apache-2.0 | component build used to identify the frozen module surface |
| mhchemParser | 4.1.1 npm package, SHA-1 `a2142fdab37a02ec8d1b48a445059287790becd5` | Apache-2.0 | `src/mhchemParser.ts` for `\ce` and `\pu` |
| Luo-Studio/go-tex | commit `60635a3613c51e710fd5cb57e1d7eeb7b3163a1f` | MIT | initial Go translation structure for `internal/mhchem`; behavior was then corrected against mhchemParser 4.1.1 |
| D2 custom-component assets | D2 commit `541941e895fcdc69ad3109770296dcac670c5be7` | `mathjax.js`: Apache-2.0; `polyfills.js`: MIT; D2-authored `setup.js`: MPL-2.0 | asset-origin commit used to extract the semantic manifests and exact package/configuration surface |
| D2 formula/integration snapshot | D2 commit `5666d9337c77a4803b9cd60cb1c5a24439d0f949` | MPL-2.0 | D2 formula corpus and CI fetch commit; the three oracle assets are byte-identical to the asset-origin commit above and retain their per-file licenses listed in the preceding row |

## Frozen D2 oracle

The optional differential runner accepts only these files:

| File | SHA-256 |
| --- | --- |
| `mathjax.js` | `cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869` |
| `polyfills.js` | `7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01` |
| `setup.js` | `a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881` |

The semantic manifests were extracted from the three assets at D2 commit
`541941e895fcdc69ad3109770296dcac670c5be7`. The formula corpus and CI fetch
the same byte-for-byte assets at D2 commit
`5666d9337c77a4803b9cd60cb1c5a24439d0f949`; the hashes above are the
authoritative identity check for either commit.

The bundle reports MathJax version 3.2.2. Its configured TeX packages are
`base`, `mathtools`, `ams`, `amscd`, `braket`, `cancel`, `cases`, `color`,
`gensymb`, `mhchem`, and `physics`. Its SVG output uses `fontCache: 'none'`.
The call boundary is `html.convert(tex, {em: 16, ex: 8})`, whose default is
display mode, followed by serialization of the container's inner HTML.

`polyfills.js` is used only by the external oracle. It contains code from
XMLDOM and Wicked Good XPath under their respective MIT licenses; none of that
code is translated into, embedded in, or distributed as part of the production
Go library.

## Go source mapping

| Go path | Upstream source or role |
| --- | --- |
| `internal/mml/attributes.go` | `ts/core/MmlTree/Attributes.ts`, translated and modified |
| `internal/mml/node.go` | `ts/core/Tree/Node.ts` and `ts/core/MmlTree/MmlNode.ts`, translated and modified into a compositional Go tree |
| `internal/mml/factory.go` | `ts/core/MmlTree/MmlFactory.ts`, translated and modified |
| `internal/tex/compiler.go`, `parser.go`, `nodes.go`, `inheritance.go`, `mml_definitions.go` | `ts/input/tex/TexParser.ts`, `Stack.ts`, `StackItem.ts`, `NodeFactory.ts`, `Configuration.ts`, and `ts/core/MmlTree/MmlNodes/*.ts` |
| `internal/tex/commands.go`, `environments.go`, `newcommand.go`, `symbols.go`, `operator_dictionary_generated.go` | `ts/input/tex/base/*.ts`, `ts/input/tex/newcommand/*.ts`, `ts/core/MmlTree/OperatorDictionary.ts`, and command/environment methods from the selected AMS, AMS-CD, braket, cancel, cases, color, empheq, enclose, gensymb, mathtools, mhchem, and physics packages; each aggregate file header lists its exact source paths |
| `internal/tex/*_handlers.go`, `internal/tex/extensions/**`, `internal/tex/source_tables_*.go` | selected MathJax 3.2.2 TeX package configurations, mappings, and methods for AMS, AMS-CD, braket, cancel, cases, color, empheq, enclose, gensymb, mathtools, mhchem, newcommand, and physics |
| `internal/mhchem/**` | `mhchemparser@4.1.1/src/mhchemParser.ts`, translated and modified; the initial Go structure identified above is retained under the accompanying MIT notice |
| `internal/layout/**` | `ts/output/common/BBox.ts`, `Wrapper.ts`, `OutputJax.ts`, `FontData.ts`, and `ts/util/lengths.ts` |
| `internal/svg/**` | `ts/output/common/Wrappers/*.ts`, `ts/output/svg/Wrappers/*.ts`, `ts/output/svg/Notation.ts`, and the SVG DOM/serializer behavior of `ts/output/svg.ts` |
| `internal/font/generate.py`, `internal/font/data_*_gen.go`, `font.go`, `delimiter.go`, `smp.go` | mechanically generated and translated from `ts/output/common/fonts/tex/**` and `ts/output/svg/fonts/tex/**`; exact MathJax numeric literals, delimiter records, and SVG glyph paths are preserved |
| `internal/jscompat/number.go` | ECMAScript Number behavior used by `ts/output/common/OutputJax.ts` and `ts/util/lengths.ts` |
| `internal/ordered/map.go` | deterministic replacement for observable JavaScript object/DOM insertion ordering |
| `internal/pipeline`, root API | Go-specific adapter seam replacing MathJax's document/handler startup layer |
| `internal/oracle`, `differential_test.go`, `internal/tex/testdata/d2_rich_mml_mathjax_3_2_2.json` | Go-specific, test-only compatibility harnesses and the rich semantic manifest extracted from the pinned D2 component assets |
| `internal/tex/testdata/handler_oracle_mathjax_3_2_2.json` | test-only MathML manifest extracted from pinned MathJax 3.2.2 with the explicitly recorded augmented package set |
| `internal/tex/testdata/operator_dictionary_mathjax_3_2_2.json` | test-only source manifest extracted from `ts/core/MmlTree/OperatorDictionary.ts` at the pinned MathJax source commit |
| `internal/tex/testdata/source_tables_mathjax_3_2_2.json` | test-only source manifest extracted from pinned MathJax package mappings/configurations and constrained by D2's configured package closure |
| `testdata/differential/oracle.mjs` | Go-specific test driver for the unmodified, externally supplied D2 assets |

Translated files retain upstream copyright attribution, identify the Apache
License 2.0 with an SPDX header, and state that they are Go translations and
modifications. Generated font files retain MathJax's applicable copyright and
Apache 2.0 license notice. No JavaScript oracle asset is embedded in or needed
by the production library.
