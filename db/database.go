package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"simple-go/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Username string
	Password string
	Email    string
	Created  time.Time
}

var database *sql.DB

func InitDB() error {
	var err error
	database, err = sql.Open("sqlite3", config.AppConfig.DatabaseURL)
	if err != nil {
		return err
	}

	// Run migrations using golang-migrate
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	log.Println("Database initialized successfully with migrations")
	return nil
}

func runMigrations() error {
	driver, err := sqlite3.WithInstance(database, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
}

func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := database.QueryRow(
		"SELECT id, username, password, email, created FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Created)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func CreateUser(username, password, email string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"INSERT INTO users (username, password, email) VALUES (?, ?, ?)",
		username, string(hashedPassword), email,
	)
	return err
}

func GetAllUsers(limit int) ([]map[string]interface{}, error) {
	rows, err := database.Query("SELECT id, username, email FROM users ORDER BY created DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var username, email string
		if err := rows.Scan(&id, &username, &email); err != nil {
			continue
		}
		users = append(users, map[string]interface{}{
			"id":    fmt.Sprintf("%d", id),
			"name":  username,
			"email": email,
		})
	}

	return users, nil
}

func CloseDB() {
	if database != nil {
		database.Close()
	}
}

// GetDB returns the database instance for use in migrations and other packages
func GetDB() *sql.DB {
	return database
}