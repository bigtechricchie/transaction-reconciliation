package main

import (
	"os"
	"testing"
)

func TestGetEnvironmentHostUsesSecureDefaultWhenUnset(t *testing.T) {
	previousHost, hostWasConfigured := os.LookupEnv("HOST")
	if err := os.Unsetenv("HOST"); err != nil {
		t.Fatalf("failed to unset HOST: %v", err)
	}
	t.Cleanup(func() {
		if hostWasConfigured {
			if err := os.Setenv("HOST", previousHost); err != nil {
				t.Errorf("failed to restore HOST: %v", err)
			}
			return
		}

		if err := os.Unsetenv("HOST"); err != nil {
			t.Errorf("failed to keep HOST unset: %v", err)
		}
	})

	host, err := getEnvironmentHost()
	if err != nil {
		t.Fatalf("getEnvironmentHost() returned an unexpected error: %v", err)
	}

	if host != defaultHost {
		t.Fatalf("host = %q, want %q", host, defaultHost)
	}
}

func TestGetEnvironmentHostAcceptsExplicitHost(t *testing.T) {
	t.Setenv("HOST", "0.0.0.0")

	host, err := getEnvironmentHost()
	if err != nil {
		t.Fatalf("getEnvironmentHost() returned an unexpected error: %v", err)
	}

	if host != "0.0.0.0" {
		t.Fatalf("host = %q, want %q", host, "0.0.0.0")
	}
}

func TestGetEnvironmentHostRejectsExplicitlyEmptyHost(t *testing.T) {
	t.Setenv("HOST", "")

	_, err := getEnvironmentHost()
	if err == nil {
		t.Fatal("getEnvironmentHost() expected an error for an explicitly empty HOST")
	}
}

func TestGetEnvironmentHostRejectsWhitespace(t *testing.T) {
	t.Setenv("HOST", "127.0.0.1 invalid")

	_, err := getEnvironmentHost()
	if err == nil {
		t.Fatal("getEnvironmentHost() expected an error for HOST containing whitespace")
	}
}
