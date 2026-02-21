package vault

import (
	"fmt"
	"os"
	"time"
)

func ensureExpectedMtime(fullPath, expected string) error {
	if expected == "" {
		return nil
	}

	expectedTime, err := time.Parse(time.RFC3339Nano, expected)
	if err != nil {
		return fmt.Errorf("invalid expected_mtime, must be RFC3339Nano: %v", err)
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("target does not exist for expected_mtime check")
		}
		return fmt.Errorf("failed to stat file for expected_mtime check: %v", err)
	}

	actual := info.ModTime().UTC()
	if !actual.Equal(expectedTime.UTC()) {
		return fmt.Errorf("mtime mismatch: expected %s, got %s", expectedTime.UTC().Format(time.RFC3339Nano), actual.Format(time.RFC3339Nano))
	}
	return nil
}

func fileMtimeRFC3339Nano(fullPath string) (string, error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return "", err
	}
	return info.ModTime().UTC().Format(time.RFC3339Nano), nil
}
