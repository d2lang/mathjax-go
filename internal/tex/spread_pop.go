// Copyright (c) 2021-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: MathtoolsMethods.SpreadLines and MathtoolsUtil.spreadLines.
package tex

import (
	"fmt"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// Pop.toMml is selected before the result is pushed into any outer prefix.
// Only a direct mtable or immediate inferred-row children receive spacing.
// A nil spread denotes the absent property on a popped non-Begin stack item.
func mathtoolsSpreadPop(content *mml.Node, spread *string) error {
	for _, current := range unwrapInferred(content) {
		if current.Kind != "mtable" {
			continue
		}
		spacing, _ := current.Attributes.Get("rowspacing")
		if !limitsTruthy(spacing) {
			// The source assigns undefined on this branch without calling
			// dimen2em. An explicit nil retains the absent effective value.
			var value any
			if spread != nil {
				value = *spread
			}
			current.Attributes.Set("rowspacing", value)
			continue
		}
		if spread == nil {
			// dimen2em(undefined) throws at this first Pop, before any
			// subsequent token or closing-item reduction can run.
			return fmt.Errorf("spreadlines popped a table without a spread dimension")
		}
		mathtoolsAddRowSpacing(current, *spread)
	}
	return nil
}

func mathtoolsSpreadPending(nodes []*mml.Node) error {
	return mathtoolsSpreadPop(row(nodes, true), nil)
}
