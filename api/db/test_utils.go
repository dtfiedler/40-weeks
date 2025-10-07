package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"simple-go/api/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	testDBPath := filepath.Join(os.TempDir(), fmt.Sprintf("test_%d.db", os.Getpid()))
	
	testDB, err := sql.Open("sqlite3", testDBPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	driver, err := sqlite3.WithInstance(testDB, &sqlite3.Config{})
	if err != nil {
		t.Fatalf("Failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		t.Fatalf("Failed to create migration instance: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		testDB.Close()
		os.Remove(testDBPath)
	})

	return testDB
}

func SetupTestConfig() {
	config.AppConfig = &config.Config{
		JWTSecret:   "test-jwt-secret",
		ServerPort:  "8080",
		DatabaseURL: ":memory:",
	}
}

func SetupTestDatabase(t *testing.T) {
	t.Helper()
	
	testDB := SetupTestDB(t)
	
	originalDB := database
	database = testDB
	
	t.Cleanup(func() {
		database = originalDB
	})
}