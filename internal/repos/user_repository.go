package repos

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/jmoiron/sqlx"
)

// UserRepository stores and access data in a sqlite database.
type UserRepository struct {
	db *sqlx.DB
}

type User struct {
	Email        string `db:"email"`
	PasswordHash string `db:"password"`
	CreatedAt    int64  `db:"created_at"`
	UpdatedAt    int64  `db:"updated_at"`
	LastSeenAt   int64  `db:"last_seen_at"`
}

var (
	ErrUserNotFound     = errors.New("queried user could not be not found")
	ErrUserAlreadyExist = errors.New("user with same email already exist in database")
)

// NewUserRepository returns a ready to use UserRepository with a new database connexion.
func NewUserRepository(db *sqlx.DB) (*UserRepository, error) {
	if err := createUserTable(db); err != nil {
		return nil, err
	}

	return &UserRepository{db: db}, nil
}

func createUserTable(db *sqlx.DB) error {
	createTx := db.MustBegin()
	defer func() { _ = createTx.Rollback() }()

	createTx.MustExec("CREATE TABLE IF NOT EXISTS users (" +
		"email string NOT NULL, " +
		"password string NOT NULL, " +
		"created_at int64, " +
		"updated_at int64, " +
		"last_seen_at int64)")
	createTx.MustExec("CREATE INDEX IF NOT EXISTS idx_users_email ON users (email)")

	if err := createTx.Commit(); err != nil {
		return fmt.Errorf("cannot create users table: %w", err)
	}

	return nil
}

func CreatePasswordHash(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("creting password hash failed: %w", err)
	}

	return hash, nil
}

func (u *User) PasswordMatch(password string) bool {
	valid, err := argon2id.ComparePasswordAndHash(password, u.PasswordHash)
	if err != nil {
		log.Printf("could not validte password hash: %v", err)

		return false
	}

	return valid
}

// UserWithEmail Looks for a user record that matches the provided email and returns a pointer to User stuck if found or nil if not.
func (r *UserRepository) UserWithEmail(email string) (user *User, err error) {
	var u User
	if err := r.db.Get(&u, "SELECT email,password,created_at,updated_at,last_seen_at FROM users WHERE email == $1 LIMIT 1", email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("could not retrieve user %s: %w", email, err)
	}

	return &u, nil
}

// CreateUser INSERT a new user record with email and argon2id password hash from the provided password.
// If the user already exists returns the record from the Database, otherwise return the newly created User.
func (r *UserRepository) CreateUser(email string, password string) (*User, error) {
	dbUser, err := r.UserWithEmail(email)
	if !errors.Is(err, ErrUserNotFound) && err != nil {
		return nil, err
	}

	// Return user if it already exists
	if dbUser != nil {
		return dbUser, ErrUserAlreadyExist
	}

	// Create User
	hash, err := CreatePasswordHash(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	newUser := User{email, hash, now.Unix(), now.Unix(), now.Unix()}

	insertTx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("could not create user %s: %w", email, err)
	}
	defer func() { _ = insertTx.Rollback() }() //nolint:wsl

	// Insert record in the database
	insert, err := insertTx.Prepare("INSERT INTO users (email, password, created_at, updated_at, last_seen_at) VALUES ($1,$2,$3,$4,$5)")
	if err != nil {
		return nil, fmt.Errorf("could not create user %s: %w", email, err)
	}
	defer insert.Close()

	_, err = insert.Exec(newUser.Email, newUser.PasswordHash, newUser.CreatedAt, newUser.UpdatedAt, newUser.LastSeenAt)
	if err != nil {
		return nil, fmt.Errorf("could not create user %s: %w", email, err)
	}

	if err := insertTx.Commit(); err != nil {
		return nil, fmt.Errorf("could not create user %s: %w", email, err)
	}

	return &newUser, nil
}

// UpdateUserPassword replaces the specified user's (found by email address) by the password provided.
// The password is not stored as is. It is hashed with argon2id.
func (r *UserRepository) UpdateUserPassword(email string, password string) (bool, error) {
	hash, err := CreatePasswordHash(password)
	if err != nil {
		return false, err
	}

	updateTx, err := r.db.Begin()
	if err != nil {
		return false, fmt.Errorf("could not update %s password: %w", email, err)
	}
	defer func() { _ = updateTx.Rollback() }() //nolint:wsl

	now := time.Now()

	// Insert or update record in the database
	stmt, err := updateTx.Prepare("UPDATE users SET password = $1, updated_at = $2, last_seen_at = $3 WHERE email == $4")
	if err != nil {
		return false, fmt.Errorf("could not update %s password: %w", email, err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(hash, now.Unix(), now.Unix(), email)
	if err != nil {
		return false, fmt.Errorf("could not update %s password: %w", email, err)
	}

	err = updateTx.Commit()
	if err != nil {
		return false, fmt.Errorf("could not update %s password: %w", email, err)
	}

	rowsAffected, _ := result.RowsAffected()

	return rowsAffected >= 1, nil
}

// TouchUser updates user's last_seen_at value with current Unix time.
func (r *UserRepository) TouchUser(email string) error {
	now := time.Now()

	updateTx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("could not update %s last_seen_at: %w", email, err)
	}
	defer func() { _ = updateTx.Rollback() }() //nolint:wsl

	update, err := updateTx.Prepare("UPDATE users SET last_seen_at = $1 WHERE email == $2")
	if err != nil {
		return fmt.Errorf("could not update %s last_seen_at: %w", email, err)
	}
	defer update.Close()

	result, err := update.Exec(now.Unix(), email)
	if err != nil {
		return fmt.Errorf("could not update %s last_seen_at: %w", email, err)
	}

	err = updateTx.Commit()
	if err != nil {
		return fmt.Errorf("could not update %s last_seen_at: %w", email, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not update %s last_seen_at: %w", email, err)
	}

	if rowsAffected != 1 {
		return ErrUserNotFound
	}

	return nil
}

// AuthenticateUser validates if the email and password combination matches an existing user
// in the database with a password resulting in the same password hash.
func (r *UserRepository) AuthenticateUser(email string, password string) (bool, error) {
	user, err := r.UserWithEmail(email)
	if err != nil {
		return false, err
	}

	return user.PasswordMatch(password), nil
}
