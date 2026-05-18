package splitter_test

import (
	"disturbedcli/internal/splitter"
	"testing"
)

func TestSplit_EvenDivision(t *testing.T) {
	lines := []string{"a", "b", "c", "d"}
	chunks := splitter.Split(lines, 2)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if len(chunks[0].Lines) != 2 || len(chunks[1].Lines) != 2 {
		t.Fatalf("expected 2 lines per chunk, got %d and %d",
			len(chunks[0].Lines), len(chunks[1].Lines))
	}
	if chunks[0].StartLine != 1 {
		t.Errorf("expected StartLine=1, got %d", chunks[0].StartLine)
	}
	if chunks[1].StartLine != 3 {
		t.Errorf("expected StartLine=3, got %d", chunks[1].StartLine)
	}
}

func TestSplit_OddDivision(t *testing.T) {
	lines := []string{"a", "b", "c", "d", "e"}
	chunks := splitter.Split(lines, 3)
	total := 0
	for _, ch := range chunks {
		total += len(ch.Lines)
	}
	if total != 5 {
		t.Errorf("expected 5 total lines, got %d", total)
	}
}

func TestSplit_MoreWorkersThanLines(t *testing.T) {
	lines := []string{"only"}
	chunks := splitter.Split(lines, 10)
	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(chunks))
	}
}

func TestSplit_Empty(t *testing.T) {
	chunks := splitter.Split(nil, 4)
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for empty input, got %d", len(chunks))
	}
}