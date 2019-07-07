package parser

import (
	"strings"
)

// ParseOutput returns test statuses for pass and fail
func ParseOutput(output string) (pass, fail int) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.Index(trimmedLine, "--- PASS:") == 0 {
			pass++
		} else if strings.Index(trimmedLine, "--- FAIL:") == 0 {
			fail++
		}
	}

	return pass, fail
}
