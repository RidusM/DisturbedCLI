package splitter

type Chunk struct {
	StartLine int
	Lines     []string
}

func Split(lines []string, n int) []Chunk {
	if n <= 0 {
		n = 1
	}
	total := len(lines)
	if total == 0 {
		return nil
	}
	if n > total {
		n = total
	}

	chunks := make([]Chunk, 0, n)
	size := total / n
	remainder := total % n
	offset := 0

	for i := range n {
		chunkLen := size
		if i < remainder {
			chunkLen++
		}
		if chunkLen == 0 {
			break
		}
		chunks = append(chunks, Chunk{
			StartLine: offset + 1,
			Lines:     lines[offset : offset+chunkLen],
		})
		offset += chunkLen
	}
	return chunks
}
