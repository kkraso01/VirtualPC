package safety

import "strings"

type Decision string

const (
	DecisionAllow   Decision = "allow"
	DecisionBlock   Decision = "block"
	DecisionApprove Decision = "require_approval"
)

type CommandPolicy struct {
	AllowPrefixes []string
	DenyContains  []string
}

func DefaultCommandPolicy() CommandPolicy {
	return CommandPolicy{
		AllowPrefixes: []string{"ls", "pwd", "cat", "echo", "grep", "sed", "awk", "go", "python", "npm", "make", "git", "find", "cp", "mv", "mkdir", "touch", "service", "systemctl"},
		DenyContains:  []string{"rm -rf /", "mkfs", "shutdown", "reboot", " mount ", "chmod 777 /", "dd if=/dev/zero", "kill -9 1", "nmap", "masscan"},
	}
}

func (p CommandPolicy) Evaluate(command string, dangerousMode string) (Decision, string) {
	norm := " " + strings.ToLower(strings.TrimSpace(command)) + " "
	for _, d := range p.DenyContains {
		if strings.Contains(norm, strings.ToLower(d)) {
			if dangerousMode == "approve" {
				return DecisionApprove, "dangerous command requires approval"
			}
			return DecisionBlock, "command matches denylist"
		}
	}
	for _, a := range p.AllowPrefixes {
		if strings.HasPrefix(strings.TrimSpace(strings.ToLower(command)), strings.ToLower(a)+" ") || strings.EqualFold(strings.TrimSpace(command), a) {
			return DecisionAllow, ""
		}
	}
	if dangerousMode == "approve" {
		return DecisionApprove, "command is outside allowlist"
	}
	return DecisionAllow, "outside allowlist but not denylisted"
}
