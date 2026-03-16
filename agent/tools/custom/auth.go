package custom

import "strings"

func RedactHeaders(h map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range h {
		if strings.EqualFold(k, "authorization") || strings.Contains(strings.ToLower(k), "token") {
			out[k] = "<redacted>"
		} else {
			out[k] = v
		}
	}
	return out
}
