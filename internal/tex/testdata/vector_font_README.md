# Physics vector font references

`vector_font_mathjax_3_2_2.json` records 48 expressions in inline and display mode from unmodified, hash-pinned D2 v0.8.1 MathJax 3.2.2. `generate_vector_fonts.cjs` verifies the three source asset hashes, starts a fresh VM for each expression/mode, captures the complete pre-output explicit MathML tree, and hashes the unchanged complete SVG. Optional argument 3 saves each original SVG. No parser, MML or renderer overlay generates these references.

Coverage includes ordinary/star vectors; outer/inner MathFont and declaration scopes; Latin, upper/lower Greek, digits, accents and operators; grouped identifiers; authored MML and kept variants; nested vectors and their environment deletion; arrow/unit aliases and long names; script arguments; text/internalText and operatorname creation boundaries; and ordinary/post-vector controls. Every case compares every ordered token's kind and font variant, and asserts that private parser properties do not escape compilation.

All 96 cases compare the complete primary trees directly, including ordinary accent attributes and MtLap's explicit mstyle displaystyle=false/scriptlevel=0. No tree qualification remains, and the raw primary fixture is unchanged.

All 96 actual complete SVGs must equal the primary hash. The earlier ordinary vec stretch correction removed the six rendering exceptions and stretchy qualification; subsequent accent and MtLap repairs removed the remaining metadata qualifications.

The implementation follows PhysicsMethods.createVectorToken/VectorBold and the registered va/vu StarMacro expansions at MathJax commit ad8f5c21cb810236551da8c6512ba733e67357ee. Authored MmlToken and ParseUtil.internalText use the node factory and bypass the vector token factory; AMS HandleOperatorName installs an explicit normal-font environment, suppressing it. The generic boldsymbol extension and ordinary vec renderer/handler are not reimplemented here.
