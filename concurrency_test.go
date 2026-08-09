// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package mathjax

import (
	"fmt"
	"sync"
	"testing"
)

func TestRenderIsConcurrent(t *testing.T) {
	formulas := []string{`a+b=c`, `\frac{1}{2}`, `x_1^2`}
	want := make([]string, len(formulas))
	for i, formula := range formulas {
		var err error
		want[i], err = Render(formula)
		if err != nil {
			t.Fatal(err)
		}
	}

	const workers = 24
	start := make(chan struct{})
	errors := make(chan error, workers)
	var wait sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wait.Add(1)
		go func(worker int) {
			defer wait.Done()
			<-start
			for iteration := 0; iteration < 12; iteration++ {
				index := (worker + iteration) % len(formulas)
				got, err := Render(formulas[index])
				if err != nil {
					errors <- err
					return
				}
				if got != want[index] {
					errors <- fmt.Errorf("worker %d iteration %d returned nondeterministic SVG", worker, iteration)
					return
				}
			}
		}(worker)
	}
	close(start)
	wait.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}
