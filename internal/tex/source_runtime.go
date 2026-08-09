// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/input/tex/{NodeUtil,SymbolMap,Symbol,MapHandler,ParseMethods,
// Configuration}.ts and base/BaseConfiguration.ts.

package tex

import (
	"fmt"
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func lookupMJSourceEntry(kind mjSourceMapKind, name string) (mjSourceMap, mjSourceEntry, bool) {
	for mapIndex := len(mjSourceMaps) - 1; mapIndex >= 0; mapIndex-- {
		sourceMap := mjSourceMaps[mapIndex]
		if sourceMap.Kind != kind {
			continue
		}
		for entryIndex := len(sourceMap.Entries) - 1; entryIndex >= 0; entryIndex-- {
			entry := sourceMap.Entries[entryIndex]
			entryName := entry.Name
			if kind == mjSourceDelimiterMap {
				entryName = strings.TrimPrefix(entryName, "\\")
			}
			if entryName == name {
				return sourceMap, entry, true
			}
		}
	}
	return mjSourceMap{}, mjSourceEntry{}, false
}

func lookupMJSourceSymbol(name string) (*mml.Node, bool) {
	for _, kind := range []mjSourceMapKind{mjSourceCharacterMap, mjSourceDelimiterMap} {
		sourceMap, entry, ok := lookupMJSourceEntry(kind, name)
		if !ok {
			continue
		}
		character, attributes, ok := sourceSymbolValue(entry.Value)
		if !ok {
			return nil, false
		}
		tokenKind := "mo"
		class := mml.TeXClassOrd
		switch {
		case strings.Contains(sourceMap.Parser, "mathchar0mi"):
			tokenKind = "mi"
			// ParseMethods.mathchar0mi supplies italic when the source entry
			// has no attribute object.  This is an explicit parser attribute,
			// distinct from MmlMi's later one-character inheritance default.
			if attributes == nil {
				attributes = mjSourceObject{{Name: "mathvariant", Value: "italic"}}
			}
		case strings.Contains(sourceMap.Parser, "mathchar7"):
			tokenKind = "mi"
			// ParseMethods.mathchar7 supplies normal for its current-family
			// symbols (notably upper-case Greek) unless the entry overrides it.
			if attributes == nil {
				attributes = mjSourceObject{{Name: "mathvariant", Value: "normal"}}
			}
		case strings.Contains(sourceMap.Parser, "delimiter"):
			class = mml.TeXClassOrd
			if _, present := sourceObjectValue(attributes, "fence"); !present {
				attributes = append(attributes, mjSourceProperty{Name: "fence", Value: false})
			}
			if _, present := sourceObjectValue(attributes, "stretchy"); !present {
				attributes = append(attributes, mjSourceProperty{Name: "stretchy", Value: false})
			}
		}
		n := token(tokenKind, character)
		n.TeXClass = class
		applySourceObject(n, attributes)
		if tokenKind == "mo" && strings.Contains(sourceMap.Parser, "mathchar0mo") {
			n.Attributes.Set("stretchy", false)
			n.SetProperty("fixStretchy", true)
			n.SetProperty("in-lists", "fixStretchy")
		}
		return n, true
	}
	return nil, false
}

func sourceSymbolValue(value any) (string, mjSourceObject, bool) {
	switch value := value.(type) {
	case string:
		return value, nil, true
	case mjSourceList:
		if len(value) == 0 {
			return "", nil, false
		}
		character, ok := value[0].(string)
		if !ok {
			return "", nil, false
		}
		var attributes mjSourceObject
		if len(value) > 1 {
			attributes, _ = value[1].(mjSourceObject)
		}
		return character, attributes, true
	}
	return "", nil, false
}

func applySourceObject(n *mml.Node, object mjSourceObject) {
	for _, property := range object {
		switch property.Name {
		case "texClass":
			if class, ok := sourceInt(property.Value); ok {
				n.TeXClass = mml.TeXClass(class)
				n.SetProperty(property.Name, class)
			}
		case "movablelimits":
			n.SetProperty(property.Name, property.Value)
			if n.Kind == "mo" || n.Kind == "mstyle" {
				n.Attributes.Set(property.Name, property.Value)
			}
		case "autoOP", "fnOP", "movesupsub", "subsupOK", "texprimestyle", "useHeight", "variantForm", "withDelims", "mathaccent", "open", "close":
			n.SetProperty(property.Name, property.Value)
		case "inferred":
			// NodeUtil.setProperties intentionally ignores this marker.
		default:
			n.Attributes.Set(property.Name, property.Value)
		}
	}
}

func sourceObjectValue(object mjSourceObject, name string) (any, bool) {
	for _, property := range object {
		if property.Name == name {
			return property.Value, true
		}
	}
	return nil, false
}

func sourceInt(value any) (int, bool) {
	switch value := value.(type) {
	case int:
		return value, true
	case float64:
		return int(value), value == float64(int(value))
	}
	return 0, false
}

func sourceString(value any) (string, bool) {
	valueString, ok := value.(string)
	return valueString, ok
}

func sourceBool(value any) (bool, bool) {
	valueBool, ok := value.(bool)
	return valueBool, ok
}

func sourceHandler(value any) (string, mjSourceList, bool) {
	switch value := value.(type) {
	case string:
		return value, nil, true
	case mjSourceList:
		if len(value) == 0 {
			return "", nil, false
		}
		handler, ok := value[0].(string)
		if !ok {
			return "", nil, false
		}
		return handler, value[1:], true
	}
	return "", nil, false
}

func sourceValueString(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
