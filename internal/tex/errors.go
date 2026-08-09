// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Source: ts/input/tex/TexError.ts.

package tex

import "fmt"

// Error is the Go equivalent of MathJax's TexError.  ID is deliberately kept
// separate from the rendered message because the JavaScript parser and its
// tests use the stable identifier when classifying failures.
type Error struct {
	ID      string
	Message string
}

func (e *Error) Error() string { return e.Message }

func texError(id, format string, args ...any) error {
	return &Error{ID: id, Message: fmt.Sprintf(format, args...)}
}
