package db

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/alexedwards/argon2id"

	// select sqlite3 driver for database/sql.
	_ "github.com/mattn/go-sqlite3"
)

// Repository stores and access data in a sqlite database.
type Repository struct {
	db *sql.DB
}

func createTables(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS users (" +
		"email TEXT PRIMARY KEY NOT NULL, " +
		"password TEXT NOT NULL, " +
		"created_at INTEGER, " +
		"updated_at INTEGER, " +
		"lastseen_at INTEGER)")
	if err != nil {
		return fmt.Errorf("could not create user table: %w", err)
	}

	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("creting password hash failed: %w", err)
	}

	return hash, nil
}

func closeStatement(stmt io.Closer) {
	if err := stmt.Close(); err != nil {
		log.Fatalf("Could not close DB statement: %v", err)
	}
}

func closeRows(r *sql.Rows) {
	if err := r.Close(); err != nil {
		log.Fatalf("Could not close DB ooooo: %v", err)
	}

	if err := r.Err(); err != nil {
		log.Fatalf("Could not close DB ooooo with error: %v", err)
	}
}

func validatePasswordAgainstHash(password string, hash string) bool {
	valid, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		log.Printf("could not validte password hash: %v", err)

		return false
	}

	return valid
}

// NewRepository returns Repository with a new database connexion.
func NewRepository() (*Repository, error) {
	connexion, err := sql.Open("sqlite3", "./fringe.db")
	if err != nil {
		return nil, fmt.Errorf("could not open or create ./fringe.db: %w", err)
	}

	err = createTables(connexion)
	if err != nil {
		return nil, err
	}

	return &Repository{db: connexion}, nil
}

// CreateUser INSERT OR IGNORE a new user record with email and argon2id password hash from the provided password.
func (r *Repository) CreateUser(email string, password string) (bool, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return false, err
	}

	now := time.Now()
	// Insert or update record in the database
	insert, err := r.db.Prepare("INSERT OR IGNORE INTO users (email, password, created_at, updated_at, lastseen_at) VALUES (?,?,?,?,?)") //nolint:sqlclosecheck
	defer closeStatement(insert)

	if err != nil {
		return false, fmt.Errorf("could not prepare user insert statement: %w", err)
	}

	result, err := insert.Exec(email, hash, now.Unix(), now.Unix(), now.Unix())
	if err != nil {
		return false, fmt.Errorf("insert user statement failed: %w", err)
	}

	lastInsertID, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()
	log.Printf("INSERT: LastInsertID: %d, RowsAffected: %d", lastInsertID, rowsAffected)

	return rowsAffected >= 1, nil
}

// UpdateUserPassword replaces the specified user's (found by email address) by the password provided.
// The password is not stored as is. It is hashed with argon2id.
func (r *Repository) UpdateUserPassword(email string, password string) (bool, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return false, err
	}

	now := time.Now()
	// Insert or update record in the database
	insert, err := r.db.Prepare("UPDATE users SET password = ?,  updated_at = ?, lastseen_at = ? WHERE email = ?") //nolint:sqlclosecheck
	defer closeStatement(insert)

	if err != nil {
		return false, fmt.Errorf("cloud not prepaer update statement: %w", err)
	}

	result, err := insert.Exec(hash, now.Unix(), now.Unix(), email)
	if err != nil {
		log.Printf("user %s password update failed: %v", email, err)

		return false, fmt.Errorf("user %s password update failed: %w", email, err)
	}

	rowsAffected, _ := result.RowsAffected()

	return rowsAffected >= 1, nil
}

// TouchUser updates user's lastseen_at value with current Unix time.
func (r *Repository) TouchUser(email string) (bool, error) {
	now := time.Now()

	update, err := r.db.Prepare("UPDATE users SET lastseen_at = ? WHERE email = ?") //nolint:sqlclosecheck
	defer closeStatement(update)

	if err != nil {
		return false, fmt.Errorf("could not prepare update statement: %w", err)
	}

	result, err := update.Exec(now.Unix(), email)
	if err != nil {
		return false, fmt.Errorf("last seen update has failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("could not confirm number of rows affected: %w", err)
	}

	return rowsAffected >= 1, nil
}

// AuthenticateUser validates if the email and password combination matches an existing user
// in the database with a password resulting in the same password hash.
func (r *Repository) AuthenticateUser(email string, password string) (bool, error) {
	stmt, err := r.db.Prepare(`SELECT email, password FROM users WHERE email = ? LIMIT 1`) //nolint:sqlclosecheck
	defer closeStatement(stmt)

	if err != nil {
		return false, fmt.Errorf("could not prepare select statement: %w", err)
	}

	rows, err := stmt.Query(email) //nolint:sqlclosecheck,rowserrcheck
	defer closeRows(rows)

	if err != nil {
		return false, fmt.Errorf("could execute select for user %s: %w", email, err)
	}

	for rows.Next() {
		var queryEmail, queryHash string

		err := rows.Scan(&queryEmail, &queryHash)
		if err != nil {
			return false, fmt.Errorf("could not read data from user row for user %s: %w", email, err)
		}

		if queryEmail == email {
			return validatePasswordAgainstHash(password, queryHash), nil
		}
	}

	return false, nil
}

// Close the database connexion. No more actions can be taken with Repository.
func (r *Repository) Close() {
	if err := r.db.Close(); err != nil {
		log.Fatalf("Could  not close DB: %v", err)
	}
}
