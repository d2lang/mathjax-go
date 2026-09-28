# Lowercase rule command

The fix follows `BaseMappings.rule` and `BaseMethods.rule` in MathJax 3.2.2. Width and height are read with `TexParser.GetDimen`, so comma decimals, unbraced dimensions, all registered TeX units, and invalid-dimension errors behave like the original. The filled mspace uses the lexical color or black. A nonempty optional vertical offset wraps that mspace in mpadded, adjusting height and (for an initial minus sign) depth exactly as the source does; the optional offset itself is deliberately not reinterpreted as GetDimen input.

The fixture records 308 complete, unmodified SVGs from D2's frozen MathJax runtime at `ad8f5c21cb810236551da8c6512ba733e67357ee`. The generator validates the original three asset hashes and runs Node `--jitless`. Regenerate with `node --jitless testdata/generate_lowercase_rule.cjs /path/to/pinned/assets`.

Coverage includes dot/comma dimensions in all nine units, whitespace/sign/zero dimensions, braced and unbraced arguments, positive/negative/relative/unusual offset strings, default and explicit colors, lexical scope through groups/matrices/internal math/boxed math, local sizes/scripts, fractions/roots, macros, functions/primes, and missing/invalid arguments. Capital `\Rule`, blank `\Space`, raise, and horizontal-space controls remain exact. All 308 results match byte for byte; 278 differed on baseline `e6465ae`.

D2 witness math: `\rule[-0.4em]{2,5em}{0.2em}\quad x\quad\rule[0.6em]{2em}{0.2em}`. It demonstrates both comma dimensions and visible vertical positioning.
