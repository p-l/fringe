package db

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/alexedwards/argon2id"
	"modernc.org/ql"
)

// Repository stores and access data in a sqlite database.
type Repository struct {
	db *sql.DB
}

func createTables(db *sql.DB) error {
	createTx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not create transaction %w", err)
	}

	_, err = createTx.Exec("CREATE TABLE IF NOT EXISTS users (" +
		"email string NOT NULL, " +
		"password string NOT NULL, " +
		"created_at int, " +
		"updated_at int, " +
		"last_seen_at int)")
	if err != nil {
		return fmt.Errorf("could not create user table: %w", err)
	}

	_, err = createTx.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users (email)")
	if err != nil {
		return fmt.Errorf("could not create user:email index: %w", err)
	}

	err = createTx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit user table and index: %w", err)
	}

	return nil
}

type dbUser struct {
	Email      string
	Password   string
	CreatedAt  int
	UpdatedAt  int
	LastSeenAt int
}

func (r *Repository) userByEmail(findEmail string) (user *dbUser, err error) {
	find, err := r.db.Prepare("SELECT email,password,created_at,updated_at,last_seen_at FROM users WHERE email == $1 LIMIT 1;") //nolint:sqlclosecheck
	if err != nil {
		return nil, fmt.Errorf("could not prepare select statement: %w", err)
	}
	defer closeStatement(find)

	rows, err := find.Query(findEmail) //nolint:sqlclosecheck,rowserrcheck
	defer closeRows(rows)

	if err != nil {
		return nil, fmt.Errorf("could execute select for user %s: %w", findEmail, err)
	}

	var u dbUser
	for rows.Next() {
		err := rows.Scan(&u.Email, &u.Password, &u.CreatedAt, &u.UpdatedAt, &u.LastSeenAt)
		if err != nil {
			return nil, fmt.Errorf("could not read data from user row for user %s: %w", findEmail, err)
		}

		if findEmail == u.Email {
			return &u, nil
		}
	}

	return nil, nil
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
func NewRepository(location string) (*Repository, error) {
	ql.RegisterDriver()

	connexion, err := sql.Open("ql", location)
	if err != nil {
		return nil, fmt.Errorf("could not open or create at %s: %w", location, err)
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

	user, err := r.userByEmail(email)
	if err != nil {
		return false, err
	}

	if user != nil {
		// User not created as it already exists
		return false, nil
	}

	// Create User
	insertTx, err := r.db.Begin()
	if err != nil {
		return false, err
	}

	now := time.Now()

	// Insert record in the database
	insert, err := insertTx.Prepare("INSERT INTO users (email, password, created_at, updated_at, last_seen_at) VALUES ($1,$2,$3,$4,$5)") //nolint:sqlclosecheck
	if err != nil {
		return false, fmt.Errorf("could not prepare user insert statement: %w", err)
	}
	defer closeStatement(insert)

	result, err := insert.Exec(email, hash, now.Unix(), now.Unix(), now.Unix())
	if err != nil {
		return false, fmt.Errorf("insert user statement failed: %w", err)
	}

	err = insertTx.Commit()
	if err != nil {
		return false, err
	}

	lastInsertID, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()
	log.Printf("INSERT new user:%s, rowID: %d, RowsAffected: %d", email, lastInsertID, rowsAffected)

	return rowsAffected >= 1, nil
}

// UpdateUserPassword replaces the specified user's (found by email address) by the password provided.
// The password is not stored as is. It is hashed with argon2id.
func (r *Repository) UpdateUserPassword(email string, password string) (bool, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return false, err
	}

	updateTx, err := r.db.Begin()
	if err != nil {
		return false, fmt.Errorf("cloud not being transaction: %w", err)
	}

	now := time.Now()
	// Insert or update record in the database
	stmt, err := updateTx.Prepare("UPDATE users SET password = $1, updated_at = $2, last_seen_at = $3 WHERE email == $4") //nolint:sqlclosecheck
	if err != nil {
		return false, fmt.Errorf("cloud not prepaer update statement: %w", err)
	}

	defer closeStatement(stmt)

	result, err := stmt.Exec(hash, now.Unix(), now.Unix(), email)
	if err != nil {
		return false, fmt.Errorf("user %s password update failed: %w", email, err)
	}

	err = updateTx.Commit()
	if err != nil {
		return false, fmt.Errorf("user %s password update failed: %w", email, err)
	}

	rowsAffected, _ := result.RowsAffected()

	return rowsAffected >= 1, nil
}

// TouchUser updates user's last_seen_at value with current Unix time.
func (r *Repository) TouchUser(email string) (bool, error) {
	now := time.Now()

	updateTx, err := r.db.Begin()
	if err != nil {
		return false, fmt.Errorf("could not prepare transaction: %w", err)
	}

	update, err := updateTx.Prepare("UPDATE users SET last_seen_at = $1 WHERE email == $2") //nolint:sqlclosecheck
	defer closeStatement(update)

	if err != nil {
		return false, fmt.Errorf("could not prepare update statement: %w", err)
	}

	result, err := update.Exec(now.Unix(), email)
	if err != nil {
		return false, fmt.Errorf("last seen update has failed: %w", err)
	}

	err = updateTx.Commit()
	if err != nil {
		return false, fmt.Errorf("last_seen_at commit failed: %w", err)
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
