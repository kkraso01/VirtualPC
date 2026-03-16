package skills

import (
	"fmt"
	"strings"
)

func Validate(m SkillManifest) error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("skill name is required")
	}
	if strings.TrimSpace(m.Description) == "" {
		return fmt.Errorf("skill %s missing description", m.Name)
	}
	if len(m.Tools) == 0 {
		return fmt.Errorf("skill %s must declare at least one tool", m.Name)
	}
	for _, t := range m.Tools {
		if strings.TrimSpace(t.Name) == "" {
			return fmt.Errorf("skill %s has empty tool binding", m.Name)
		}
	}
	return nil
}
