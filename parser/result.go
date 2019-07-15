package parser

import (
	"strings"
)

// ParseResult returns test statuses for pass and fail
func ParseResult(result string) (pass, fail int) {
	for _, line := range strings.Split(result, "\n") {
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
