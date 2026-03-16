package skills

import "strings"

func ComposePrompt(base string, fragments []string) string {
	parts := []string{strings.TrimSpace(base)}
	for _, f := range fragments {
		if strings.TrimSpace(f) != "" {
			parts = append(parts, strings.TrimSpace(f))
		}
	}
	return strings.Join(parts, "\n\n")
}
