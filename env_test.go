package cli

import (
	"testing"
)

func TestEnvDefault(t *testing.T) {
	t.Run("returns env value when set", func(t *testing.T) {
		t.Setenv("TEST_ENV_STR", "hello")
		if got := EnvDefault("TEST_ENV_STR", "default"); got != "hello" {
			t.Fatalf("want %q, got %q", "hello", got)
		}
	})

	t.Run("returns fallback when unset", func(t *testing.T) {
		if got := EnvDefault("TEST_ENV_STR_UNSET_XYZ", "default"); got != "default" {
			t.Fatalf("want %q, got %q", "default", got)
		}
	})

	t.Run("returns fallback when empty", func(t *testing.T) {
		t.Setenv("TEST_ENV_STR_EMPTY", "")
		if got := EnvDefault("TEST_ENV_STR_EMPTY", "default"); got != "default" {
			t.Fatalf("want %q, got %q", "default", got)
		}
	})
}

func TestEnvDefaultBool(t *testing.T) {
	cases := []struct {
		val      string
		expected bool
	}{
		{"1", true},
		{"t", true},
		{"T", true},
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"0", false},
		{"f", false},
		{"false", false},
		{"FALSE", false},
	}

	for _, tc := range cases {
		t.Run(tc.val, func(t *testing.T) {
			t.Setenv("TEST_ENV_BOOL", tc.val)
			if got := EnvDefaultBool("TEST_ENV_BOOL", !tc.expected); got != tc.expected {
				t.Fatalf("val=%q: want %v, got %v", tc.val, tc.expected, got)
			}
		})
	}

	t.Run("returns fallback when unset", func(t *testing.T) {
		if got := EnvDefaultBool("TEST_ENV_BOOL_UNSET_XYZ", true); got != true {
			t.Fatalf("want true, got %v", got)
		}
	})

	t.Run("returns fallback on invalid value", func(t *testing.T) {
		t.Setenv("TEST_ENV_BOOL_BAD", "notabool")
		if got := EnvDefaultBool("TEST_ENV_BOOL_BAD", true); got != true {
			t.Fatalf("want true (fallback), got %v", got)
		}
	})
}

func TestEnvDefaultInt(t *testing.T) {
	t.Run("returns env value when set", func(t *testing.T) {
		t.Setenv("TEST_ENV_INT", "9090")
		if got := EnvDefaultInt("TEST_ENV_INT", 8080); got != 9090 {
			t.Fatalf("want 9090, got %d", got)
		}
	})

	t.Run("returns fallback when unset", func(t *testing.T) {
		if got := EnvDefaultInt("TEST_ENV_INT_UNSET_XYZ", 8080); got != 8080 {
			t.Fatalf("want 8080, got %d", got)
		}
	})

	t.Run("returns fallback on invalid value", func(t *testing.T) {
		t.Setenv("TEST_ENV_INT_BAD", "notanint")
		if got := EnvDefaultInt("TEST_ENV_INT_BAD", 42); got != 42 {
			t.Fatalf("want 42 (fallback), got %d", got)
		}
	})

	t.Run("handles negative values", func(t *testing.T) {
		t.Setenv("TEST_ENV_INT_NEG", "-1")
		if got := EnvDefaultInt("TEST_ENV_INT_NEG", 0); got != -1 {
			t.Fatalf("want -1, got %d", got)
		}
	})
}
