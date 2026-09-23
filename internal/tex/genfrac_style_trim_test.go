// SPDX-License-Identifier: Apache-2.0
package tex

import "testing"

// Source-derived ParseUtil.trimSpaces guards, not additional primary renders.
func TestGenfracStyleTrimSourceControls(t *testing.T) {
	for _, c := range []struct{ name, raw, want string }{
		{"escaped-final-space", " \\\t ", "\\ "},
		{"final-tab-only", " \\\t", "\\"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := amsGenfracTrimStyle(c.raw); got != c.want {
				t.Fatalf("trim(%q)=%q; want %q", c.raw, got, c.want)
			}
		})
	}
}
