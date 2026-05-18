package coordinator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"disturbedcli/internal/splitter"
	"disturbedcli/internal/worker"
)

const _workerHTTPTimeout = 30 * time.Second

type RemoteOptions struct {
	Peers       []string
	Quorum      int
	Pattern     string
	IgnoreCase  bool
	InvertMatch bool
	OnlyMatch   bool
}

func RunDistributed(lines []string, opts RemoteOptions) ([]worker.Match, error) {
	chunks := splitter.Split(lines, len(opts.Peers))

	resultsCh := make(chan worker.Result, len(opts.Peers))

	for i, peer := range opts.Peers {
		var chunk splitter.Chunk
		if i < len(chunks) {
			chunk = chunks[i]
		}
		go func(addr string, ch splitter.Chunk) {
			resultsCh <- callWorker(addr, ch, opts)
		}(normalizeAddr(peer), chunk)
	}

	failures := 0
	var allMatches []worker.Match

	for range len(opts.Peers) {
		r := <-resultsCh
		if r.Err != nil {
			failures++
			slog.Warn("worker error", "err", r.Err)
		} else {
			allMatches = append(allMatches, r.Matches...)
		}
	}

	successes := len(opts.Peers) - failures
	if successes < opts.Quorum {
		return nil, fmt.Errorf("quorum not reached: %d/%d succeeded, need %d",
			successes, len(opts.Peers), opts.Quorum)
	}

	sortByLine(allMatches)
	return allMatches, nil
}

func callWorker(addr string, chunk splitter.Chunk, opts RemoteOptions) worker.Result {
	req := GrepRequest{
		Pattern:     opts.Pattern,
		Lines:       chunk.Lines,
		StartLine:   chunk.StartLine,
		IgnoreCase:  opts.IgnoreCase,
		InvertMatch: opts.InvertMatch,
		OnlyMatch:   opts.OnlyMatch,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return worker.Result{Err: fmt.Errorf("marshal request: %w", err)}
	}

	client := &http.Client{Timeout: _workerHTTPTimeout}

	httpReq, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		addr+"/grep",
		bytes.NewReader(body),
	)
	if err != nil {
		return worker.Result{Err: fmt.Errorf("build request: %w", err)}
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return worker.Result{Err: fmt.Errorf("worker %s unreachable: %w", addr, err)}
	}
	defer resp.Body.Close()

	var gResp GrepResponse
	if decErr := json.NewDecoder(resp.Body).Decode(&gResp); decErr != nil {
		return worker.Result{Err: fmt.Errorf("worker %s bad response: %w", addr, decErr)}
	}
	if gResp.Error != "" {
		return worker.Result{Err: fmt.Errorf("worker %s: %s", addr, gResp.Error)}
	}

	matches := make([]worker.Match, 0, len(gResp.Matches))
	for _, m := range gResp.Matches {
		matches = append(matches, worker.Match{
			LineNum: m.LineNum,
			Line:    m.Line,
			Excerpt: m.Excerpt,
		})
	}
	return worker.Result{Matches: matches, Count: gResp.Count}
}

func normalizeAddr(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "http://localhost" + addr
	}
	if !strings.HasPrefix(addr, "http") {
		return "http://" + addr
	}
	return addr
}

func sortByLine(m []worker.Match) {
	for i := 1; i < len(m); i++ {
		key := m[i]
		j := i - 1
		for j >= 0 && m[j].LineNum > key.LineNum {
			m[j+1] = m[j]
			j--
		}
		m[j+1] = key
	}
}
