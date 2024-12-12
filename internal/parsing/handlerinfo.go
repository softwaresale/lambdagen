package parsing

import (
	"errors"
	"regexp"
)

// ParseHttpInfo pulls the handler route info from the arg string for a handler function
func ParseHttpInfo(args string) (string, string, error) {
	parser := regexp.MustCompile(`(GET|POST|PUT|PATCH|DELETE)\s+(\S+)`)
	matches := parser.FindStringSubmatch(args)
	if matches == nil {
		return "", "", errors.New("failed to parse http info")
	}

	return matches[1], matches[2], nil
}
