// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: BaseItems.NonscriptItem and BaseConfiguration.filterNonscript.
package tex

import (
	"fmt"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// NonscriptItem remembers only the immediately following MML item when it
// contains a space. Other stack items consume it; commands which push no item
// leave it pending. The inherited scriptlevel is not available until the
// post-filter, so no parser-time style approximation belongs here.
func (p *parser) nonscriptSpace(space *mml.Node) *mml.Node {
	if space == nil {
		return space
	}
	inner := space
	if inner.Kind == "mstyle" && inner.Flags.NotParent {
		if len(inner.Children) == 0 || len(inner.Children[0].Children) == 0 {
			return space
		}
		inner = inner.Children[0].Children[0]
	}
	if inner.Kind != "mspace" {
		return space
	}
	if inner != space {
		// Fixed-size spacing macros explicitly set their mstyle's level to
		// zero. This temporary outer row records the surrounding level.
		space = node("mrow", space)
	}
	p.state.nonscriptSpaces = append(p.state.nonscriptSpaces, space)
	return space
}

// The source filter runs after inheritance, before moveLimits/cleanStretchy.
// Retained styled spaces shed only the temporary row; script spaces are
// removed with their wrapper. Genuine authored groups never enter this list.
func filterNonscript(spaces []*mml.Node) error {
	for _, space := range spaces {
		parent := space.Parent
		if parent == nil {
			continue
		}
		level, exists := space.Attributes.Get("scriptlevel")
		// The source's > 0 comparison uses the same ECMAScript scalar
		// number conversion as the existing maction selection helper.
		if mml.MactionSelectionNumber(level, exists) > 0 {
			// The original filter splices required children too, which can
			// make its output wrapper throw. Report that conversion failure
			// instead of handing an incomplete fixed-arity node to layout.
			if arity := parent.Flags.Arity; arity > 0 && arity != mmlUnboundedArity && len(parent.Children)-1 < arity {
				return fmt.Errorf("nonscript removal leaves %s without a required child", parent.Kind)
			}
			_ = parent.RemoveChild(space)
		} else if space.Kind == "mrow" && len(space.Children) != 0 {
			_ = parent.ReplaceChild(space.Children[0], space)
		}
	}
	return nil
}
