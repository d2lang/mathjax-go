// Copyright (c) 2015-2021 Martin Hensel
// Copyright (c) 2026 luo-studio
// Copyright (c) 2026 erweixin (upstream RaTeX)
// SPDX-License-Identifier: Apache-2.0 AND MIT
//
// Go translation and modification of mhchemparser 4.1.1 and upstream RaTeX
// table-loading behavior; see PROVENANCE.md for the pinned sources.

package mhchem

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

// machinesJSON and patternsJSON are generated from the private transition and
// pattern tables in mhchemparser 4.1.1's dist/mhchemParser.js.  Keeping the
// tables as data preserves JavaScript property insertion order in the expanded
// transition arrays while the Go engine implements the behavior natively.
//
//go:embed data/machines.json
var machinesJSON []byte

//go:embed data/patterns.json
var patternsJSON []byte

type Machines map[string]MachineDef

type MachineDef struct {
	Transitions     map[string][]Transition `json:"transitions"`
	HasLocalActions bool                    `json:"hasLocalActions"`
}

type Transition struct {
	Pattern string `json:"pattern"`
	Task    Task   `json:"task"`
}

type Task struct {
	NextState  *string      `json:"nextState,omitempty"`
	Revisit    bool         `json:"revisit,omitempty"`
	ToContinue bool         `json:"toContinue,omitempty"`
	Action     []ActionSpec `json:"action_,omitempty"`
}

type ActionSpec struct {
	Type   string          `json:"type_"`
	Option json.RawMessage `json:"option,omitempty"`
}

type Data struct {
	Machines Machines
	Regexes  map[string]*jsRegexp
}

var (
	loadedData     *Data
	loadedDataOnce sync.Once
	loadedDataErr  error
)

func LoadData() (*Data, error) {
	loadedDataOnce.Do(func() {
		loadedData, loadedDataErr = loadDataImpl()
	})
	return loadedData, loadedDataErr
}

func loadDataImpl() (*Data, error) {
	machineTable := Machines{}
	if err := json.Unmarshal(machinesJSON, &machineTable); err != nil {
		return nil, fmt.Errorf("mhchem: parse machine table: %w", err)
	}
	var patterns struct {
		Regex map[string]string `json:"regex"`
	}
	if err := json.Unmarshal(patternsJSON, &patterns); err != nil {
		return nil, fmt.Errorf("mhchem: parse pattern table: %w", err)
	}
	regexes := make(map[string]*jsRegexp, len(patterns.Regex))
	for name, source := range patterns.Regex {
		re, err := compileJSRegexp(source)
		if err != nil {
			return nil, fmt.Errorf("mhchem: compile pattern %q (%q): %w", name, source, err)
		}
		regexes[name] = re
	}
	return &Data{Machines: machineTable, Regexes: regexes}, nil
}

func MustData() *Data {
	data, err := LoadData()
	if err != nil {
		panic(err)
	}
	return data
}
