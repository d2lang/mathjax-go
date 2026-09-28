// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: base/BaseMethods.ts Rule, color/ColorMethods.ts, and ArrayItem.copyEnv.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

func (p *parser) rule3D(name string) ([]*mml.Node, error) {
	space := node("mspace")
	for _, dimension := range []string{"width", "height", "depth"} {
		value, err := p.readDimension(name)
		if err != nil {
			return nil, err
		}
		space.Attributes.Set(dimension, value)
	}
	if name != "Space" {
		color := p.activeColor
		if color == "" {
			color = "black"
		}
		space.Attributes.Set("mathbackground", color)
	}
	return []*mml.Node{space}, nil
}
