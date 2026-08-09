// Copyright (c) 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package cancel

import (
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/tex/extensions/enclose"
	"github.com/d2lang/mathjax-go/internal/tex/extensions/spec"
)

func TestConfiguration(t *testing.T) {
	wantCommands := []spec.Command{
		{Name: "cancel", Handler: "Cancel", Arguments: []string{"updiagonalstrike"}},
		{Name: "bcancel", Handler: "Cancel", Arguments: []string{"downdiagonalstrike"}},
		{Name: "xcancel", Handler: "Cancel", Arguments: []string{"updiagonalstrike downdiagonalstrike"}},
		{Name: "cancelto", Handler: "CancelTo"},
	}
	if Configuration.Name != "cancel" || Configuration.MacroMap != "cancel" {
		t.Fatalf("Configuration identity = %#v", Configuration)
	}
	if !reflect.DeepEqual(Configuration.Commands, wantCommands) {
		t.Fatalf("Configuration.Commands = %#v, want %#v", Configuration.Commands, wantCommands)
	}
	if !reflect.DeepEqual(Configuration.AllowedOptions, enclose.AllowedOptions) {
		t.Fatalf("Configuration.AllowedOptions = %v, want enclose options", Configuration.AllowedOptions)
	}
}

func TestCancelToRecipe(t *testing.T) {
	if CancelToNotation != "updiagonalstrike updiagonalarrow northeastarrow" {
		t.Errorf("CancelToNotation = %q", CancelToNotation)
	}
	want := []spec.Attribute{
		{Name: "depth", Value: "-.1em"},
		{Name: "height", Value: "+.1em"},
		{Name: "voffset", Value: ".1em"},
	}
	if !reflect.DeepEqual(CancelToPadding, want) {
		t.Errorf("CancelToPadding = %#v, want %#v", CancelToPadding, want)
	}
}
