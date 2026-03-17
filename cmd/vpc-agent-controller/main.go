package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"
	"time"

	"virtualpc/agent/capabilities"
	"virtualpc/agent/config"
	"virtualpc/agent/controller"
	"virtualpc/agent/mcp"
	"virtualpc/agent/providers"
	"virtualpc/agent/skills"
	"virtualpc/agent/tools"
	"virtualpc/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "start":
		start(os.Args[2:])
	case "attach":
		attach(os.Args[2:])
	case "logs":
		logs(os.Args[2:])
	case "stop":
		stop(os.Args[2:])
	case "list":
		listSessions()
	case "schemas":
		printJSON(tools.Catalog())
	case "skill":
		handleSkill(os.Args[2:])
	case "tool":
		handleTool(os.Args[2:])
	case "provider":
		handleProvider(os.Args[2:])
	case "mcp":
		handleMCP(os.Args[2:])
	case "approvals":
		handleApprovals(os.Args[2:])
	case "approve":
		approveDecision(os.Args[2:], true)
	case "deny":
		approveDecision(os.Args[2:], false)
	default:
		usage()
		os.Exit(1)
	}
}

func start(args []string) {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	machineID := fs.String("machine", "", "machine id")
	goal := fs.String("goal", "", "task goal")
	sock := fs.String("socket", env("VPC_UNIX_SOCKET", "/tmp/virtualpcd.sock"), "vpc unix socket")
	cfgPath := fs.String("config", "agent/config/agent_config.yaml", "agent config file")
	providerName := fs.String("provider", "", "openai|anthropic|openai_compatible|ollama|vllm")
	providerProfile := fs.String("provider-profile", "", "named provider profile")
	skillFlags := multiFlag{}
	mcpFlags := multiFlag{}
	fs.Var(&skillFlags, "skill", "attach skill pack")
	fs.Var(&mcpFlags, "mcp", "activate mcp endpoint")
	approval := fs.Bool("approval", false, "require operator approval for dangerous operations")
	_ = fs.Parse(args)
	if *machineID == "" || *goal == "" {
		fmt.Fprintln(os.Stderr, "--machine and --goal are required")
		os.Exit(2)
	}
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		cfg = config.Default()
	}
	registry, skillManifests, profiles, mcpServers, err := capabilities.Load(context.Background(), capabilities.LoaderOptions{SkillsRoot: "skills", LocalToolsRoot: "tools/local", HTTPToolsRoot: "tools/http", ProviderProfilesRoot: "agent/providers/profiles", MCPConfigPath: "mcp/servers.yaml", MCPClient: mcp.NoopClient{}})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *providerProfile != "" {
		p, ok := findProfile(profiles, *providerProfile)
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown provider profile: %s\n", *providerProfile)
			os.Exit(2)
		}
		cfg.Provider = p.Provider
		cfg.Model = p.Model
		cfg.BaseURL = p.BaseURL
		if p.APIKeyEnv != "" {
			cfg.APIKey = os.Getenv(p.APIKeyEnv)
		}
		cfg.SupportsChatCompletions = p.SupportsChatCompletions
		cfg.SupportsToolCalling = p.SupportsToolCalling
		cfg.SupportsResponsesAPI = p.SupportsResponsesAPI
		cfg.SupportsStatefulResponses = p.SupportsStatefulResponses
		if !cfg.SupportsToolCalling {
			fmt.Fprintf(os.Stderr, "provider profile %s does not support tool calling\n", p.Name)
			os.Exit(2)
		}
	}
	if *providerName != "" {
		cfg.Provider = *providerName
	}
	capabilities.ApplyPolicyBindings(&cfg, registry.All())
	session := &controller.Session{SessionID: fmt.Sprintf("s-%d", time.Now().Unix()), MachineID: *machineID, Goal: *goal, Status: "created", Provider: cfg.Provider, Model: cfg.Model, AttachedSkills: skillFlags, ActiveProviderProfile: *providerProfile, ActiveMCPServers: mcpFlags, PolicyOverrides: map[string]any{}}
	resolveSessionState(session, registry, skillManifests, mcpServers)
	if err := session.Save(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctr, err := controller.New(cfg, cli.New(*sock), chooseProvider(cfg), "agent/prompts/system_prompt.txt", *approval, controller.SessionLogPath(session.SessionID), registry, mcpServers)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer ctr.Close()
	if err := ctr.Run(session); err != nil {
		fmt.Fprintf(os.Stderr, "session stopped: %v\n", err)
	}
	_ = session.Save()
	printJSON(map[string]any{"session_id": session.SessionID, "status": session.Status, "log_path": controller.SessionLogPath(session.SessionID), "skills": session.AttachedSkills, "provider_profile": session.ActiveProviderProfile, "mcp": session.ActiveMCPServers})
}

func resolveSessionState(s *controller.Session, reg *capabilities.Registry, manifests []skills.SkillManifest, servers []mcp.ServerConfig) {
	rt := skills.BuildRuntime("", s.AttachedSkills, manifests)
	for _, c := range reg.All() {
		if c.Type != capabilities.TypeTool || !c.Enabled {
			continue
		}
		if len(rt.Overlay.ToolAllow) > 0 {
			if v, ok := rt.Overlay.ToolAllow[c.Name]; ok {
				if !v {
					continue
				}
			} else if c.Source == capabilities.SourceBuiltin {
				continue
			}
		}
		s.ResolvedTools = append(s.ResolvedTools, c.Name)
	}
	s.PolicyOverrides["approval_mode"] = rt.Overlay.Policy.ApprovalMode
	s.PolicyOverrides["allowed_network"] = rt.Overlay.Policy.AllowedNetwork
	for _, sk := range s.AttachedSkills {
		if !skillExists(manifests, sk) {
			s.PolicyOverrides["unknown_skill."+sk] = true
		}
	}
	for _, m := range s.ActiveMCPServers {
		if !mcpExists(servers, m) {
			s.PolicyOverrides["unknown_mcp."+m] = true
		}
	}
}

func chooseProvider(cfg config.Config) providers.Provider {
	switch cfg.Provider {
	case "anthropic":
		return providers.NewAnthropic(cfg.Model)
	case "openai", "":
		return providers.NewOpenAI(cfg.Model)
	case "openai_compatible", "ollama", "vllm":
		return providers.NewOpenAICompatible(cfg.Model, cfg.BaseURL, cfg.APIKey, cfg.ProviderCapabilities())
	default:
		return nil
	}
}

func handleSkill(args []string) {
	manifests, err := skills.LoadAll("skills")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(args) == 0 || args[0] == "list" {
		rows := [][]string{{"NAME", "DESCRIPTION", "SOURCE", "ENABLED"}}
		for _, m := range manifests {
			rows = append(rows, []string{m.Name, m.Description, "skills/" + m.Name, "true"})
		}
		printTable(rows)
		return
	}
	if args[0] == "inspect" && len(args) > 1 {
		for _, m := range manifests {
			if m.Name == args[1] {
				printJSON(m)
				return
			}
		}
	}
	fmt.Fprintln(os.Stderr, "usage: vpc-agent-controller skill list|inspect <name>")
	os.Exit(2)
}

func handleTool(args []string) {
	reg, _, _, _, _ := capabilities.Load(context.Background(), capabilities.LoaderOptions{SkillsRoot: "skills", LocalToolsRoot: "tools/local", HTTPToolsRoot: "tools/http", ProviderProfilesRoot: "agent/providers/profiles", MCPConfigPath: "mcp/servers.yaml", MCPClient: mcp.NoopClient{}})
	toolsCaps := []capabilities.Capability{}
	for _, c := range reg.All() {
		if c.Type == capabilities.TypeTool {
			toolsCaps = append(toolsCaps, c)
		}
	}
	if len(args) == 0 || args[0] == "list" {
		rows := [][]string{{"NAME", "SOURCE", "TYPE", "EXECUTION", "ENABLED", "APPROVAL", "NETWORK"}}
		slices.SortFunc(toolsCaps, func(a, b capabilities.Capability) int { return strings.Compare(a.Name, b.Name) })
		for _, c := range toolsCaps {
			rows = append(rows, []string{c.Name, string(c.Source), string(c.Type), string(c.ExecutionLocation), fmt.Sprintf("%t", c.Enabled), fmt.Sprintf("%t", c.ApprovalRequired || len(c.Policy.ApprovalsRequired) > 0), fmt.Sprintf("%t", c.NetworkRequired)})
		}
		printTable(rows)
		return
	}
	if args[0] == "inspect" && len(args) > 1 {
		for _, c := range toolsCaps {
			if c.Name == args[1] {
				printJSON(c)
				return
			}
		}
	}
	fmt.Fprintln(os.Stderr, "usage: vpc-agent-controller tool list|inspect <name>")
	os.Exit(2)
}

func handleProvider(args []string) {
	_, _, profiles, _, _ := capabilities.Load(context.Background(), capabilities.LoaderOptions{ProviderProfilesRoot: "agent/providers/profiles"})
	if len(args) == 0 || args[0] == "list" {
		rows := [][]string{{"NAME", "PROVIDER", "MODEL", "SOURCE", "TOOLS", "RESPONSES_API", "STATEFUL"}}
		for _, p := range profiles {
			rows = append(rows, []string{p.Name, p.Provider, p.Model, "provider-profile", fmt.Sprintf("%t", p.SupportsToolCalling), fmt.Sprintf("%t", p.SupportsResponsesAPI), fmt.Sprintf("%t", p.SupportsStatefulResponses)})
		}
		printTable(rows)
		return
	}
	if args[0] == "inspect" && len(args) > 1 {
		for _, p := range profiles {
			if p.Name == args[1] {
				printJSON(p)
				return
			}
		}
	}
	fmt.Fprintln(os.Stderr, "usage: vpc-agent-controller provider list|inspect <name>")
	os.Exit(2)
}

func handleMCP(args []string) {
	servers, err := mcp.LoadConfig("mcp/servers.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(args) == 0 || args[0] == "list" {
		rows := [][]string{{"NAME", "SOURCE", "TYPE", "EXECUTION", "ENABLED"}}
		for _, s := range servers {
			loc := "sidecar"
			if s.Mode == "remote" {
				loc = "remote"
			}
			rows = append(rows, []string{s.Name, "mcp", s.Mode, loc, "true"})
		}
		printTable(rows)
		return
	}
	if args[0] == "inspect" && len(args) > 1 {
		for _, s := range servers {
			if s.Name == args[1] {
				printJSON(s)
				return
			}
		}
	}
	fmt.Fprintln(os.Stderr, "usage: vpc-agent-controller mcp list|inspect <name>")
	os.Exit(2)
}

func attach(args []string) {
	s, err := controller.LoadSession(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	printJSON(s)
}
func logs(args []string) {
	b, err := os.ReadFile(controller.SessionLogPath(args[0]))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(string(b))
}
func stop(args []string) {
	s, err := controller.LoadSession(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	s.StopRequested = true
	s.Status = "stop_requested"
	_ = s.Save()
	printJSON(map[string]any{"session_id": s.SessionID, "status": s.Status})
}
func listSessions() {
	sessions, err := controller.ListSessions()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	printJSON(sessions)
}

func handleApprovals(args []string) {
	sessionID := ""
	if len(args) > 0 {
		sessionID = args[0]
	}
	items, err := controller.ListApprovals(sessionID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	rows := [][]string{{"APPROVAL_ID", "SESSION", "CAPABILITY", "STATUS", "REASON", "UPDATED"}}
	for _, a := range items {
		rows = append(rows, []string{a.ID, a.SessionID, a.CapabilityID, string(a.Status), a.Reason, a.UpdatedAt.Format(time.RFC3339)})
	}
	printTable(rows)
}

func approveDecision(args []string, approve bool) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: approve|deny <session-id> <approval-id>")
		os.Exit(2)
	}
	if err := controller.ResolveApproval(args[0], args[1], approve); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	action := "denied"
	if approve {
		action = "approved"
	}
	printTable([][]string{{"SESSION", "APPROVAL_ID", "ACTION"}, {args[0], args[1], action}})
}

func usage() {
	fmt.Println("vpc-agent-controller start --machine <id> --goal <text> [--skill <name>] [--provider-profile <name>] [--mcp <name>]")
	fmt.Println("vpc-agent-controller attach <session-id>")
	fmt.Println("vpc-agent-controller logs <session-id>")
	fmt.Println("vpc-agent-controller stop <session-id>")
	fmt.Println("vpc-agent-controller list")
	fmt.Println("vpc-agent-controller schemas")
	fmt.Println("vpc-agent-controller skill list|inspect <name>")
	fmt.Println("vpc-agent-controller tool list|inspect <name>")
	fmt.Println("vpc-agent-controller provider list|inspect <name>")
	fmt.Println("vpc-agent-controller mcp list|inspect <name>")
	fmt.Println("vpc-agent-controller approvals [session-id]")
	fmt.Println("vpc-agent-controller approve <session-id> <approval-id>")
	fmt.Println("vpc-agent-controller deny <session-id> <approval-id>")
}
func env(k, f string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return f
}
func printJSON(v any) { b, _ := json.MarshalIndent(v, "", "  "); fmt.Println(string(b)) }

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func findProfile(profiles []capabilities.ProviderProfile, name string) (capabilities.ProviderProfile, bool) {
	for _, p := range profiles {
		if p.Name == name {
			return p, true
		}
	}
	return capabilities.ProviderProfile{}, false
}
func skillExists(skills []skills.SkillManifest, name string) bool {
	for _, s := range skills {
		if s.Name == name {
			return true
		}
	}
	return false
}
func mcpExists(s []mcp.ServerConfig, name string) bool {
	for _, e := range s {
		if e.Name == name {
			return true
		}
	}
	return false
}

func printTable(rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	for _, r := range rows {
		fmt.Fprintln(w, strings.Join(r, "\t"))
	}
	_ = w.Flush()
}
