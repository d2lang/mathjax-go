// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package svg

import "github.com/d2lang/mathjax-go/internal/mml"

// CommonWrapper.createMo and MmlMfenced obtain their synthetic operators from
// the MML factory, including the complete kind and global default layers.
// Create clones the immutable registry maps, so concurrent renders are safe.
var syntheticMMLFactory = mml.NewMathJaxFactory()
