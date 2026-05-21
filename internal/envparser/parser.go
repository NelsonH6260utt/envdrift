package envparser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// EnvMap represents a map of environment variable key-value pairs.
type EnvMap map[string]string

// ParseFile reads a .env file and returns an EnvMap.
// Lines starting with '#' are treated as comments and ignored.
// Empty lines are also ignored.
func ParseFile(path string) (EnvMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("envparser: opening file %q: %w", path, err)
	}
	defer f.Close()

	envMap := make(EnvMap)
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("envparser: file %q line %d: %w", path, lineNum, err)
		}

		envMap[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("envparser: scanning file %q: %w", path, err)
	}

	return envMap, nil
}

// parseLine splits a single KEY=VALUE line into its components.
func parseLine(line string) (string, string, error) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid format, expected KEY=VALUE, got %q", line)
	}

	key := strings.TrimSpace(parts[0])
	if key == "" {
		return "", "", fmt.Errorf("empty key in line %q", line)
	}

	value := strings.TrimSpace(parts[1])
	value = stripQuotes(value)

	return key, value, nil
}

// stripQuotes removes surrounding single or double quotes from a value.
func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
