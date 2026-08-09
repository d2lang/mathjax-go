// Copyright (c) 2018-2022 Omar Al-Ithawi and The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package color

import (
	"errors"
	"testing"
)

func TestGetColorNamedAndPassThrough(t *testing.T) {
	model := NewModel()
	tests := []struct {
		colorModel string
		definition string
		want       string
	}{
		{"", "Red", "#ED1B23"},
		{"named", "Blue", "#2D2F92"},
		{"named", "red", "red"},
		{"", "#f80", "#f80"},
		{"named", "rebeccapurple", "rebeccapurple"},
	}
	for _, test := range tests {
		got, err := model.GetColor(test.colorModel, test.definition)
		if err != nil || got != test.want {
			t.Errorf("GetColor(%q, %q) = %q, %v; want %q, nil", test.colorModel, test.definition, got, err, test.want)
		}
	}
}

func TestNumericModels(t *testing.T) {
	model := NewModel()
	tests := []struct {
		colorModel string
		definition string
		want       string
	}{
		{"rgb", "0.5,0,1", "#7f00ff"},
		{"rgb", " 1. , .0 , .999 ", "#ff00fe"},
		{"rgb", "\u00a0.5 , 0, 1\uFEFF", "#7f00ff"},
		{"RGB", "128,0,255", "#8000ff"},
		{"RGB", " 000000128, 00, 0255 ", "#8000ff"},
		{"gray", "0", "#000000"},
		{"gray", " .5 ", "#7f7f7f"},
		{"gray", "\uFEFF1\u00a0", "#ffffff"},
	}
	for _, test := range tests {
		got, err := model.GetColor(test.colorModel, test.definition)
		if err != nil || got != test.want {
			t.Errorf("GetColor(%q, %q) = %q, %v; want %q, nil", test.colorModel, test.definition, got, err, test.want)
		}
	}
}

func TestColorModelErrors(t *testing.T) {
	tests := []struct {
		colorModel string
		definition string
		id         string
		message    string
	}{
		{"cmyk", "0,0,0,0", "UndefinedColorModel", "Color model 'cmyk' not defined"},
		{"rgb", "0,1", "ModelArg1", "Color values for the rgb model require 3 numbers"},
		{"rgb", "-0.1,0,0", "InvalidDecimalNumber", "Invalid decimal number"},
		{"rgb", "1.1,0,0", "ModelArg2", "Color values for the rgb model must be between 0 and 1"},
		{"RGB", "0,1", "ModelArg1", "Color values for the RGB model require 3 numbers"},
		{"RGB", "1.0,0,0", "InvalidNumber", "Invalid number"},
		{"RGB", "256,0,0", "ModelArg2", "Color values for the RGB model must be between 0 and 255"},
		{"RGB", "999999999999999999999999,0,0", "ModelArg2", "Color values for the RGB model must be between 0 and 255"},
		{"gray", "-0.1", "InvalidDecimalNumber", "Invalid decimal number"},
		{"gray", "1.1", "ModelArg2", "Color values for the gray model must be between 0 and 1"},
	}
	model := NewModel()
	for _, test := range tests {
		_, err := model.GetColor(test.colorModel, test.definition)
		var colorErr *Error
		if !errors.As(err, &colorErr) {
			t.Errorf("GetColor(%q, %q) error = %T %v, want *Error", test.colorModel, test.definition, err, err)
			continue
		}
		if colorErr.ID != test.id || colorErr.Message != test.message || err.Error() != test.message {
			t.Errorf("GetColor(%q, %q) error = %#v, want ID %q message %q", test.colorModel, test.definition, colorErr, test.id, test.message)
		}
	}
}

func TestDefineColorIsParserLocal(t *testing.T) {
	first, second := NewModel(), NewModel()
	if err := first.DefineColor("RGB", "accent", "12,34,56"); err != nil {
		t.Fatal(err)
	}
	if err := first.DefineColor("named", "Red", "#010203"); err != nil {
		t.Fatal(err)
	}
	if got, _ := first.GetColor("named", "accent"); got != "#0c2238" {
		t.Errorf("first accent = %q, want #0c2238", got)
	}
	if got, _ := second.GetColor("named", "accent"); got != "accent" {
		t.Errorf("second accent = %q, want pass-through", got)
	}
	if got, _ := first.GetColor("named", "Red"); got != "#010203" {
		t.Errorf("user Red = %q, want override", got)
	}
	if got, _ := second.GetColor("named", "Red"); got != "#ED1B23" {
		t.Errorf("built-in Red = %q", got)
	}
}

func TestZeroValueModel(t *testing.T) {
	var model Model
	if err := model.DefineColor("gray", "middle", ".5"); err != nil {
		t.Fatal(err)
	}
	if got, err := model.GetColor("named", "middle"); err != nil || got != "#7f7f7f" {
		t.Fatalf("zero-value model = %q, %v", got, err)
	}
}
