package auth

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ReadTokenFromInput reads a token from an input stream (for piped input)
func ReadTokenFromInput(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		token := strings.TrimSpace(scanner.Text())
		if token == "" {
			return "", fmt.Errorf("empty token provided")
		}
		return token, nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading token: %w", err)
	}
	return "", fmt.Errorf("no token provided")
}
