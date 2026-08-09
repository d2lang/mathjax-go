// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go declarative registration model derived from MathJax 3.2.2 SymbolMap.ts
// and Configuration.ts. This file contains data only; parser integration lives
// elsewhere.

package tex

type mjSourceMapKind string

const (
	mjSourceCharacterMap   mjSourceMapKind = "CharacterMap"
	mjSourceCommandMap     mjSourceMapKind = "CommandMap"
	mjSourceDelimiterMap   mjSourceMapKind = "DelimiterMap"
	mjSourceEnvironmentMap mjSourceMapKind = "EnvironmentMap"
	mjSourceMacroMap       mjSourceMapKind = "MacroMap"
	mjSourceRegExpMap      mjSourceMapKind = "RegExpMap"
)

// mjSourceObject preserves JavaScript object property order. Values are
// strings, float64s, bools, nil, mjSourceList, or nested mjSourceObjects.
type mjSourceObject []mjSourceProperty

type mjSourceProperty struct {
	Name  string
	Value any
}

// mjSourceList preserves array order. Explicit JavaScript null is nil; omitted
// array elements and undefined use the distinct mjUndefined sentinel.
type mjSourceList []any

type mjSourceEntry struct {
	Name  string
	Value any
}

// mjSourceMap is one CharacterMap, CommandMap, DelimiterMap, EnvironmentMap,
// MacroMap, or RegExpMap constructor from the upstream source.
type mjSourceMap struct {
	Name     string
	Kind     mjSourceMapKind
	Source   string
	Line     int
	Parser   string
	Methods  string
	Pattern  string
	Flags    string
	Entries  []mjSourceEntry
	Dynamic  bool
	Priority int
}

type mjSourceHandler struct {
	Kind string
	Maps []string
}

type mjSourceReference struct {
	Name           string
	Implementation string
}

type mjSourcePostprocessor struct {
	Name     string
	Priority int
}

// mjSourceHook preserves the Configuration.create property name separately
// from the referenced implementation.  They differ for named callbacks such
// as mathtools' initMathtools and configMathtools functions.
type mjSourceHook struct {
	Kind           string
	Implementation string
}

// mjSourceExpandable marks option objects wrapped by util/Options.expandable.
type mjSourceExpandable struct {
	Value mjSourceObject
}

// mjSourcePackage preserves the declarative portion of Configuration.create.
// Hooks name source functions whose imperative behavior is not duplicated in
// these tables.
type mjSourcePackage struct {
	Name            string
	Source          string
	Line            int
	Handlers        []mjSourceHandler
	Items           []mjSourceReference
	Tags            []mjSourceReference
	Fallbacks       []mjSourceReference
	Options         mjSourceObject
	DynamicMapNames []string
	DynamicHandlers []mjSourceHandler
	DynamicPriority int
	Hooks           []mjSourceHook
	Postprocessors  []mjSourcePostprocessor
}

// mjSourceInfinity represents JavaScript Infinity in command payloads.
type mjSourceInfinity int8

const mjPositiveInfinity mjSourceInfinity = 1

type mjSourceUndefined int8

const mjUndefined mjSourceUndefined = 1

// mjSourceExpression preserves a source expression whose value is resolved by
// the JavaScript host rather than statically by package configuration.
type mjSourceExpression string
