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
