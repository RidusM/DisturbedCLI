package merger

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	scannerBufSize = 64 * 1024
	scannerMaxSize = 10 * 1024 * 1024
)

func ReadLines(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, scannerBufSize)
	scanner.Buffer(buf, scannerMaxSize)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read lines: %w", err)
	}
	return lines, nil
}

func ReadFiles(paths []string) (map[string][]string, error) {
	result := make(map[string][]string, len(paths))
	var errs []string

	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		lines, err := ReadLines(f)
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", p, err))
			continue
		}
		result[p] = lines
	}

	if len(errs) > 0 {
		return result, fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return result, nil
}
