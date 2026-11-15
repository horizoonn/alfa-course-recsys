package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadEnvFile(t *testing.T) {
	for _, tc := range []struct {
		name     string
		parent   bool
		existing string
	}{{"current directory", false, ""}, {"parent directory", true, ""}, {"environment priority", false, "127.0.0.1:9090"}} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, ".env")
			if err := os.WriteFile(path, []byte("DATABASE_URL=postgres://example/db\nHTTP_ADDR=127.0.0.1:8088\n"), 0600); err != nil {
				t.Fatal(err)
			}
			directory := root
			if tc.parent {
				directory = filepath.Join(root, "backend")
				if err := os.Mkdir(directory, 0700); err != nil {
					t.Fatal(err)
				}
			}
			t.Chdir(directory)
			for _, name := range []string{"DATABASE_URL", "HTTP_ADDR"} {
				t.Setenv(name, "")
				if err := os.Unsetenv(name); err != nil {
					t.Fatal(err)
				}
			}
			if tc.existing != "" {
				t.Setenv("HTTP_ADDR", tc.existing)
			}
			cfg, err := Load()
			if err != nil {
				t.Fatal(err)
			}
			expected := tc.existing
			if expected == "" {
				expected = "127.0.0.1:8088"
			}
			if cfg.DatabaseURL != "postgres://example/db" || cfg.HTTPAddr != expected {
				t.Fatal("unexpected configuration")
			}
		})
	}
}

func TestOptionalAndInvalidEnvFile(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("DATABASE_URL", "postgres://example/db")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".env", []byte("PASSWORD=\"private-value"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := Load()
	if err == nil || strings.Contains(err.Error(), "private-value") {
		t.Fatal("invalid env handling")
	}
}
