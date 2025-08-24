// utility/env_common.go
package Utility

import (
	"errors"
	"os"
)

// Cross-platform environment variable helpers

// SetEnvironmentVariable sets a variable in the current process environment.
func SetEnvironmentVariable(key string, value string) error {
	return os.Setenv(key, value)
}

// GetEnvironmentVariable retrieves a variable from the current process environment.
func GetEnvironmentVariable(key string) (string, error) {
	return os.Getenv(key), nil
}

// UnsetEnvironmentVariable removes a variable from the current process environment.
func UnsetEnvironmentVariable(key string) error {
	return os.Unsetenv(key)
}

// Windows-specific stubs — these are implemented in env_windows.go.
// On non-Windows they just return an error.

func SetWindowsEnvironmentVariable(key string, value string) error {
	return errors.New("SetWindowsEnvironmentVariable is available on windows only")
}

func GetWindowsEnvironmentVariable(key string) (string, error) {
	return "", errors.New("GetWindowsEnvironmentVariable is available on windows only")
}

func UnsetWindowsEnvironmentVariable(key string) error {
	return errors.New("UnsetWindowsEnvironmentVariable is available on windows only")
}

