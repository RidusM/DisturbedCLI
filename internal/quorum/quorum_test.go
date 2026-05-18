package quorum_test

import (
	"disturbedcli/internal/quorum"
	"disturbedcli/internal/worker"
	"errors"
	"testing"
)

func makeChannel(results []worker.Result) <-chan worker.Result {
	ch := make(chan worker.Result, len(results))
	for _, r := range results {
		ch <- r
	}
	return ch
}

func TestCollect_AllSucceed(t *testing.T) {
	results := []worker.Result{
		{Matches: []worker.Match{{LineNum: 3, Line: "c"}}, Count: 1},
		{Matches: []worker.Match{{LineNum: 1, Line: "a"}}, Count: 1},
		{Matches: []worker.Match{{LineNum: 2, Line: "b"}}, Count: 1},
	}
	ch := makeChannel(results)
	matches, err := quorum.Collect(ch, 3, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 3 {
		t.Errorf("expected 3 matches, got %d", len(matches))
	}

	if matches[0].LineNum != 1 || matches[1].LineNum != 2 || matches[2].LineNum != 3 {
		t.Errorf("not sorted: %v", matches)
	}
}

func TestCollect_QuorumReached(t *testing.T) {
	results := []worker.Result{
		{Matches: []worker.Match{{LineNum: 1}}, Count: 1},
		{Err: errors.New("boom")},
		{Matches: []worker.Match{{LineNum: 2}}, Count: 1},
	}
	ch := makeChannel(results)
	_, err := quorum.Collect(ch, 3, 2)
	if err != nil {
		t.Errorf("quorum should be reached (2/3): %v", err)
	}
}

func TestCollect_QuorumNotReached(t *testing.T) {
	results := []worker.Result{
		{Err: errors.New("fail1")},
		{Err: errors.New("fail2")},
		{Matches: []worker.Match{{LineNum: 1}}, Count: 1},
	}
	ch := makeChannel(results)
	_, err := quorum.Collect(ch, 3, 3)
	if err == nil {
		t.Fatal("expected quorum error")
	}
	if !errors.Is(err, quorum.ErrQuorumNotReached) {
		t.Errorf("wrong error type: %v", err)
	}
}