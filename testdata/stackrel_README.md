# Primary stackrel references

`generate_stackrel.cjs` runs the unmodified, SHA-pinned D2 v0.8.1 MathJax 3.2.2 assets in a fresh VM for each expression/mode. Its 72-case matrix retains 10 intentional primary error results, complete SVG hashes, display measurements, and every node's kind, text, explicit attributes, own properties and ordered children. Optional evidence also saves the raw SVG and full observed parser metadata before unchanged rendering. No transform or error is normalized.

The sole production change is the exact BaseMappings macro `\mathrel{\mathop{#2}\limits^{#1}}`. The existing macro table takes precedence over the legacy command switch, so no parser or alias deletion is needed. Accepted D056 supplies the source Limits/family behavior this macro requires. Nested `\stackrel{a}{\sum_i^n}` previously measured 37×21 inline instead of primary 20×54; the full primary SVG, not only its size, is required.

Cases cover nested operators and stackrel, sum/product/integral/lim, explicit limits/nolimits, one/both scripts, grouped operands, scripts around and inside stackrel, fonts, fractions, empty arguments, relation spacing, explicit expansion/overset controls, and exact missing-argument/duplicate-script/misplaced-Limits errors. The independent D053 annotation attachment and D060 pending-prime behavior are not changed.

Two inherited fixtures remove only obsolete stackrel inline/display qualifications. Their original primary fixtures remain unchanged. All other D050/D056 error and metadata qualifications remain explicit. Run the generator with pinned Node and the pinned d2latex directory; `NODE_OPTIONS=--jitless` avoids the separately retained intermittent Node optimization crash without changing primary assets or assertions.
