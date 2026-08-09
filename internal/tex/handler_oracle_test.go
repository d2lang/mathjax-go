// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"
)

type handlerOracleManifest struct {
	SchemaVersion int `json:"schema_version"`
	Oracle        struct {
		Project              string   `json:"project"`
		Version              string   `json:"version"`
		GitCommit            string   `json:"git_commit"`
		NPMPackage           string   `json:"npm_package"`
		NPMIntegrity         string   `json:"npm_integrity"`
		NPMTarballSHA256     string   `json:"npm_tarball_sha256"`
		Runtime              string   `json:"runtime"`
		State                string   `json:"state"`
		Serializer           string   `json:"serializer"`
		Display              bool     `json:"display"`
		FreshDocumentPerCase bool     `json:"fresh_document_per_case"`
		Packages             []string `json:"packages"`
		Provenance           string   `json:"provenance"`
	} `json:"oracle"`
	HandlerScope struct {
		Name     string   `json:"name"`
		Packages []string `json:"packages"`
		MapKinds []string `json:"map_kinds"`
	} `json:"handler_scope"`
	HandlerCount int `json:"handler_count"`
	Cases        []struct {
		Name     string   `json:"name"`
		Handlers []string `json:"handlers"`
		TeX      string   `json:"tex"`
		MML      string   `json:"mml"`
		Class    string   `json:"class"`
		Variant  string   `json:"variant,omitempty"`
		ErrorID  string   `json:"error_id,omitempty"`
	} `json:"cases"`
}

func TestMathJax322HandlerOracle(t *testing.T) {
	data, err := os.ReadFile("testdata/handler_oracle_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest handlerOracleManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}

	if manifest.SchemaVersion != 1 {
		t.Fatalf("oracle schema version = %d, want 1", manifest.SchemaVersion)
	}
	if manifest.Oracle.Project != "MathJax" || manifest.Oracle.Version != "3.2.2" {
		t.Fatalf("oracle source = %q %q, want MathJax 3.2.2", manifest.Oracle.Project, manifest.Oracle.Version)
	}
	if manifest.Oracle.GitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" {
		t.Fatalf("oracle git commit = %q", manifest.Oracle.GitCommit)
	}
	if manifest.Oracle.NPMIntegrity != "sha512-Bt+SSVU8eBG27zChVewOicYs7Xsdt40qm4+UpHyX7k0/O9NliPc+x77k1/FEsPsjKPZGJvtRZM1vO+geW0OhGw==" {
		t.Fatalf("oracle npm integrity = %q", manifest.Oracle.NPMIntegrity)
	}
	if manifest.Oracle.State != "MathJax._.core.MathItem.STATE.COMPILED" || manifest.Oracle.Serializer != "MmlNode.toString()" {
		t.Fatalf("oracle extraction = state %q, serializer %q", manifest.Oracle.State, manifest.Oracle.Serializer)
	}
	if !manifest.Oracle.Display || !manifest.Oracle.FreshDocumentPerCase {
		t.Fatal("oracle must use display mode and a fresh MathJax document per case")
	}
	if manifest.HandlerScope.Name != "original-95-selected-package-handler-set" {
		t.Fatalf("oracle handler scope = %q", manifest.HandlerScope.Name)
	}

	required := selectedSourceHandlerIDs()
	if manifest.HandlerCount != len(required) {
		t.Fatalf("manifest handler_count = %d, source tables contain %d", manifest.HandlerCount, len(required))
	}
	covered := make(map[string]bool)
	for _, test := range manifest.Cases {
		if test.Name == "" || test.TeX == "" || test.MML == "" || len(test.Handlers) == 0 {
			t.Fatalf("incomplete oracle case: %#v", test)
		}
		for _, handler := range test.Handlers {
			covered[handler] = true
		}
		switch test.Class {
		case "valid":
			if strings.HasPrefix(test.MML, "math([merror(") {
				t.Fatalf("valid oracle case %q contains merror: %s", test.Name, test.MML)
			}
		case "merror":
			if test.ErrorID == "" {
				t.Fatalf("merror oracle case %q has no typed error_id", test.Name)
			}
			if !strings.HasPrefix(test.MML, "math([merror(") {
				t.Fatalf("merror oracle case %q is not an merror: %s", test.Name, test.MML)
			}
		default:
			t.Fatalf("oracle case %q has unknown class %q", test.Name, test.Class)
		}
	}

	var uncovered []string
	for _, handler := range required {
		if !covered[handler] {
			uncovered = append(uncovered, handler)
		}
	}
	if len(uncovered) != 0 {
		t.Fatalf("%d uncovered handler IDs: %s", len(uncovered), strings.Join(uncovered, ", "))
	}
	var extra []string
	for handler := range covered {
		if !containsSorted(required, handler) {
			extra = append(extra, handler)
		}
	}
	sort.Strings(extra)
	if len(extra) != 0 {
		t.Fatalf("%d oracle handler IDs absent from selected source tables: %s", len(extra), strings.Join(extra, ", "))
	}

	for _, test := range manifest.Cases {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			// This manifest intentionally loads official empheq and newcommand
			// components beyond D2's runtime package list.  Keep that additive
			// source-handler coverage without exposing those registrations from
			// the public fixed-D2 compiler.
			root, err := (&Compiler{augmentedPackages: true}).Compile(test.TeX, true)
			if err != nil {
				t.Fatal(err)
			}
			if got := mmlString(root); got != test.MML {
				t.Fatalf("MML mismatch:\n got %s\nwant %s", got, test.MML)
			}
		})
	}
}

func selectedSourceHandlerIDs() []string {
	selected := [][]mjSourceMap{
		mjSourceAMSMaps,
		mjSourceMathtoolsMaps,
		mjSourceAmsCdBraketCasesMaps,
		mjSourcePhysicsMaps,
		mjSourceEmpheqNewcommandMaps,
	}
	set := make(map[string]bool)
	for _, maps := range selected {
		for _, sourceMap := range maps {
			if sourceMap.Kind != mjSourceCommandMap && sourceMap.Kind != mjSourceEnvironmentMap && sourceMap.Kind != mjSourceMacroMap {
				continue
			}
			for _, entry := range sourceMap.Entries {
				handler := ""
				switch value := entry.Value.(type) {
				case string:
					handler = value
				case mjSourceList:
					if len(value) != 0 {
						handler, _ = value[0].(string)
					}
				}
				if handler != "" {
					set[handler] = true
				}
			}
		}
	}
	ids := make([]string, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func containsSorted(values []string, value string) bool {
	i := sort.SearchStrings(values, value)
	return i < len(values) && values[i] == value
}
