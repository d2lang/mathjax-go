// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package oracle

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/d2lang/mathjax-go/internal/pipeline"
)

func TestFrozenOracle(t *testing.T) {
	runner, err := FromEnvironment()
	if errors.Is(err, ErrUnavailable) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	results, err := runner.RenderBatch(ctx, []Case{
		{TeX: `a + b = c`, Options: pipeline.DefaultOptions()},
		{TeX: `\frac{1}{2}`, Options: pipeline.DefaultOptions()},
	})
	if err != nil {
		t.Fatal(err)
	}
	for i, result := range results {
		if result.Error != "" {
			t.Fatalf("case %d: %s", i, result.Error)
		}
		if !strings.HasPrefix(result.SVG, `<svg `) || !strings.HasSuffix(result.SVG, `</svg>`) {
			t.Fatalf("case %d did not return a bare SVG: %q", i, result.SVG)
		}
	}
}
