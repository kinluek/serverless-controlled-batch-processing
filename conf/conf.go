package conf

import (
	"fmt"
	"os"
)

// LoadEnvDefault returns the environment variable for the given name, if the variable is empty
// then the default value is returned.
func GetEnvDefault(envName, defaultVal string) string {
	val := os.Getenv(envName)
	if val == "" {
		return defaultVal
	}
	return val
}

// GetEnvRequired returns the environment variable for the given name, if the variable is empty
// then
func GetEnvRequired(envName string) (string, error) {
	val := os.Getenv(envName)
	if val == "" {
		return "", fmt.Errorf("envar %s not set", envName)
	}
	return val, nil
}
