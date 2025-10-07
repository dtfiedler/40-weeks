package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"simple-go/api/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Name     string
	Password string
	Email    string
	IsAdmin  bool
	Created  time.Time
}

var database *sql.DB

// GetDB returns the database connection
func GetDB() *sql.DB {
	return database
}

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

func GetUserByName(name string) (*User, error) {
	user := &User{}
	err := database.QueryRow(
		"SELECT id, name, password, email, is_admin, created FROM users WHERE name = ?",
		name,
	).Scan(&user.ID, &user.Name, &user.Password, &user.Email, &user.IsAdmin, &user.Created)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := database.QueryRow(
		"SELECT id, name, password, email, is_admin, created FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Name, &user.Password, &user.Email, &user.IsAdmin, &user.Created)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func CreateUser(name, password, email string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = database.Exec(
		"INSERT INTO users (name, password, email) VALUES (?, ?, ?)",
		name, string(hashedPassword), email,
	)
	return err
}

func GetAllUsers(limit int) ([]map[string]interface{}, error) {
	rows, err := database.Query("SELECT id, name, email, created FROM users ORDER BY created DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var name, email string
		var created time.Time
		if err := rows.Scan(&id, &name, &email, &created); err != nil {
			continue
		}
		users = append(users, map[string]interface{}{
			"id":      fmt.Sprintf("%d", id),
			"name":    name,
			"email":   email,
			"created": created.Format("2006-01-02 15:04:05"),
		})
	}

	return users, nil
}


func GetUserByID(userID int) (*User, error) {
	user := &User{}
	err := database.QueryRow(
		"SELECT id, name, password, email, is_admin, created FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Name, &user.Password, &user.Email, &user.IsAdmin, &user.Created)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func CloseDB() {
	if database != nil {
		database.Close()
	}
}