package cli

import (
	"fmt"
	"os"
	"strings"
)

// ReadAPIToken reads an API token. Takes an optional filename and env var.
// If the filename is specified, attempts to read from the file.
// If the filename is omitted, pulls from the environment variable.
func ReadAPIToken(fname string, envVar string) (string, error) {
	if len(fname) > 0 {
		tokenBytes, err := os.ReadFile(fname)
		if err != nil {
			return "", fmt.Errorf("unable to load API token file: %w", err)
		}

		return strings.TrimSpace(string(tokenBytes)), nil
	}

	apiToken := os.Getenv(envVar)
	if len(apiToken) == 0 {
		return "", fmt.Errorf("%s not set", envVar)
	}

	return apiToken, nil
}
