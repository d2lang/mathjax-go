# Unicode mathematical alphabet fallback references

The 54 complete-SVG references use the unmodified MathJax3.2.2 component from
D2v0.8.1, with exact asset hashes and a fresh VM per formula/display mode.
The initial examples distinguish double-struck k and bold raw alpha from plain
fallback characters even when their measured bounds are identical.

Fourteen explicit variants cover effective-map inheritance, known outlines,
mathematical-alphabet fallback, ASCII/Greek/digit/punctuation combinations and
fallback CSS. Additional controls cover known R/Gamma paths, ordinary math,
CJK and unmapped supplementary characters, literal mathematical-alphabet text,
and the upper boundary of that Unicode block.

CommonWrapper.unicodeChars remaps effective variant characters before lookup.
SVG.unknownText omits CSS font attributes for a single mathematical-alphabet
character, whose codepoint already specifies its style. Both decisions are
required to preserve the reference output; replacing only the text leaves
extra font-family/weight/style attributes. Font tables, parser policy and known
path lookup remain unchanged.

Regenerate with:

```sh
node testdata/generate_unicode_fallback.cjs /path/to/d2latex /tmp/unicode-fallback-svg
```
