package coordinator

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"disturbedcli/internal/splitter"
	"disturbedcli/internal/worker"
)

const (
	_workerReadTimeout  = 15 * time.Second
	_workerWriteTimeout = 30 * time.Second
	_workerIdleTimeout  = 60 * time.Second
)

func ServeWorker(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/grep", handleGrep)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  _workerReadTimeout,
		WriteTimeout: _workerWriteTimeout,
		IdleTimeout:  _workerIdleTimeout,
	}

	slog.Info("worker listening", "addr", addr)
	return srv.Serve(ln)
}

func handleGrep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	var req GrepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request: "+err.Error())
		return
	}

	re, err := worker.CompilePattern(req.Pattern, req.IgnoreCase)
	if err != nil {
		respondError(w, "invalid pattern: "+err.Error())
		return
	}

	chunk := splitter.Chunk{
		StartLine: req.StartLine,
		Lines:     req.Lines,
	}

	result := worker.Run(chunk, re, worker.Options{
		IgnoreCase:  req.IgnoreCase,
		InvertMatch: req.InvertMatch,
		OnlyMatch:   req.OnlyMatch,
	})

	resp := GrepResponse{Count: result.Count}
	for _, m := range result.Matches {
		resp.Matches = append(resp.Matches, MatchResult{
			LineNum: m.LineNum,
			Line:    m.Line,
			Excerpt: m.Excerpt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("encode response", "err", err)
	}
}

func respondError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(GrepResponse{Error: msg}); err != nil {
		slog.Error("encode error response", "err", err)
	}
}
