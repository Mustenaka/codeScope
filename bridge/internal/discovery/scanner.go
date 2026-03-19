package discovery

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"codescope/bridge/internal/session"
)

type Process struct {
	PID         int
	ParentPID   int
	Name        string
	CommandLine string
}

type Candidate struct {
	Meta        session.Metadata
	PID         int
	ParentPID   int
	ProcessName string
	CommandLine string
	ObservedAt  time.Time
	FallbackWS  bool
}

func (c Candidate) Key() string {
	return c.Meta.SessionID
}

type ProcessLister interface {
	ListProcesses(ctx context.Context) ([]Process, error)
}

type Scanner interface {
	Scan(ctx context.Context) ([]Candidate, error)
}

type ProcessScanner struct {
	lister            ProcessLister
	machineID         string
	fallbackWorkspace string
	currentPID        int
	now               func() time.Time
	stabilityWindow   time.Duration
	remembered        map[string]rememberedSession
}

type rememberedSession struct {
	sessionID string
	expiresAt time.Time
}

func NewProcessScanner(lister ProcessLister, machineID, fallbackWorkspace string) *ProcessScanner {
	if lister == nil {
		lister = systemProcessLister{}
	}
	return &ProcessScanner{
		lister:            lister,
		machineID:         machineID,
		fallbackWorkspace: fallbackWorkspace,
		currentPID:        os.Getpid(),
		now:               time.Now,
		stabilityWindow:   10 * time.Second,
		remembered:        make(map[string]rememberedSession),
	}
}

func (s *ProcessScanner) ConfigureSessionStabilityWindow(window time.Duration) {
	s.stabilityWindow = window
}

func (s *ProcessScanner) Scan(ctx context.Context) ([]Candidate, error) {
	processes, err := s.lister.ListProcesses(ctx)
	if err != nil {
		return nil, err
	}

	processByPID := make(map[int]Process, len(processes))
	for _, process := range processes {
		processByPID[process.PID] = process
	}

	candidates := make([]Candidate, 0, len(processes))
	for _, process := range processes {
		candidate, ok := s.classify(process)
		if !ok {
			continue
		}
		candidates = append(candidates, candidate)
	}
	merged := mergeCandidates(candidates, processByPID)
	s.assignStableSessionIDs(merged)
	return merged, nil
}

func (s *ProcessScanner) classify(process Process) (Candidate, bool) {
	if process.PID <= 0 || process.PID == s.currentPID {
		return Candidate{}, false
	}
	if excludedAgentProcess(process.Name, process.CommandLine) {
		return Candidate{}, false
	}

	agentName := classifyAgent(process.Name, process.CommandLine)
	if agentName == "" {
		return Candidate{}, false
	}

	workspaceRoot, fallbackWS := inferWorkspaceRoot(process.CommandLine, s.fallbackWorkspace)
	workspaceRoot = promoteProjectRoot(workspaceRoot)
	meta := session.Metadata{
		AgentName:     agentName,
		WorkspaceRoot: workspaceRoot,
		MachineID:     s.machineID,
		SessionID:     stableSessionID(agentName, s.machineID, workspaceRoot, process.PID),
	}

	return Candidate{
		Meta:        meta,
		PID:         process.PID,
		ParentPID:   process.ParentPID,
		ProcessName: process.Name,
		CommandLine: process.CommandLine,
		ObservedAt:  s.now(),
		FallbackWS:  fallbackWS,
	}, true
}

func mergeCandidates(candidates []Candidate, processByPID map[int]Process) []Candidate {
	if len(candidates) < 2 {
		return candidates
	}

	candidateByPID := make(map[int]Candidate, len(candidates))
	for _, candidate := range candidates {
		candidateByPID[candidate.PID] = candidate
	}

	merged := make(map[string]Candidate, len(candidates))
	for _, candidate := range candidates {
		canonical := canonicalCandidate(candidate, candidateByPID)
		key := mergedKey(canonical.Meta.AgentName, canonical.Meta.WorkspaceRoot, canonical.PID)
		if existing, ok := merged[key]; ok {
			if existing.ObservedAt.Before(canonical.ObservedAt) {
				merged[key] = canonical
			}
			continue
		}
		if _, ok := merged[key]; !ok {
			merged[key] = canonical
		}
	}

	result := make([]Candidate, 0, len(merged))
	for _, candidate := range merged {
		candidate.Meta.SessionID = stableSessionID(candidate.Meta.AgentName, candidate.Meta.MachineID, candidate.Meta.WorkspaceRoot, candidate.PID)
		result = append(result, candidate)
	}
	return result
}

func canonicalCandidate(candidate Candidate, candidateByPID map[int]Candidate) Candidate {
	current := candidate
	for {
		parent, ok := candidateByPID[current.ParentPID]
		if !ok {
			return current
		}
		if parent.Meta.AgentName != current.Meta.AgentName {
			return current
		}
		if !sameSessionWorkspace(parent, current) {
			return current
		}
		if current.FallbackWS && !parent.FallbackWS {
			current.Meta.WorkspaceRoot = parent.Meta.WorkspaceRoot
			current.FallbackWS = false
		}
		current = parent
	}
}

func excludedAgentProcess(name, commandLine string) bool {
	tokens := commandTokens(name, commandLine)
	for index, token := range tokens {
		if token != "app-server" {
			continue
		}
		if index+1 < len(tokens) && tokens[index+1] == "--analytics-default-enabled" {
			return true
		}
	}
	return false
}

func sameSessionWorkspace(parent, child Candidate) bool {
	if parent.Meta.WorkspaceRoot == child.Meta.WorkspaceRoot {
		return true
	}
	return parent.FallbackWS || child.FallbackWS
}

func mergedKey(agentName, workspaceRoot string, pid int) string {
	return fmt.Sprintf("%s|%s|%d", agentName, workspaceRoot, pid)
}

func (s *ProcessScanner) assignStableSessionIDs(candidates []Candidate) {
	now := s.now()
	for fingerprint, remembered := range s.remembered {
		if !remembered.expiresAt.IsZero() && now.After(remembered.expiresAt) {
			delete(s.remembered, fingerprint)
		}
	}

	seen := make(map[string]struct{}, len(candidates))
	for index := range candidates {
		fingerprint := sessionFingerprint(candidates[index])
		seen[fingerprint] = struct{}{}

		if remembered, ok := s.remembered[fingerprint]; ok {
			candidates[index].Meta.SessionID = remembered.sessionID
		}

		s.remembered[fingerprint] = rememberedSession{
			sessionID: candidates[index].Meta.SessionID,
			expiresAt: now.Add(s.stabilityWindow),
		}
	}

	for fingerprint, remembered := range s.remembered {
		if _, ok := seen[fingerprint]; ok {
			continue
		}
		if remembered.expiresAt.IsZero() {
			remembered.expiresAt = now.Add(s.stabilityWindow)
			s.remembered[fingerprint] = remembered
		}
	}
}

func sessionFingerprint(candidate Candidate) string {
	return fmt.Sprintf("%s|%s|%s", candidate.Meta.AgentName, candidate.Meta.WorkspaceRoot, normalizeCommandIdentity(candidate.ProcessName, candidate.CommandLine))
}

func normalizeCommandIdentity(processName, commandLine string) string {
	parts := commandTokens(processName, commandLine)
	normalized := make([]string, 0, len(parts))
	skipNext := false
	for _, part := range parts {
		if skipNext {
			skipNext = false
			continue
		}
		switch part {
		case "--resume", "--session", "--session-id":
			skipNext = true
			continue
		}
		normalized = append(normalized, part)
	}
	return strings.Join(normalized, "|")
}

func classifyAgent(name, commandLine string) string {
	if matchesAgentExecutable(name, commandLine, "codex") {
		return "codex"
	}
	if matchesAgentExecutable(name, commandLine, "claude") {
		return "claude"
	}
	return ""
}

func matchesAgentExecutable(name, commandLine, agent string) bool {
	tokens := commandTokens(name, commandLine)
	for index, token := range tokens {
		afterNode := index > 0 && isNodeLauncher(tokens[index-1])
		if matchesAgentToken(token, agent, index == 0, afterNode) {
			return true
		}
	}
	return false
}

func commandTokens(name, commandLine string) []string {
	tokens := make([]string, 0, 8)
	if cleaned := cleanToken(name); cleaned != "" {
		tokens = append(tokens, cleaned)
	}
	for _, part := range strings.Fields(commandLine) {
		if cleaned := cleanToken(part); cleaned != "" {
			tokens = append(tokens, cleaned)
		}
	}
	return tokens
}

func cleanToken(token string) string {
	token = strings.TrimSpace(token)
	token = strings.Trim(token, "\"'")
	return strings.ToLower(token)
}

func matchesAgentToken(token, agent string, first bool, afterNode bool) bool {
	base := strings.ToLower(filepath.Base(token))
	hasPath := strings.Contains(token, "/") || strings.Contains(token, `\`)
	switch agent {
	case "codex":
		switch {
		case token == "codex" && (first || afterNode):
			return true
		case token == "codex.exe":
			return true
		case hasPath && (base == "codex.exe" || base == "codex.js"):
			return true
		default:
			return false
		}
	case "claude":
		switch {
		case token == "claude" && (first || afterNode):
			return true
		case token == "claude-code" && (first || afterNode):
			return true
		case token == "claude.exe" || token == "claude-code.exe":
			return true
		case hasPath && (base == "claude.exe" || base == "claude-code.exe"):
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func isNodeLauncher(token string) bool {
	base := strings.ToLower(filepath.Base(token))
	return base == "node" || base == "node.exe"
}

func inferWorkspaceRoot(commandLine, fallback string) (string, bool) {
	tokens := strings.Fields(commandLine)
	for i := 0; i < len(tokens)-1; i++ {
		switch tokens[i] {
		case "--cwd", "-C", "--workspace", "--workspace-root":
			return normalizeWorkspace(tokens[i+1], fallback), false
		}
	}
	return normalizeWorkspace("", fallback), true
}

func normalizeWorkspace(candidate, fallback string) string {
	value := strings.Trim(candidate, "\"'")
	if value == "" {
		value = fallback
	}
	if value == "" {
		return "."
	}
	clean := filepath.Clean(value)
	if abs, err := filepath.Abs(clean); err == nil {
		return filepath.ToSlash(abs)
	}
	return filepath.ToSlash(clean)
}

func promoteProjectRoot(workspaceRoot string) string {
	if workspaceRoot == "" || workspaceRoot == "." {
		return workspaceRoot
	}

	current := workspaceRoot
	bestPath := workspaceRoot
	bestScore := workspaceMarkerScore(workspaceRoot)

	for {
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		score := workspaceMarkerScore(parent)
		if score > bestScore {
			bestPath = parent
			bestScore = score
		}
		current = parent
	}

	return filepath.ToSlash(bestPath)
}

func workspaceMarkerScore(path string) int {
	if path == "" || path == "." {
		return 0
	}

	type marker struct {
		name  string
		score int
	}

	markers := []marker{
		{name: ".git", score: 100},
		{name: "go.work", score: 90},
		{name: "pnpm-workspace.yaml", score: 90},
		{name: "turbo.json", score: 85},
		{name: "nx.json", score: 85},
		{name: "Cargo.toml", score: 60},
		{name: "package.json", score: 60},
		{name: "pyproject.toml", score: 60},
		{name: "pubspec.yaml", score: 60},
		{name: "go.mod", score: 55},
	}

	best := 0
	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(path, marker.name)); err == nil && marker.score > best {
			best = marker.score
		}
	}
	return best
}

func stableSessionID(agentName, machineID, workspaceRoot string, pid int) string {
	sum := sha1.Sum([]byte(fmt.Sprintf("%s|%s|%s|%d", agentName, machineID, workspaceRoot, pid)))
	return "session-" + hex.EncodeToString(sum[:8])
}

type systemProcessLister struct{}

func (systemProcessLister) ListProcesses(ctx context.Context) ([]Process, error) {
	if runtime.GOOS == "windows" {
		return listWindowsProcesses(ctx)
	}
	return listPOSIXProcesses(ctx)
}

func listPOSIXProcesses(ctx context.Context) ([]Process, error) {
	cmd := exec.CommandContext(ctx, "ps", "-eo", "pid=,ppid=,comm=,args=")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list processes with ps: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	processes := make([]Process, 0, len(lines))
	for _, line := range lines {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 3 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		parentPID, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		name := fields[2]
		commandLine := strings.Join(fields[3:], " ")
		if commandLine == "" {
			commandLine = name
		}
		processes = append(processes, Process{
			PID:         pid,
			ParentPID:   parentPID,
			Name:        name,
			CommandLine: commandLine,
		})
	}
	return processes, nil
}

func listWindowsProcesses(ctx context.Context) ([]Process, error) {
	command := "Get-CimInstance Win32_Process | Select-Object ProcessId,ParentProcessId,Name,CommandLine | ConvertTo-Json -Compress"
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", command)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list processes with powershell: %w", err)
	}
	return parseWindowsProcessJSON(output)
}

func parseWindowsProcessJSON(data []byte) ([]Process, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		return nil, nil
	}

	type record struct {
		ProcessID       int    `json:"ProcessId"`
		ParentProcessID int    `json:"ParentProcessId"`
		Name            string `json:"Name"`
		CommandLine     string `json:"CommandLine"`
	}

	if strings.HasPrefix(trimmed, "{") {
		var single record
		if err := json.Unmarshal([]byte(trimmed), &single); err != nil {
			return nil, fmt.Errorf("decode process list: %w", err)
		}
		return []Process{{
			PID:         single.ProcessID,
			ParentPID:   single.ParentProcessID,
			Name:        single.Name,
			CommandLine: single.CommandLine,
		}}, nil
	}

	var raw []record
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, fmt.Errorf("decode process list: %w", err)
	}

	processes := make([]Process, 0, len(raw))
	for _, item := range raw {
		processes = append(processes, Process{
			PID:         item.ProcessID,
			ParentPID:   item.ParentProcessID,
			Name:        item.Name,
			CommandLine: item.CommandLine,
		})
	}
	return processes, nil
}
