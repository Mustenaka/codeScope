package discovery

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fakeProcessLister struct {
	processes []Process
	err       error
}

func (l fakeProcessLister) ListProcesses(context.Context) ([]Process, error) {
	return l.processes, l.err
}

func TestProcessScannerFindsSupportedAgents(t *testing.T) {
	scanner := NewProcessScanner(fakeProcessLister{
		processes: []Process{
			{PID: 100, Name: "bridge.exe", CommandLine: "bridge"},
			{PID: 200, Name: "codex.exe", CommandLine: "codex --cwd D:/repo-one"},
			{PID: 300, Name: "node.exe", CommandLine: "claude-code --cwd D:/repo-two"},
			{PID: 400, Name: "powershell.exe", CommandLine: "powershell"},
		},
	}, "machine-dev", "D:/fallback")
	scanner.currentPID = 100

	candidates, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}

	byAgent := make(map[string]Candidate, len(candidates))
	for _, candidate := range candidates {
		byAgent[candidate.Meta.AgentName] = candidate
	}
	if byAgent["codex"].Meta.AgentName != "codex" {
		t.Fatalf("expected codex candidate, got %#v", candidates)
	}
	if byAgent["codex"].Meta.WorkspaceRoot != "D:/repo-one" {
		t.Fatalf("expected parsed codex workspace root, got %q", byAgent["codex"].Meta.WorkspaceRoot)
	}
	if byAgent["claude"].Meta.AgentName != "claude" {
		t.Fatalf("expected claude candidate, got %#v", candidates)
	}
	if byAgent["claude"].Meta.WorkspaceRoot != "D:/repo-two" {
		t.Fatalf("expected parsed claude workspace root, got %q", byAgent["claude"].Meta.WorkspaceRoot)
	}
	if byAgent["codex"].Meta.SessionID == "" || byAgent["claude"].Meta.SessionID == "" {
		t.Fatal("expected generated session ids")
	}
}

func TestProcessScannerFallsBackToConfiguredWorkspace(t *testing.T) {
	scanner := NewProcessScanner(fakeProcessLister{
		processes: []Process{
			{PID: 500, Name: "codex", CommandLine: "codex run"},
		},
	}, "machine-dev", "D:/fallback")

	candidates, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0].Meta.WorkspaceRoot != "D:/fallback" {
		t.Fatalf("expected fallback workspace root, got %q", candidates[0].Meta.WorkspaceRoot)
	}
}

func TestProcessScannerIgnoresWrapperAndNoiseProcesses(t *testing.T) {
	scanner := NewProcessScanner(fakeProcessLister{
		processes: []Process{
			{PID: 100, Name: "cmd.exe", CommandLine: `cmd.exe /c "C:\nvm4w\nodejs\npx.cmd -y @modelcontextprotocol/server-github"`},
			{PID: 200, Name: "node.exe", CommandLine: `"C:\nvm4w\nodejs\node.exe" "C:\nvm4w\nodejs\node_modules\npm\bin\npx-cli.js" -y @openai/codex`},
			{PID: 300, Name: "node.exe", CommandLine: `"C:\nvm4w\nodejs\node.exe" "C:\nvm4w\nodejs\node_modules\@openai\codex\bin\codex.js" cli --cwd D:/repo-one`},
			{PID: 400, Name: "powershell.exe", CommandLine: `powershell -Command "Write-Host claude"`},
			{PID: 500, Name: "claude.exe", CommandLine: `claude.exe --cwd D:/repo-two`},
		},
	}, "machine-dev", "D:/fallback")

	candidates, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(candidates) != 2 {
		t.Fatalf("expected 2 true cli candidates, got %d (%#v)", len(candidates), candidates)
	}

	if candidates[0].PID != 300 {
		t.Fatalf("expected codex wrapper entry process pid 300, got %d", candidates[0].PID)
	}
	if candidates[1].PID != 500 {
		t.Fatalf("expected claude binary pid 500, got %d", candidates[1].PID)
	}
}

func TestProcessScannerIgnoresCodexDesktopHelperProcess(t *testing.T) {
	scanner := NewProcessScanner(fakeProcessLister{
		processes: []Process{
			{
				PID:         100,
				Name:        "codex.exe",
				CommandLine: `C:\Users\Mumte\AppData\Local\codex\codex.exe app-server --analytics-default-enabled`,
			},
			{
				PID:         200,
				Name:        "codex.exe",
				CommandLine: `C:\Users\Mumte\AppData\Local\codex\codex.exe cli --cwd D:/repo-one`,
			},
		},
	}, "machine-dev", "D:/fallback")

	candidates, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(candidates) != 1 {
		t.Fatalf("expected only real cli candidate, got %d (%#v)", len(candidates), candidates)
	}
	if candidates[0].PID != 200 {
		t.Fatalf("expected cli pid 200, got %d", candidates[0].PID)
	}
}

func TestProcessScannerCollapsesParentChildCandidatesIntoSingleSession(t *testing.T) {
	scanner := NewProcessScanner(fakeProcessLister{
		processes: []Process{
			{PID: 100, ParentPID: 10, Name: "node.exe", CommandLine: `"C:\nvm4w\nodejs\node.exe" "C:\nvm4w\nodejs\node_modules\@openai\codex\bin\codex.js" cli --cwd D:/repo-one`},
			{PID: 110, ParentPID: 100, Name: "codex.exe", CommandLine: `C:\Users\Mumte\AppData\Local\codex\codex.exe cli`},
			{PID: 120, ParentPID: 110, Name: "cmd.exe", CommandLine: `cmd.exe /c "echo helper"`},
		},
	}, "machine-dev", "D:/fallback")

	candidates, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(candidates) != 1 {
		t.Fatalf("expected merged single candidate, got %d (%#v)", len(candidates), candidates)
	}

	if candidates[0].PID != 100 {
		t.Fatalf("expected canonical root pid 100, got %d", candidates[0].PID)
	}
	if candidates[0].Meta.WorkspaceRoot != "D:/repo-one" {
		t.Fatalf("expected workspace from merged chain, got %q", candidates[0].Meta.WorkspaceRoot)
	}
}

func TestProcessScannerDoesNotMergeDifferentWorkspaces(t *testing.T) {
	scanner := NewProcessScanner(fakeProcessLister{
		processes: []Process{
			{PID: 100, ParentPID: 10, Name: "node.exe", CommandLine: `"C:\nvm4w\nodejs\node.exe" "C:\nvm4w\nodejs\node_modules\@openai\codex\bin\codex.js" cli --cwd D:/repo-one`},
			{PID: 110, ParentPID: 100, Name: "codex.exe", CommandLine: `C:\Users\Mumte\AppData\Local\codex\codex.exe cli --cwd D:/repo-two`},
		},
	}, "machine-dev", "D:/fallback")

	candidates, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(candidates) != 2 {
		t.Fatalf("expected separate candidates for different workspaces, got %d", len(candidates))
	}
}

func TestPromoteProjectRootPrefersGoWorkParentOverLeafModule(t *testing.T) {
	root := t.TempDir()
	projectRoot := filepath.Join(root, "codeScope")
	leafRoot := filepath.Join(projectRoot, "bridge")
	if err := os.MkdirAll(leafRoot, 0o755); err != nil {
		t.Fatalf("mkdir leaf root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "go.work"), []byte("go 1.25"), 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}
	if err := os.WriteFile(filepath.Join(leafRoot, "go.mod"), []byte("module bridge"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	got := promoteProjectRoot(filepath.ToSlash(leafRoot))
	want := filepath.ToSlash(projectRoot)
	if got != want {
		t.Fatalf("expected project root %q, got %q", want, got)
	}
}

func TestPromoteProjectRootKeepsLeafModuleWithoutStrongerParentMarker(t *testing.T) {
	root := t.TempDir()
	leafRoot := filepath.Join(root, "bridge")
	if err := os.MkdirAll(leafRoot, 0o755); err != nil {
		t.Fatalf("mkdir leaf root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(leafRoot, "go.mod"), []byte("module bridge"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	got := promoteProjectRoot(filepath.ToSlash(leafRoot))
	want := filepath.ToSlash(leafRoot)
	if got != want {
		t.Fatalf("expected leaf project root %q, got %q", want, got)
	}
}

type scriptedProcessLister struct {
	snapshots [][]Process
	index     int
}

func (l *scriptedProcessLister) ListProcesses(context.Context) ([]Process, error) {
	if l.index >= len(l.snapshots) {
		return l.snapshots[len(l.snapshots)-1], nil
	}
	current := l.snapshots[l.index]
	l.index++
	return current, nil
}

func TestProcessScannerReusesSessionIDWithinStabilityWindow(t *testing.T) {
	lister := &scriptedProcessLister{
		snapshots: [][]Process{
			{{PID: 100, Name: "codex.exe", CommandLine: `codex.exe --cwd D:/repo-one`}},
			{},
			{{PID: 200, Name: "codex.exe", CommandLine: `codex.exe --cwd D:/repo-one`}},
		},
	}
	scanner := NewProcessScanner(lister, "machine-dev", "D:/fallback")
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	scanner.now = func() time.Time { return now }
	scanner.ConfigureSessionStabilityWindow(10 * time.Second)

	first, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("first scan: %v", err)
	}
	now = now.Add(5 * time.Second)
	if _, err := scanner.Scan(context.Background()); err != nil {
		t.Fatalf("second scan: %v", err)
	}
	now = now.Add(2 * time.Second)
	third, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("third scan: %v", err)
	}

	if len(first) != 1 || len(third) != 1 {
		t.Fatalf("expected one candidate before/after restart, got %d and %d", len(first), len(third))
	}
	if first[0].Meta.SessionID != third[0].Meta.SessionID {
		t.Fatalf("expected session id reuse within window, got %q vs %q", first[0].Meta.SessionID, third[0].Meta.SessionID)
	}
}

func TestProcessScannerAllocatesNewSessionIDAfterStabilityWindowExpires(t *testing.T) {
	lister := &scriptedProcessLister{
		snapshots: [][]Process{
			{{PID: 100, Name: "codex.exe", CommandLine: `codex.exe --cwd D:/repo-one`}},
			{},
			{{PID: 200, Name: "codex.exe", CommandLine: `codex.exe --cwd D:/repo-one`}},
		},
	}
	scanner := NewProcessScanner(lister, "machine-dev", "D:/fallback")
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	scanner.now = func() time.Time { return now }
	scanner.ConfigureSessionStabilityWindow(3 * time.Second)

	first, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("first scan: %v", err)
	}
	now = now.Add(4 * time.Second)
	if _, err := scanner.Scan(context.Background()); err != nil {
		t.Fatalf("second scan: %v", err)
	}
	now = now.Add(4 * time.Second)
	third, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatalf("third scan: %v", err)
	}

	if len(first) != 1 || len(third) != 1 {
		t.Fatalf("expected one candidate before/after restart, got %d and %d", len(first), len(third))
	}
	if first[0].Meta.SessionID == third[0].Meta.SessionID {
		t.Fatalf("expected new session id after window expiry, got same %q", third[0].Meta.SessionID)
	}
}
