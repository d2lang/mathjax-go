// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package layout

// Parameters contains the TeX font parameters consumed by MathJax's common
// layout wrappers.
type Parameters struct {
	XHeight float64
	Quad    float64
	Num1    float64
	Num2    float64
	Num3    float64
	Denom1  float64
	Denom2  float64
	Sup1    float64
	Sup2    float64
	Sup3    float64
	Sub1    float64
	Sub2    float64
	SupDrop float64
	SubDrop float64
	Delim1  float64
	Delim2  float64
	Axis    float64
	Rule    float64
	BigOp1  float64
	BigOp2  float64
	BigOp3  float64
	BigOp4  float64
	BigOp5  float64
	Surd    float64

	ScriptSpace        float64
	NullDelimiterSpace float64
	DelimiterFactor    float64
	DelimiterShortfall float64
	MinRuleThickness   float64
	SeparationFactor   float64
	ExtraIC            float64
}

// TeXParameters is SVG TeXFont.defaultParams in MathJax 3.2.2.
var TeXParameters = Parameters{
	XHeight: .442, Quad: 1,
	Num1: .676, Num2: .394, Num3: .444,
	Denom1: .686, Denom2: .345,
	Sup1: .413, Sup2: .363, Sup3: .289,
	Sub1: .15, Sub2: .247, SupDrop: .386, SubDrop: .05,
	Delim1: 2.39, Delim2: 1, Axis: .25, Rule: .06,
	BigOp1: .111, BigOp2: .167, BigOp3: .2, BigOp4: .6, BigOp5: .1,
	Surd:        .075,
	ScriptSpace: .05, NullDelimiterSpace: .12,
	DelimiterFactor: 901, DelimiterShortfall: .3,
	MinRuleThickness: 1.25, SeparationFactor: 1.75, ExtraIC: .033,
}
