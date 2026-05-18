package worker_test

import (
	"disturbedcli/internal/splitter"
	"disturbedcli/internal/worker"
	"testing"
)

func mustCompile(t *testing.T, pattern string) interface{ FindStringIndex(string) []int } {
	t.Helper()
	re, err := worker.CompilePattern(pattern, false)
	if err != nil {
		t.Fatalf("compile %q: %v", pattern, err)
	}
	return re
}

func TestRun_BasicMatch(t *testing.T) {
	re, _ := worker.CompilePattern("error", false)
	chunk := splitter.Chunk{
		StartLine: 1,
		Lines:     []string{"no match", "error here", "another error", "clean"},
	}
	result := worker.Run(chunk, re, worker.Options{})
	if result.Count != 2 {
		t.Errorf("expected 2 matches, got %d", result.Count)
	}
	if result.Matches[0].LineNum != 2 || result.Matches[1].LineNum != 3 {
		t.Errorf("wrong line numbers: %v", result.Matches)
	}
}

func TestRun_IgnoreCase(t *testing.T) {
	re, _ := worker.CompilePattern("error", true)
	chunk := splitter.Chunk{
		StartLine: 1,
		Lines:     []string{"ERROR found", "no match"},
	}
	result := worker.Run(chunk, re, worker.Options{IgnoreCase: true})
	if result.Count != 1 {
		t.Errorf("expected 1 match (case-insensitive), got %d", result.Count)
	}
}

func TestRun_InvertMatch(t *testing.T) {
	re, _ := worker.CompilePattern("error", false)
	chunk := splitter.Chunk{
		StartLine: 1,
		Lines:     []string{"ok line", "error line", "another ok"},
	}
	result := worker.Run(chunk, re, worker.Options{InvertMatch: true})
	if result.Count != 2 {
		t.Errorf("expected 2 inverted matches, got %d", result.Count)
	}
}

func TestRun_OnlyMatch(t *testing.T) {
	re, _ := worker.CompilePattern("err[a-z]+", false)
	chunk := splitter.Chunk{
		StartLine: 1,
		Lines:     []string{"fatal error occurred"},
	}
	result := worker.Run(chunk, re, worker.Options{OnlyMatch: true})
	if result.Count != 1 {
		t.Fatalf("expected 1 match")
	}
	if result.Matches[0].Excerpt != "error" {
		t.Errorf("expected excerpt %q, got %q", "error", result.Matches[0].Excerpt)
	}
}

func TestRun_StartLineOffset(t *testing.T) {
	re, _ := worker.CompilePattern("x", false)
	chunk := splitter.Chunk{
		StartLine: 100,
		Lines:     []string{"no", "x here"},
	}
	result := worker.Run(chunk, re, worker.Options{})
	if result.Matches[0].LineNum != 101 {
		t.Errorf("expected LineNum=101, got %d", result.Matches[0].LineNum)
	}
}