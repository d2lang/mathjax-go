// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2 behavior.
// Source: ts/input/tex/ParseUtil.ts, readKeyval and keyvalOptions.

package tex

import "github.com/d2lang/mathjax-go/internal/ordered"

// parseUtilKeyvalOptions collects the complete list before validating its own keys.
// The original readKeyval uses an ordinary JavaScript object: primitive
// assignments to __proto__ do not create an own property, and Object.keys
// enumerates array-index properties before other keys.
func parseUtilKeyvalOptions(text string, allowed func(string) bool, errorOnUnknown bool) (*ordered.Map[any], error) {
	result := ordered.New[any]()
	rest := text
	for rest != "" {
		key, end, next, err := empheqReadOptionValue(rest, "=,")
		if err != nil {
			return nil, err
		}
		rest = next
		var value any = true
		if end == '=' {
			var raw string
			raw, _, rest, err = empheqReadOptionValue(rest, ",")
			if err != nil {
				return nil, err
			}
			switch raw {
			case "true":
				value = true
			case "false":
				value = false
			default:
				value = raw
			}
		} else if key == "" {
			continue
		}
		if key != "__proto__" {
			result.Set(key, value)
		}
	}
	if allowed != nil {
		for _, key := range result.JavaScriptKeys() {
			if !allowed(key) {
				if errorOnUnknown {
					return nil, texError("InvalidOption", "Invalid option: %s", key)
				}
				result.Delete(key)
			}
		}
	}
	return result, nil
}
