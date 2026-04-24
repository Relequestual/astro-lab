package auth

import (
	"regexp"
	"strings"
)

var tokenPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(ghp_[a-zA-Z0-9]{36,})`),
	regexp.MustCompile(`(gho_[a-zA-Z0-9]{36,})`),
	regexp.MustCompile(`(github_pat_[a-zA-Z0-9_]{22,})`),
}

func RedactSecrets(s string) string {
	for _, p := range tokenPatterns {
		s = p.ReplaceAllStringFunc(s, func(match string) string {
			if len(match) > 8 {
				return match[:4] + strings.Repeat("*", len(match)-8) + match[len(match)-4:]
			}
			return strings.Repeat("*", len(match))
		})
	}
	return s
}
