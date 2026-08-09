// Copyright (c) 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package enclose

import (
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/tex/extensions/spec"
)

func TestConfiguration(t *testing.T) {
	want := spec.Package{
		Name:     "enclose",
		MacroMap: "enclose",
		Commands: []spec.Command{{Name: "enclose", Handler: "Enclose"}},
		AllowedOptions: []string{
			"data-arrowhead", "color", "mathcolor", "background",
			"mathbackground", "data-padding", "data-thickness",
		},
	}
	if !reflect.DeepEqual(Configuration, want) {
		t.Fatalf("Configuration = %#v, want %#v", Configuration, want)
	}
}
