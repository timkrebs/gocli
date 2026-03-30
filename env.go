package cli

import (
	"os"
	"strconv"
)

// EnvDefault returns the value of the environment variable named by envKey.
// If the variable is unset or empty, fallback is returned instead.
//
// Typical use: supply an environment-variable override as a flag default so
// that the flag still takes precedence when explicitly provided by the user.
//
//	fs.String("token", cli.EnvDefault("APP_TOKEN", ""), "API token ($APP_TOKEN)")
func EnvDefault(envKey, fallback string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return fallback
}

// EnvDefaultBool returns the boolean value of the environment variable named
// by envKey. If the variable is unset, empty, or cannot be parsed as a bool,
// fallback is returned.
//
// Accepted truthy values (case-insensitive): 1, t, T, TRUE, true, True.
// Accepted falsy values: 0, f, F, FALSE, false, False.
//
//	fs.Bool("verbose", cli.EnvDefaultBool("APP_VERBOSE", false), "verbose output ($APP_VERBOSE)")
func EnvDefaultBool(envKey string, fallback bool) bool {
	v := os.Getenv(envKey)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

// EnvDefaultInt returns the integer value of the environment variable named
// by envKey. If the variable is unset, empty, or cannot be parsed as an int,
// fallback is returned.
//
//	fs.Int("port", cli.EnvDefaultInt("APP_PORT", 8080), "listen port ($APP_PORT)")
func EnvDefaultInt(envKey string, fallback int) int {
	v := os.Getenv(envKey)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
