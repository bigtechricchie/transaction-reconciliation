package main

import (
	"os"
	"testing"
)

func TestGetEnvironmentPortUsesDefaultWhenUnset(t *testing.T) {
	previousPort, portWasConfigured := os.LookupEnv("PORT")
	if err := os.Unsetenv("PORT"); err != nil {
		t.Fatalf("failed to unset PORT: %v", err)
	}
	t.Cleanup(func() {
		if portWasConfigured {
			if err := os.Setenv("PORT", previousPort); err != nil {
				t.Errorf("failed to restore PORT: %v", err)
			}
			return
		}

		if err := os.Unsetenv("PORT"); err != nil {
			t.Errorf("failed to keep PORT unset: %v", err)
		}
	})

	port, err := getEnvironmentPort()
	if err != nil {
		t.Fatalf("getEnvironmentPort() returned an unexpected error: %v", err)
	}

	if port != defaultPort {
		t.Fatalf("port = %d, want %d", port, defaultPort)
	}
}

func TestGetEnvironmentPortAcceptsExplicitPort(t *testing.T) {
	t.Setenv("PORT", "8080")

	port, err := getEnvironmentPort()
	if err != nil {
		t.Fatalf("getEnvironmentPort() returned an unexpected error: %v", err)
	}

	if port != 8080 {
		t.Fatalf("port = %d, want 8080", port)
	}
}

func TestGetEnvironmentPortRejectsEmptyValue(t *testing.T) {
	t.Setenv("PORT", "")

	_, err := getEnvironmentPort()
	if err == nil {
		t.Fatal("getEnvironmentPort() expected an error for an empty PORT")
	}
}

func TestGetEnvironmentPortRejectsNonIntegerValue(t *testing.T) {
	t.Setenv("PORT", "not-a-port")

	_, err := getEnvironmentPort()
	if err == nil {
		t.Fatal("getEnvironmentPort() expected an error for a non-integer PORT")
	}
}

func TestGetEnvironmentPortRejectsOutOfRangeValues(t *testing.T) {
	testCases := []string{"0", "65536", "-1"}

	for _, portValue := range testCases {
		t.Run(portValue, func(t *testing.T) {
			t.Setenv("PORT", portValue)

			_, err := getEnvironmentPort()
			if err == nil {
				t.Fatalf(
					"getEnvironmentPort() expected an error for PORT=%s",
					portValue,
				)
			}
		})
	}
}
