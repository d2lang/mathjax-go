// Evidence-only adapter: D2 calls the unmodified frozen MathJax bundle.
package mathjax

import (
 "bytes"
 "encoding/json"
 "fmt"
 "math"
 "os"
 "os/exec"
 "regexp"
 "strconv"
 "sync"
)

var cache sync.Map
var dimensions = regexp.MustCompile(`<svg[^>]+width="([0-9.]+)ex" height="([0-9.]+)ex"[^>]+>`)

func Render(tex string) (string, error) {
 if cached, ok := cache.Load(tex); ok { return cached.(string), nil }
 input, _ := json.Marshal(map[string]any{"tex": tex})
 cmd := exec.Command(os.Getenv("MATHJAX_GO_NODE"), os.Getenv("MATHJAX_GO_ORACLE_SCRIPT"), "--asset-dir", os.Getenv("MATHJAX_GO_ORACLE_DIR"))
 cmd.Stdin = bytes.NewReader(append(input, '\n'))
 cmd.Stderr = os.Stderr
 out, err := cmd.Output()
 if err != nil { return "", err }
 var result struct { SVG string `json:"svg"`; Error string `json:"error"` }
 if err := json.Unmarshal(out, &result); err != nil { return "", err }
 if result.Error != "" { return "", fmt.Errorf("original MathJax: %s", result.Error) }
 cache.Store(tex, result.SVG)
 return result.SVG, nil
}

func Measure(tex string) (int, int, error) {
 svg, err := Render(tex)
 if err != nil { return 0, 0, err }
 m := dimensions.FindStringSubmatch(svg)
 if len(m) != 3 { return 0, 0, fmt.Errorf("no SVG dimensions") }
 w, err := strconv.ParseFloat(m[1], 64)
 if err != nil { return 0, 0, err }
 h, err := strconv.ParseFloat(m[2], 64)
 if err != nil { return 0, 0, err }
 return int(math.Ceil(w * 8)), int(math.Ceil(h * 8)), nil
}
