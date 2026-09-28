# Radical bounding-box references

`radical_bbox_mathjax_3_2_2.json` contains 298 complete SVG strings from D2's
frozen MathJax 3.2.2 component, upstream commit
`ad8f5c21cb810236551da8c6512ba733e67357ee`. These are original output strings,
without normalization or values produced by the Go implementation.

The fixture uses `testdata/differential/oracle.mjs` with the original D2
`mathjax.js`, `polyfills.js`, and `setup.js`. Their SHA256 values are recorded in
the JSON. Each case creates a fresh VM and uses font cache none, `em=16`,
`ex=8`, and its recorded display setting. The oracle was run with Node's
`--jitless` option; no external runtime is needed to run the regression test.

The source behavior is `CommonMsqrt.computeBBox` in
[`ts/output/common/Wrappers/msqrt.ts`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/output/common/Wrappers/msqrt.ts)
and the constructor in
[`ts/util/BBox.ts`](https://github.com/mathjax/MathJax-src/blob/ad8f5c21cb810236551da8c6512ba733e67357ee/ts/util/BBox.ts).
The radicand's outer box is reconstructed from width, height, and depth.
The constructor resets scale, relative scale, spacing, and other metadata.
Copying all fields instead applies local `mathsize` or `scriptlevel` scaling
again when combining the radicand into the radical's box. Subsequent terms,
accents, fractions, and table cells then use incorrect bounds.

Coverage includes all ten TeX font sizes in square roots, indexed roots,
plain-TeX roots, groups, and fractions; independent and combined index sizes;
nested roots, scripts, surrounding size scopes, color, index offsets, explicit
MathML sizes and spacing, padding and borders; accents, underlines, overbraces,
raise/lower positioning, fences, arrays, and all four TeX math styles. Every
expression is checked in display and inline modes. Before the fix, 100 of the
298 complete SVG comparisons fail.

A visible public witness is
`\sqrt[\tiny n]{\tiny abcdef}+\sqrt{\Tiny uvwxyz}=0`: the incorrect short
boxes place following terms beneath the preceding radical's overbar.

Run `go test ./... -run TestRadicalBBoxPublicReferences`.
