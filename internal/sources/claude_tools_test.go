package sources

import (
	"encoding/json"
	"testing"
)

func TestClaudeCommandsPullsBashInvocations(t *testing.T) {
	raw := json.RawMessage(`[
	  {"type":"text","text":"running the suite"},
	  {"type":"tool_use","name":"Bash","input":{"command":"go test ./internal/index/ -count=1"}},
	  {"type":"tool_use","name":"Read","input":{"file_path":"/tmp/x.go"}},
	  {"type":"tool_use","name":"Bash","input":{"command":""}},
	  {"type":"tool_result","content":"ok  github.com/x 1.2s"}
	]`)
	got := claudeCommands(raw)
	if len(got) != 1 || got[0] != "$ go test ./internal/index/ -count=1" {
		t.Fatalf("commands=%#v", got)
	}
}

// The shapes that must not panic or invent a command.
func TestClaudeCommandsIgnoresEverythingElse(t *testing.T) {
	for _, raw := range []string{`"plain string"`, `null`, ``, `[]`, `[1,2]`, `[{"type":"text","text":"hi"}]`,
		`{"type":"tool_use","name":"Bash","input":{"command":"ls"}}`} {
		if got := claudeCommands(json.RawMessage(raw)); len(got) != 0 {
			t.Fatalf("%s yielded %#v", raw, got)
		}
	}
}

// Tool output has always been indexed under the user role; only the invocation
// was missing. This pins that the text half is untouched by the new path.
func TestClaudeTextStillReadsToolResults(t *testing.T) {
	raw := json.RawMessage(`[{"type":"tool_result","content":"--- FAIL: TestX"}]`)
	if got := claudeText(raw); got != "--- FAIL: TestX" {
		t.Fatalf("text=%q", got)
	}
}

// The selection is the point: three quarters of command text on a real corpus
// is navigation that answers nothing and matches queries by accident —
// `cat internal/index/index.go` looks relevant to a question about the index.
func TestWorthIndexingKeepsWhatHappened(t *testing.T) {
	keep := []string{
		"go test ./internal/index/ -count=1",
		"golangci-lint run",
		"gh pr checks 558",
		"git commit -m 'fix: x'",
		"make build && deja index --rebuild",
		"docker compose up -d",
	}
	for _, c := range keep {
		if !worthIndexing(c) {
			t.Errorf("dropped a command worth keeping: %q", c)
		}
	}
	drop := []string{
		"cat internal/index/index.go",
		"ls -la",
		"grep -rn foo internal/",
		"cd /repo && pwd",
		"sed -n 1,40p internal/index/index.go",
		"echo hi",
	}
	for _, c := range drop {
		if worthIndexing(c) {
			t.Errorf("kept navigation: %q", c)
		}
	}
}
