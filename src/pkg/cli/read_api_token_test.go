package cli

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestReadAPIToken_FromFile(t *testing.T) {
	apiToken, err := ReadAPIToken("testdata/api-token.txt", "TEST_API_TOKEN")
	require.NoError(t, err)
	require.Equal(t, "Fake_API_Token", apiToken)
}

func TestReadAPIToken_FromEnvVar(t *testing.T) {
	err := os.Setenv("TEST_API_TOKEN", "foozle")
	require.NoError(t, err)

	defer func() { _ = os.Unsetenv("TEST_API_TOKEN") }()

	apiToken, err := ReadAPIToken("", "TEST_API_TOKEN")
	require.NoError(t, err)
	require.Equal(t, "foozle", apiToken)
}

func TestReadAPIToken_NoFileOrEnvSet(t *testing.T) {
	_, err := ReadAPIToken("", "TEST_API_TOKEN")
	require.Error(t, err)
	require.Contains(t, err.Error(), "TEST_API_TOKEN not set")
}
