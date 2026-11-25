// Copyright (c) 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Code generated from MathJax 3.2.2 ts/core/MmlTree/OperatorDictionary.ts;
// DO NOT EDIT.

package tex

type mjOperatorProperty struct {
	Name  string
	Value any
}

type mjOperatorProperties []mjOperatorProperty

type mjOperatorDefinition struct {
	Lspace     int
	Rspace     int
	TexClass   int
	Properties mjOperatorProperties
}

type mjOperatorRange struct {
	First      int
	Last       int
	TexClass   int
	Kind       string
	Variant    string
	HasVariant bool
	SourceLine int
}

type mjOperatorSpacing struct {
	TexClass   int
	Lspace     int
	Rspace     int
	SourceLine int
}

var mjOperatorRanges = []mjOperatorRange{
	{First: 32, Last: 127, TexClass: 3, Kind: "mo", SourceLine: 84},
	{First: 160, Last: 191, TexClass: 0, Kind: "mo", SourceLine: 85},
	{First: 192, Last: 591, TexClass: 0, Kind: "mi", SourceLine: 86},
	{First: 688, Last: 879, TexClass: 0, Kind: "mo", SourceLine: 87},
	{First: 880, Last: 6688, TexClass: 0, Kind: "mi", SourceLine: 88},
	{First: 6832, Last: 6911, TexClass: 0, Kind: "mo", SourceLine: 89},
	{First: 6912, Last: 7615, TexClass: 0, Kind: "mi", SourceLine: 90},
	{First: 7616, Last: 7679, TexClass: 0, Kind: "mo", SourceLine: 91},
	{First: 7680, Last: 8191, TexClass: 0, Kind: "mi", SourceLine: 92},
	{First: 8192, Last: 8303, TexClass: 0, Kind: "mo", SourceLine: 93},
	{First: 8304, Last: 8351, TexClass: 0, Kind: "mo", SourceLine: 94},
	{First: 8448, Last: 8527, TexClass: 0, Kind: "mi", SourceLine: 95},
	{First: 8528, Last: 8591, TexClass: 0, Kind: "mn", SourceLine: 96},
	{First: 8592, Last: 8703, TexClass: 3, Kind: "mo", SourceLine: 97},
	{First: 8704, Last: 8959, TexClass: 2, Kind: "mo", SourceLine: 98},
	{First: 8960, Last: 9215, TexClass: 0, Kind: "mo", SourceLine: 99},
	{First: 9312, Last: 9471, TexClass: 0, Kind: "mn", SourceLine: 100},
	{First: 9472, Last: 10223, TexClass: 0, Kind: "mo", SourceLine: 101},
	{First: 10224, Last: 10239, TexClass: 3, Kind: "mo", SourceLine: 102},
	{First: 10240, Last: 10495, TexClass: 0, Kind: "mtext", SourceLine: 103},
	{First: 10496, Last: 10623, TexClass: 3, Kind: "mo", SourceLine: 104},
	{First: 10624, Last: 10751, TexClass: 0, Kind: "mo", SourceLine: 105},
	{First: 10752, Last: 11007, TexClass: 2, Kind: "mo", SourceLine: 106},
	{First: 11008, Last: 11055, TexClass: 0, Kind: "mo", SourceLine: 107},
	{First: 11056, Last: 11087, TexClass: 3, Kind: "mo", SourceLine: 108},
	{First: 11088, Last: 11263, TexClass: 0, Kind: "mo", SourceLine: 109},
	{First: 11264, Last: 11744, TexClass: 0, Kind: "mi", SourceLine: 110},
	{First: 11776, Last: 11903, TexClass: 0, Kind: "mo", SourceLine: 111},
	{First: 11904, Last: 12255, TexClass: 0, Kind: "mi", SourceLine: 112, Variant: "normal", HasVariant: true},
	{First: 12272, Last: 12351, TexClass: 0, Kind: "mo", SourceLine: 113},
	{First: 12352, Last: 42143, TexClass: 0, Kind: "mi", SourceLine: 114, Variant: "normal", HasVariant: true},
	{First: 42192, Last: 43055, TexClass: 0, Kind: "mi", SourceLine: 115},
	{First: 43056, Last: 43071, TexClass: 0, Kind: "mn", SourceLine: 116},
	{First: 43072, Last: 55295, TexClass: 0, Kind: "mi", SourceLine: 117},
	{First: 63744, Last: 64255, TexClass: 0, Kind: "mi", SourceLine: 118, Variant: "normal", HasVariant: true},
	{First: 64256, Last: 65023, TexClass: 0, Kind: "mi", SourceLine: 119},
	{First: 65024, Last: 65135, TexClass: 0, Kind: "mo", SourceLine: 120},
	{First: 65136, Last: 65791, TexClass: 0, Kind: "mi", SourceLine: 121},
	{First: 65792, Last: 65935, TexClass: 0, Kind: "mn", SourceLine: 122},
	{First: 65936, Last: 74751, TexClass: 0, Kind: "mi", SourceLine: 123, Variant: "normal", HasVariant: true},
	{First: 74752, Last: 74879, TexClass: 0, Kind: "mn", SourceLine: 124},
	{First: 74880, Last: 113823, TexClass: 0, Kind: "mi", SourceLine: 125, Variant: "normal", HasVariant: true},
	{First: 113824, Last: 119391, TexClass: 0, Kind: "mo", SourceLine: 126},
	{First: 119648, Last: 119679, TexClass: 0, Kind: "mn", SourceLine: 127},
	{First: 119808, Last: 120781, TexClass: 0, Kind: "mi", SourceLine: 128},
	{First: 120782, Last: 120831, TexClass: 0, Kind: "mn", SourceLine: 129},
	{First: 122624, Last: 129023, TexClass: 0, Kind: "mo", SourceLine: 130},
	{First: 129024, Last: 129279, TexClass: 3, Kind: "mo", SourceLine: 131},
	{First: 129280, Last: 129535, TexClass: 0, Kind: "mo", SourceLine: 132},
	{First: 131072, Last: 195103, TexClass: 0, Kind: "mi", SourceLine: 133, Variant: "normnal", HasVariant: true},
}

var mjOperatorMMLSpacing = []mjOperatorSpacing{
	{TexClass: 0, Lspace: 0, Rspace: 0, SourceLine: 159},
	{TexClass: 1, Lspace: 1, Rspace: 2, SourceLine: 160},
	{TexClass: 2, Lspace: 3, Rspace: 3, SourceLine: 161},
	{TexClass: 3, Lspace: 4, Rspace: 4, SourceLine: 162},
	{TexClass: 4, Lspace: 0, Rspace: 0, SourceLine: 163},
	{TexClass: 5, Lspace: 0, Rspace: 0, SourceLine: 164},
	{TexClass: 6, Lspace: 0, Rspace: 3, SourceLine: 165},
}
