package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultConfig(t *testing.T) {
	dir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	cfg := Load()

	if cfg.ServerAddress != "0.0.0.0:8325" {
		t.Fatalf("expected default server address 0.0.0.0:8325, got %q", cfg.ServerAddress)
	}
	if cfg.Neo4jURI != "neo4j://localhost:7687" {
		t.Fatalf("expected default Neo4j URI neo4j://localhost:7687, got %q", cfg.Neo4jURI)
	}
	if cfg.Neo4jUser != "neo4j" {
		t.Fatalf("expected default Neo4j user neo4j, got %q", cfg.Neo4jUser)
	}
}

func TestLoad_INIOverrides(t *testing.T) {
	dir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	contents := []byte(`[server]
address = :9000

[neo4j]
uri = bolt://example:7687
user = testuser
password = testpass
database = testdb
`)

	if err := os.WriteFile(filepath.Join(dir, ".mate.ini"), contents, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := Load()

	if cfg.ServerAddress != ":9000" {
		t.Fatalf("expected server address :9000, got %q", cfg.ServerAddress)
	}
	if cfg.Neo4jURI != "bolt://example:7687" {
		t.Fatalf("expected Neo4j URI bolt://example:7687, got %q", cfg.Neo4jURI)
	}
	if cfg.Neo4jUser != "testuser" {
		t.Fatalf("expected Neo4j user testuser, got %q", cfg.Neo4jUser)
	}
	if cfg.Neo4jPassword != "testpass" {
		t.Fatalf("expected Neo4j password testpass, got %q", cfg.Neo4jPassword)
	}
	if cfg.Neo4jDatabase != "testdb" {
		t.Fatalf("expected Neo4j database testdb, got %q", cfg.Neo4jDatabase)
	}
}
