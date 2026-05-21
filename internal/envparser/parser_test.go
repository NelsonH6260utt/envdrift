package envparser

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempEnv(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp env file: %v", err)
	}
	return path
}

func TestParseFile_BasicKeyValue(t *testing.T) {
	path := writeTempEnv(t, "APP_ENV=production\nPORT=8080\n")
	got, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["APP_ENV"] != "production" {
		t.Errorf("APP_ENV: got %q, want %q", got["APP_ENV"], "production")
	}
	if got["PORT"] != "8080" {
		t.Errorf("PORT: got %q, want %q", got["PORT"], "8080")
	}
}

func TestParseFile_IgnoresComments(t *testing.T) {
	path := writeTempEnv(t, "# this is a comment\nKEY=value\n")
	got, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 entry, got %d", len(got))
	}
}

func TestParseFile_StripQuotes(t *testing.T) {
	path := writeTempEnv(t, `DB_URL="postgres://localhost/mydb"` + "\n" + `SECRET='abc123'` + "\n")
	got, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["DB_URL"] != "postgres://localhost/mydb" {
		t.Errorf("DB_URL: got %q", got["DB_URL"])
	}
	if got["SECRET"] != "abc123" {
		t.Errorf("SECRET: got %q", got["SECRET"])
	}
}

func TestParseFile_InvalidLine(t *testing.T) {
	path := writeTempEnv(t, "BADLINE\n")
	_, err := ParseFile(path)
	if err == nil {
		t.Fatal("expected error for invalid line, got nil")
	}
}

func TestParseFile_FileNotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/.env")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestParseFile_EmptyFile(t *testing.T) {
	path := writeTempEnv(t, "")
	got, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}
