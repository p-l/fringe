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
	Email             string `db:"email"`
	Name              string `db:"name"`
	Picture           string `db:"picture"`
	PasswordHash      string `db:"password"`
	CreatedAt         int64  `db:"created_at"`
	ProfileUpdatedAt  int64  `db:"profile_updated_at"`
	PasswordUpdatedAt int64  `db:"password_updated_at"`
	LastSeenAt        int64  `db:"last_seen_at"`
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
		"name string NOT NULL, " +
		"picture string," +
		"created_at int64," +
		"profile_updated_at int64," +
		"password_updated_at int64," +
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

// FindByEmail Looks for a user record that matches the provided email and returns a pointer to User stuck if found or nil if not.
func (r *UserRepository) FindByEmail(email string) (*User, error) {
	var user User

	if err := r.db.Get(&user, "SELECT * FROM users WHERE email == $1 LIMIT 1", email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("could not retrieve user %s: %w", email, err)
	}

	return &user, nil
}

// Create INSERT a new user record with email and argon2id password hash from the provided password.
// If the user already exists returns the record from the Database, otherwise return the newly created User.
func (r *UserRepository) Create(email string, name string, picture string, password string) (*User, error) {
	dbUser, err := r.FindByEmail(email)
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
	newUser := User{
		Email:             email,
		Name:              name,
		Picture:           picture,
		PasswordHash:      hash,
		LastSeenAt:        now.Unix(),
		CreatedAt:         now.Unix(),
		ProfileUpdatedAt:  now.Unix(),
		PasswordUpdatedAt: now.Unix(),
	}

	insertTx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("could not create user %s: %w", email, err)
	}
	defer func() { _ = insertTx.Rollback() }() //nolint:wsl

	// Insert record in the database
	insert, err := insertTx.Prepare("INSERT INTO users (email, name, picture, password, created_at, profile_updated_at, password_updated_at, last_seen_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)")
	if err != nil {
		return nil, fmt.Errorf("could not create user %s: %w", email, err)
	}
	defer insert.Close()

	_, err = insert.Exec(newUser.Email, newUser.Name, newUser.Picture, newUser.PasswordHash, newUser.CreatedAt, newUser.ProfileUpdatedAt, newUser.PasswordUpdatedAt, newUser.LastSeenAt)
	if err != nil {
		return nil, fmt.Errorf("could not create user %s: %w", email, err)
	}

	if err := insertTx.Commit(); err != nil {
		return nil, fmt.Errorf("could not create user %s: %w", email, err)
	}

	return &newUser, nil
}

// UpdatePassword replaces the specified user's (found by email address) by the password provided.
// The password is not stored as is. It is hashed with argon2id.
func (r *UserRepository) UpdatePassword(email string, password string) (updated bool, err error) {
	hash, err := CreatePasswordHash(password)
	if err != nil {
		return false, err
	}

	if !r.Exists(email) {
		return false, ErrUserNotFound
	}

	updateTx, err := r.db.Begin()
	if err != nil {
		return false, fmt.Errorf("could not update %s password: %w", email, err)
	}
	defer func() { _ = updateTx.Rollback() }() //nolint:wsl

	now := time.Now()

	// Insert or update record in the database
	stmt, err := updateTx.Prepare("UPDATE users SET password = $1, password_updated_at = $2 WHERE email == $3")
	if err != nil {
		return false, fmt.Errorf("could not update %s password: %w", email, err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(hash, now.Unix(), email)
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

func (r *UserRepository) UpdateProfile(email string, name string, picture string) (bool, error) {
	log.Printf("Update Profile: Email:%s Name:%s Picture:%s", email, name, picture)

	if !r.Exists(email) {
		return false, ErrUserNotFound
	}

	updateTx, err := r.db.Begin()
	if err != nil {
		return false, fmt.Errorf("could not update %s profile information: %w", email, err)
	}
	defer func() { _ = updateTx.Rollback() }() //nolint:wsl

	now := time.Now()

	// Insert or update record in the database
	stmt, err := updateTx.Prepare("UPDATE users SET name = '$1', picture = '$2', profile_updated_at = $3 WHERE email == $4")
	if err != nil {
		return false, fmt.Errorf("could not update %s profile information: %w", email, err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(name, picture, now.Unix(), email)
	if err != nil {
		return false, fmt.Errorf("could not update %s profile information: %w", email, err)
	}

	err = updateTx.Commit()
	if err != nil {
		return false, fmt.Errorf("could not update %s profile information: %w", email, err)
	}

	rowsAffected, _ := result.RowsAffected()

	return rowsAffected >= 1, nil
}

func (r *UserRepository) Exists(email string) bool {
	_, err := r.FindByEmail(email)

	// if err == null then user is present
	return err == nil
}

// Seen updates user's last_seen_at value with current Unix time.
func (r *UserRepository) Seen(email string) error {
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

// Authenticate validates if the email and password combination matches an existing user
// in the database with a password resulting in the same password hash.
// Updates last_seen_at if user is authenticated.
func (r *UserRepository) Authenticate(email string, password string) (bool, error) {
	user, err := r.FindByEmail(email)
	if err != nil {
		return false, err
	}

	authenticated := user.PasswordMatch(password)
	if authenticated {
		err = r.Seen(email)
		if err != nil {
			return false, err
		}
	}

	return authenticated, nil
}

const UserRepositoryListMaxLimit = 100

// AllUsers Return list of users sorted by email.
// Passing 0 as the limit will use UserRepositoryListMaxLimit as the limit.
// Page 0 and 1 are seen as the same page number essentially: `offset = (page - 1) * limit`.
func (r *UserRepository) AllUsers(limit int, page int) ([]User, error) {
	offset := 0

	if limit == 0 || limit > UserRepositoryListMaxLimit {
		limit = UserRepositoryListMaxLimit
	}

	if page > 1 {
		offset = (page - 1) * limit
	}

	var users []User

	err := r.db.Select(&users, "SELECT * FROM users ORDER BY email LIMIT $1 OFFSET $2 ", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve users (limit: %d offset:%d) %w", limit, offset, err)
	}

	if users == nil {
		return nil, ErrUserNotFound
	}

	return users, nil
}

// Delete delete user record for given email.
func (r *UserRepository) Delete(email string) error {
	delTx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("could not delte %s: %w", email, err)
	}
	defer func() { _ = delTx.Rollback() }() //nolint:wsl

	delStmt, err := delTx.Prepare("DELETE FROM users WHERE email == $1")
	if err != nil {
		return fmt.Errorf("could not delete %s: %w ", email, err)
	}
	defer delStmt.Close()

	result, err := delStmt.Exec(email)
	if err != nil {
		return fmt.Errorf("could not delete %s: %w", email, err)
	}

	err = delTx.Commit()
	if err != nil {
		return fmt.Errorf("could not delete %s: %w", email, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not delete %s: %w", email, err)
	}

	if rowsAffected != 1 {
		return ErrUserNotFound
	}

	return nil
}
