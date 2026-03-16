package custom

import (
	"fmt"
	"strings"
)

func Interpolate(template string, args map[string]any) string {
	out := template
	for k, v := range args {
		out = strings.ReplaceAll(out, "{{"+k+"}}", fmt.Sprint(v))
	}
	return out
}
