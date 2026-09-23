// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.
// Source: ts/input/tex/TexParser.ts and the active delimiter SymbolMaps.

package tex

// Delimiter readers use exact source keys: / differs from \/, and | differs
// from \|. Ordinary command lookup has its own dispatch policy.
func lookupDelimiter(key string) (string, bool) {
	for i := len(mjSourceMaps) - 1; i >= 0; i-- {
		sourceMap := mjSourceMaps[i]
		if sourceMap.Kind != mjSourceDelimiterMap {
			continue
		}
		switch sourceMap.Name {
		case "delimiter", "AMSmath-delimiter", "AMSsymbols-delimiter", "mathtools-delimiters":
		default:
			continue
		}
		for j := len(sourceMap.Entries) - 1; j >= 0; j-- {
			entry := sourceMap.Entries[j]
			if entry.Name == key {
				character, _, ok := sourceSymbolValue(entry.Value)
				return character, ok
			}
		}
	}
	return "", false
}
