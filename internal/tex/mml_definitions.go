// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/core/MmlTree/{MML,MmlNode,MmlFactory}.ts.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

const mmlUnboundedArity = mml.UnboundedArity

// The parser and synthesized SVG nodes use the same pinned kind definitions.
var texMMLFactory = newTeXMMLFactory()

func newTeXMMLFactory() *mml.Factory { return mml.NewMathJaxFactory() }
