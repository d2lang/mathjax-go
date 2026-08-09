// Copyright (c) 2015-2021 Martin Hensel
// Copyright (c) 2026 luo-studio
// Copyright (c) 2026 erweixin (upstream RaTeX)
// SPDX-License-Identifier: Apache-2.0 AND MIT
//
// Go translation and modification of mhchemparser 4.1.1. Initial Go
// translation structure adapted from github.com/Luo-Studio/go-tex (MIT).

package mhchem

import "fmt"

// Error preserves the identifier and message pair thrown by mhchemparser.
// TeX integrations can pass these fields directly to their native TexError.
type Error struct {
	ID      string
	Message string
}

func (e *Error) Error() string { return e.Message }

func parserError(id, message string) error { return &Error{ID: id, Message: message} }

func errMsg(format string, args ...any) error {
	return parserError("MhchemError", "mhchem: "+fmt.Sprintf(format, args...))
}

var errExtraClose = &Error{
	ID:      "ExtraCloseMissingOpen",
	Message: "Extra close brace or missing open brace",
}
