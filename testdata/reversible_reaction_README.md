# Reversible chemistry reactions and three-dimensional rule spacers

The pinned Mhchem `longleftrightarrows` macro constructs its arrow with
`\stackrel{\longrightarrow}{\smash{\longleftarrow}\Rule{0px}{.25em}{0px}}`.
Both the macro and its `Rule` dependency were missing from executable command
dispatch, so ordinary `\ce{N2O4 <--> 2NO2}` produced an undefined-command error.

`BaseMethods.Rule` reads three dimensions with `GetDimen`, then creates an
`mspace` with width, height and depth. `Rule` sets its background to the current
lexical color or black; its `Space` alias leaves the background unspecified.
Color declarations persist in their local parser environment, while groups,
arrays, internal math and genuine child parsers preserve the original copying
and reset boundaries. This change adds only the color state consumed by the
new rule command; it does not modify the existing lowercase `rule` handler.

The fixture contains 120 complete original SVGs in inline and display modes,
with 60 display dimension assertions. It covers reversible reactions with top
labels, macro overrides, direct rules and blank spaces, signed and physical
units, missing/invalid dimensions, script/size contexts, nested color scopes,
array resets, boxed/internal-math resets, and ordinary reaction/spacing
controls. These are raw primary outputs, without candidate substitutions.
Lower reaction labels remain outside this corpus because their existing
raise/lower placement differs independently.

Regenerate using the frozen assets recorded in `PROVENANCE.md`:

```sh
NODE_OPTIONS=--jitless node testdata/generate_reversible_reaction.cjs PINNED_ASSETS
```

The generator verifies all three asset hashes and creates a fresh MathJax
runtime per expression. The source contracts are MathJax 3.2.2
`MhchemConfiguration.ts`, `BaseMethods.Rule`, `ColorMethods.Color/TextColor`,
and `ArrayItem.copyEnv`.
