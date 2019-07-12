package parser

import (
	"strings"
)

// ParseResult returns test statuses for pass and fail
func ParseResult(output string) (pass, fail int) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "--- PASS"):
			pass++
		case strings.HasPrefix(trimmed, "--- FAIL"):
			fail++
		}
	}

	return pass, fail
}
