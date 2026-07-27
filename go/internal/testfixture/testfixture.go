package testfixture

import (
	"errors"
	"os"
	"testing"
)

func Read(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		Skip(t, path, err)
	}
	return data
}

func Skip(t *testing.T, path string, err error) {
	t.Helper()
	if errors.Is(err, os.ErrNotExist) {
		t.Skipf("skipping: fixture %s not seeded (run seed_test_data.py)", path)
	}
	t.Fatalf("read fixture %s: %v", path, err)
}

func RequireAll(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			Skip(t, path, err)
		}
	}
}
