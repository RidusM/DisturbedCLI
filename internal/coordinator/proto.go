package coordinator

type GrepRequest struct {
	Pattern     string   `json:"pattern"`
	Lines       []string `json:"lines"`
	StartLine   int      `json:"start_line"`
	IgnoreCase  bool     `json:"ignore_case"`
	InvertMatch bool     `json:"invert_match"`
	OnlyMatch   bool     `json:"only_match"`
}

type MatchResult struct {
	LineNum int    `json:"line_num"`
	Line    string `json:"line"`
	Excerpt string `json:"excerpt,omitempty"`
}

type GrepResponse struct {
	Matches []MatchResult `json:"matches"`
	Count   int           `json:"count"`
	Error   string        `json:"error,omitempty"`
}
