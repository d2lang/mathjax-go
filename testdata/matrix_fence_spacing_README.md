# Legacy Matrix fence spacing

MathJax 3.2.2's `Base.Matrix` stores the `pmatrix` and `cases` fence pair on its ArrayItem. `ArrayItem.toMml` calls `ParseUtil.fenced`, which gives the enclosing mrow `open`, `close`, and `texClass: INNER` properties. Those properties preserve thin TeX spacing from preceding text, ordinary atoms, and operators. The Go legacy commands used a plain delimiter row and lost that spacing; the existing `leftRightFenced` constructor restores the original row contract. Matrix environments already use that contract.

`matrix_fence_spacing_mathjax_3_2_2.json.gz` preserves 1,236 fresh original requests: 1,156 valid SVGs, 64 exact original diagnostics, and 16 separately retained original-unsupported `boldsymbol` diagnostics. The last group receives no exact-SVG assertion or parity credit because Go accepts that extra command. No original runtime failures occurred. All 1,220 strict SVG assertions match the fix; against canonical baseline `fc40841854f0b3cdad53ee346a71646c478b7ba8`, 98 valid outputs are fixed and 1,122 assertions remain unchanged. There are no SVG regressions.

The corpus covers all eight legacy Matrix commands (`matrix`, `array`, `pmatrix`, `cases`, `eqalign`, `displaylines`, `eqalignno`, `leqalignno`), preceding and following atom classes, surrounding text and operators, font and script scopes, nested fractions and roots, matrix cell bodies, ten matrix/array environment controls, and original diagnostic boundaries. Inline and display mode are tested for every input. `matrix_fence_spacing_observations.json` independently records the frozen active registrations, original methods, and original fenced-row properties after rendering.

Regenerate only with the hash-verified, unmodified D2 MathJax 3.2.2 assets; each request gets a fresh runtime:

```sh
python3 testdata/generate_matrix_fence_spacing.py /path/to/pinned/assets --node /path/to/node
node --jitless testdata/observe_matrix_fence_spacing.cjs /path/to/pinned/assets
go test ./... -run TestLegacyMatrixFenceSpacingOriginalReferences -count=1
```

The finite corpus establishes the measured fence spacing contract; it does not establish universal MathJax parity.
