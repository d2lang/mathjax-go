// Copyright (c) 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

// BaseMappings registers both argument-taking MathFont commands and SetFont
// declarations. Use the retained source registrations so aliases share the
// existing handlers, including explicit empty variants such as mathnormal.
func baseFontMappings(method string) map[string]string {
	bindings := make(map[string]string)
	for _, sourceMap := range mjSourceBaseMaps {
		if sourceMap.Kind != mjSourceCommandMap || sourceMap.Name != "macros" {
			continue
		}
		for _, entry := range sourceMap.Entries {
			handler, args, ok := sourceHandler(entry.Value)
			if !ok || handler != method || len(args) != 1 {
				continue
			}
			if variant, ok := args[0].(string); ok {
				bindings[entry.Name] = variant
			}
		}
	}
	return bindings
}
