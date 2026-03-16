package skills

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type skillMeta struct {
	Description string   `yaml:"description"`
	Purpose     string   `yaml:"purpose"`
	Workflows   []string `yaml:"workflows"`
}

type toolsFile struct {
	Tools []ToolBinding `yaml:"tools"`
}

func LoadAll(root string) ([]SkillManifest, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]SkillManifest, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		m, err := LoadOne(filepath.Join(root, e.Name()))
		if err == nil {
			out = append(out, m)
		}
	}
	return out, nil
}

func LoadOne(path string) (SkillManifest, error) {
	name := filepath.Base(path)
	m := SkillManifest{Name: name, Path: path}
	_ = parseYAMLHeader(filepath.Join(path, "SKILL.md"), &skillMeta{Description: name})
	var meta skillMeta
	_ = parseYAMLHeader(filepath.Join(path, "SKILL.md"), &meta)
	m.Description = firstNonEmpty(meta.Description, name+" skill pack")
	m.Purpose = meta.Purpose
	m.Workflows = meta.Workflows
	prompt, _ := os.ReadFile(filepath.Join(path, "prompt.md"))
	m.Prompt = string(prompt)
	var tf toolsFile
	_ = parseYAML(filepath.Join(path, "tools.yaml"), &tf)
	m.Tools = tf.Tools
	_ = parseYAML(filepath.Join(path, "policies.yaml"), &m.Policies)
	resDir := filepath.Join(path, "resources")
	if entries, err := os.ReadDir(resDir); err == nil {
		for _, r := range entries {
			if !r.IsDir() {
				m.Resources = append(m.Resources, filepath.Join(resDir, r.Name()))
			}
		}
	}
	return m, Validate(m)
}

func parseYAML(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, out)
}

func parseYAMLHeader(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	s := strings.TrimSpace(string(b))
	if strings.HasPrefix(s, "---") {
		parts := strings.SplitN(s, "---", 3)
		if len(parts) >= 3 {
			return yaml.Unmarshal([]byte(parts[1]), out)
		}
	}
	return nil
}

func firstNonEmpty(v, f string) string {
	if strings.TrimSpace(v) != "" {
		return v
	}
	return f
}
