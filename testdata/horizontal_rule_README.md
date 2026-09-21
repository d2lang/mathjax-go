# Horizontal rule references

`generate_horizontal_rules.cjs` uses fresh VMs of the unchanged SHA-pinned MathJax3.2.2 assets in D2 v0.8.1. It records40 complete-SVG hashes and serialized-ex dimensions covering underlines/overlines on low, high, long, fraction and script bases; nested rules, a vector combination, and ordinary/accent/arrow controls.

All40 actual complete SVGs and display measurements must match the primary exactly. D049 corrected the two over-arrow controls, so their former immutable-baseline qualification and data file are removed. The original strict40 failures and all complete primary/baseline/candidate SVGs remain in historical review evidence; no primary reference, coordinate or tolerance is adjusted.

The correction uses U+2015 for both overline and underline exactly as pinned BaseMappings requires. Existing renderer logic recognizes that character for horizontal-rule spacing. Ordinary bar, vector, wide accents and arrows are untouched. D047 supplies the separately reviewed cached base-height behavior required by low and nested overlines. Its former two underline qualifications are removed only after all40 inherited accent-height references match exactly on the composed source.
