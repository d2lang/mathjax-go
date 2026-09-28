// The same public-API probe is compiled from each compared module checkout.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	mathjax "github.com/d2lang/mathjax-go"
)

type request struct {
	TeX     string `json:"tex"`
	Display *bool  `json:"display"`
	Options struct {
		Display *bool `json:"Display"`
	} `json:"options"`
}

type response struct {
	SVG   string `json:"svg"`
	Error string `json:"error,omitempty"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	encoder := json.NewEncoder(os.Stdout)
	for scanner.Scan() {
		var input request
		var output response
		if err := json.Unmarshal(scanner.Bytes(), &input); err != nil {
			output.Error = err.Error()
		} else {
			options := mathjax.DefaultOptions()
			if input.Options.Display != nil {
				options.Display = *input.Options.Display
			}
			if input.Display != nil {
				options.Display = *input.Display
			}
			var err error
			output.SVG, err = mathjax.RenderWithOptions(input.TeX, options)
			if err != nil {
				output.Error = err.Error()
			}
		}
		if err := encoder.Encode(output); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
