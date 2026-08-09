// Copyright (c) 2018-2022 Omar Al-Ithawi and The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package color

import (
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/tex/extensions/spec"
)

func TestConfiguration(t *testing.T) {
	wantCommands := []spec.Command{
		{Name: "color", Handler: "Color"},
		{Name: "textcolor", Handler: "TextColor"},
		{Name: "definecolor", Handler: "DefineColor"},
		{Name: "colorbox", Handler: "ColorBox"},
		{Name: "fcolorbox", Handler: "FColorBox"},
	}
	wantOptions := []spec.Option{
		{Name: "padding", Value: "5px"},
		{Name: "borderWidth", Value: "2px"},
	}
	if Configuration.Name != "color" || Configuration.MacroMap != "color" {
		t.Fatalf("Configuration identity = %#v", Configuration)
	}
	if !reflect.DeepEqual(Configuration.Commands, wantCommands) {
		t.Errorf("Configuration.Commands = %#v, want %#v", Configuration.Commands, wantCommands)
	}
	if !reflect.DeepEqual(Configuration.Options, wantOptions) {
		t.Errorf("Configuration.Options = %#v, want %#v", Configuration.Options, wantOptions)
	}
}

func TestPaddingProperties(t *testing.T) {
	tests := []struct {
		padding string
		want    []spec.Attribute
	}{
		{
			"5px",
			[]spec.Attribute{
				{Name: "width", Value: "+10px"},
				{Name: "height", Value: "+5px"},
				{Name: "depth", Value: "+5px"},
				{Name: "lspace", Value: "5px"},
			},
		},
		{
			".25em",
			[]spec.Attribute{
				{Name: "width", Value: "+0.5em"},
				{Name: "height", Value: "+.25em"},
				{Name: "depth", Value: "+.25em"},
				{Name: "lspace", Value: ".25em"},
			},
		},
		{
			"thin",
			[]spec.Attribute{
				{Name: "width", Value: "+NaNthin"},
				{Name: "height", Value: "+thin"},
				{Name: "depth", Value: "+thin"},
				{Name: "lspace", Value: "thin"},
			},
		},
		{
			"1\npx",
			[]spec.Attribute{
				{Name: "width", Value: "+2" + "1\npx"},
				{Name: "height", Value: "+1\npx"},
				{Name: "depth", Value: "+1\npx"},
				{Name: "lspace", Value: "1\npx"},
			},
		},
	}
	for _, test := range tests {
		if got := PaddingProperties(test.padding); !reflect.DeepEqual(got, test.want) {
			t.Errorf("PaddingProperties(%q) = %#v, want %#v", test.padding, got, test.want)
		}
	}
}
