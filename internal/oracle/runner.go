// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

// Package oracle runs an optional, frozen JavaScript differential oracle from
// tests. It is not imported by the production mathjax-go package.
package oracle

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/d2lang/mathjax-go/internal/pipeline"
)

const (
	mathJaxAsset  = "mathjax.js"
	polyfillAsset = "polyfills.js"
	setupAsset    = "setup.js"

	mathJaxSHA256  = "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869"
	polyfillSHA256 = "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01"
	setupSHA256    = "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881"
)

// ErrUnavailable means the optional oracle was not configured.
var ErrUnavailable = errors.New("mathjax-go: JavaScript oracle is unavailable")

// Case is one oracle conversion request.
type Case struct {
	TeX     string           `json:"tex"`
	Options pipeline.Options `json:"options"`
}

// Result is one oracle conversion response.
type Result struct {
	SVG   string `json:"svg,omitempty"`
	Error string `json:"error,omitempty"`
}

// Runner describes a verified Node.js oracle installation.
type Runner struct {
	node      string
	assetDir  string
	oracleMJS string
}

// FromEnvironment verifies and returns the optional oracle. Set
// MATHJAX_GO_ORACLE_DIR to a directory containing the three frozen D2 assets.
// MATHJAX_GO_NODE can override the Node.js executable.
func FromEnvironment() (*Runner, error) {
	assetDir := os.Getenv("MATHJAX_GO_ORACLE_DIR")
	if assetDir == "" {
		return nil, ErrUnavailable
	}
	node := os.Getenv("MATHJAX_GO_NODE")
	if node == "" {
		var err error
		node, err = exec.LookPath("node")
		if err != nil {
			return nil, fmt.Errorf("%w: node executable not found", ErrUnavailable)
		}
	}
	assets := map[string]string{
		mathJaxAsset:  mathJaxSHA256,
		polyfillAsset: polyfillSHA256,
		setupAsset:    setupSHA256,
	}
	for name, want := range assets {
		if err := verifySHA256(filepath.Join(assetDir, name), want); err != nil {
			return nil, err
		}
	}
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("mathjax-go: locate oracle package source")
	}
	oracleMJS := filepath.Join(filepath.Dir(sourceFile), "..", "..", "testdata", "differential", "oracle.mjs")
	if _, err := os.Stat(oracleMJS); err != nil {
		return nil, fmt.Errorf("mathjax-go: locate oracle driver: %w", err)
	}
	return &Runner{node: node, assetDir: assetDir, oracleMJS: oracleMJS}, nil
}

func verifySHA256(path, want string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("mathjax-go: open oracle asset %s: %w", path, err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("mathjax-go: hash oracle asset %s: %w", path, err)
	}
	got := hex.EncodeToString(hash.Sum(nil))
	if got != want {
		return fmt.Errorf("mathjax-go: oracle asset %s has SHA-256 %s, want %s", path, got, want)
	}
	return nil
}

// RenderBatch renders cases in one Node.js process so the expensive frozen
// component is loaded only once.
func (r *Runner) RenderBatch(ctx context.Context, cases []Case) ([]Result, error) {
	command := exec.CommandContext(ctx, r.node, r.oracleMJS, "--asset-dir", r.assetDir)
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}
	command.Stderr = os.Stderr
	if err := command.Start(); err != nil {
		return nil, err
	}

	encodeErr := make(chan error, 1)
	go func() {
		encoder := json.NewEncoder(stdin)
		for _, test := range cases {
			if err := encoder.Encode(test); err != nil {
				encodeErr <- err
				stdin.Close()
				return
			}
		}
		encodeErr <- stdin.Close()
	}()

	results := make([]Result, 0, len(cases))
	scanner := bufio.NewScanner(stdout)
	buffer := make([]byte, 64*1024)
	scanner.Buffer(buffer, 16*1024*1024)
	for scanner.Scan() {
		var result Result
		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			return nil, fmt.Errorf("mathjax-go: decode oracle response: %w", err)
		}
		results = append(results, result)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if err := <-encodeErr; err != nil {
		return nil, err
	}
	if err := command.Wait(); err != nil {
		return nil, fmt.Errorf("mathjax-go: oracle process: %w", err)
	}
	if len(results) != len(cases) {
		return nil, fmt.Errorf("mathjax-go: oracle returned %d results for %d cases", len(results), len(cases))
	}
	return results, nil
}
