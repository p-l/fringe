package repositories_test

import (
	"log"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jaswdr/faker"
	"github.com/jmoiron/sqlx"
	"github.com/p-l/fringe/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func dbOpen() (*sqlx.DB, sqlmock.Sqlmock) {
	mockDB, mockSQL, err := sqlmock.New()
	if err != nil {
		log.Panicf("FATAL: an error '%s' was not expected when opening a stub database connection", err)
	}

	db := sqlx.NewDb(mockDB, "sqlmock")

	mockSQL.ExpectBegin()
	mockSQL.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(1, 1))
	mockSQL.ExpectExec("CREATE INDEX").WillReturnResult(sqlmock.NewResult(1, 1))
	mockSQL.ExpectCommit()

	return db, mockSQL
}

func TestShouldReturnNewUserWhenCreating(t *testing.T) {
	t.Parallel()

	db, mockSQL := dbOpen()
	defer db.Close()

	mockSQL.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"email", "password", "created_at", "updated_at", "last_seen_at"}))

	mockSQL.ExpectBegin()
	mockSQL.ExpectPrepare("INSERT INTO users").WillBeClosed()
	mockSQL.ExpectExec("").WillReturnResult(sqlmock.NewResult(1, 1))
	mockSQL.ExpectCommit()

	userRepo, _ := repositories.NewUserRepository(db)
	fake := faker.New()

	email := fake.Internet().Email()
	password := fake.Internet().Password()

	user, err := userRepo.CreateUser(email, password)
	if err != nil {
		t.Errorf("Creating user failed: %v", err)
	}

	// CreateUser without an existing user must return a valid user with its hashed password
	// The created user will have the provided password hash and will validate
	assert.NotNil(t, user)

	if user != nil {
		assert.Equal(t, email, user.Email)
		assert.NotEqual(t, password, user.PasswordHash)
		assert.True(t, user.PasswordMatch(password))
	}

	// we make sure that all expectations were met
	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateOnlyLastSeenWhenCreatingExistingUser(t *testing.T) {
	t.Parallel()

	db, mockSQL := dbOpen()
	defer db.Close()

	fake := faker.New()
	email := fake.Internet().Email()
	password := fake.Internet().Password()
	passwordHash, _ := repositories.CreatePasswordHash(password)
	createdAt := fake.Time().Unix(time.Now())
	updatedAt := fake.Time().Unix(time.Now())
	lastSeenAt := fake.Time().Unix(time.Now())

	newPassword := fake.Internet().Password()

	// Return the fake User
	mockSQL.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"email", "password", "created_at", "updated_at", "last_seen_at"}).AddRow(email, passwordHash, createdAt, updatedAt, lastSeenAt))

	userRepo, _ := repositories.NewUserRepository(db)

	user, err := userRepo.CreateUser(email, newPassword)
	assert.ErrorIs(t, err, repositories.ErrUserAlreadyExist)

	// CreatUser will return the existing user without modifying it
	assert.NotNil(t, user)

	if user != nil {
		assert.Equal(t, email, user.Email)
		// The retrieved user is expected to contain the fake password Hash and match
		assert.True(t, user.PasswordMatch(password))
		assert.Equal(t, createdAt, user.CreatedAt)
		assert.Equal(t, updatedAt, user.UpdatedAt)
		assert.Equal(t, lastSeenAt, user.LastSeenAt)
	}

	// we make sure that all expectations were met
	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUserWithEmailReturnsUnmodifiedUserWhenFound(t *testing.T) {
	t.Parallel()

	db, mockSQL := dbOpen()
	defer db.Close()

	fake := faker.New()
	email := fake.Internet().Email()
	password := fake.Internet().Password()
	passwordHash, _ := repositories.CreatePasswordHash(password)
	createdAt := fake.Time().Unix(time.Now())
	updatedAt := fake.Time().Unix(time.Now())
	lastSeenAt := fake.Time().Unix(time.Now())

	// Return the fake User
	mockSQL.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"email", "password", "created_at", "updated_at", "last_seen_at"}).AddRow(email, passwordHash, createdAt, updatedAt, lastSeenAt))

	userRepo, _ := repositories.NewUserRepository(db)

	user, err := userRepo.UserWithEmail(email)
	assert.Nil(t, err)
	assert.NotNil(t, user)

	if user != nil {
		assert.Equal(t, email, user.Email)
		assert.Equal(t, passwordHash, user.PasswordHash)
		assert.Equal(t, createdAt, user.CreatedAt)
		assert.Equal(t, updatedAt, user.UpdatedAt)
		assert.Equal(t, lastSeenAt, user.LastSeenAt)
	}

	// we make sure that all expectations were met
	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUserWithEmailReturnsErrorWhenUserNotFound(t *testing.T) {
	t.Parallel()

	db, mockSQL := dbOpen()
	defer db.Close()

	fake := faker.New()
	email := fake.Internet().Email()

	// Return the fake User
	mockSQL.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"email", "password", "created_at", "updated_at", "last_seen_at"}))

	userRepo, _ := repositories.NewUserRepository(db)

	user, err := userRepo.UserWithEmail(email)
	assert.ErrorIs(t, err, repositories.ErrUserNotFound)
	assert.Nil(t, user)

	// we make sure that all expectations were met
	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateUserPasswordFailsWithUnknownUser(t *testing.T) {
	t.Parallel()

	db, mockSQL := dbOpen()
	defer db.Close()

	fake := faker.New()
	email := fake.Internet().Email()
	newPassword := fake.Internet().Password()

	mockSQL.ExpectBegin()
	mockSQL.ExpectPrepare("UPDATE users").WillBeClosed()
	mockSQL.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 0))
	mockSQL.ExpectCommit()

	userRepo, _ := repositories.NewUserRepository(db)

	success, err := userRepo.UpdateUserPassword(email, newPassword)
	assert.Nil(t, err)
	assert.False(t, success)

	// we make sure that all expectations were met
	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateUserPasswordUpdatesKnownUser(t *testing.T) {
	t.Parallel()

	db, mockSQL := dbOpen()
	defer db.Close()

	fake := faker.New()
	email := fake.Internet().Email()
	newPassword := fake.Internet().Password()

	mockSQL.ExpectBegin()
	// Ensure that password, updated_at AND last_seen_at are updated
	mockSQL.ExpectPrepare("UPDATE users SET password = .*, updated_at = .*, last_seen_at = .* WHERE").WillBeClosed()
	mockSQL.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
	mockSQL.ExpectCommit()

	userRepo, _ := repositories.NewUserRepository(db)

	success, err := userRepo.UpdateUserPassword(email, newPassword)
	assert.Nil(t, err)
	assert.True(t, success)

	// we make sure that all expectations were met
	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestTouchUserFailsWithUnknownUser(t *testing.T) {
	t.Parallel()

	db, mockSQL := dbOpen()
	defer db.Close()

	fake := faker.New()
	email := fake.Internet().Email()

	mockSQL.ExpectBegin()
	mockSQL.ExpectPrepare("UPDATE users").WillBeClosed()
	mockSQL.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 0))
	mockSQL.ExpectCommit()

	userRepo, _ := repositories.NewUserRepository(db)

	err := userRepo.TouchUser(email)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, repositories.ErrUserNotFound)

	// we make sure that all expectations were met
	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestTouchUserUpdatesOnlyKnownUser(t *testing.T) {
	t.Parallel()

	db, mockSQL := dbOpen()
	defer db.Close()

	fake := faker.New()
	email := fake.Internet().Email()

	mockSQL.ExpectBegin()
	// Ensures target "last_seen_at" column
	mockSQL.ExpectPrepare("UPDATE users SET last_seen_at = .* WHERE").WillBeClosed()
	mockSQL.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
	mockSQL.ExpectCommit()

	userRepo, _ := repositories.NewUserRepository(db)

	err := userRepo.TouchUser(email)
	assert.Nil(t, err)

	// we make sure that all expectations were met
	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAuthenticateReturnsUserNotFoundWhenNotFound(t *testing.T) {
	t.Parallel()

	db, mockSQL := dbOpen()
	defer db.Close()

	fake := faker.New()
	email := fake.Internet().Email()
	password := fake.Internet().Password()

	// Return the fake User
	mockSQL.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"email", "password", "created_at", "updated_at", "last_seen_at"}))

	userRepo, _ := repositories.NewUserRepository(db)

	success, err := userRepo.AuthenticateUser(email, password)
	assert.ErrorIs(t, err, repositories.ErrUserNotFound)
	assert.False(t, success)

	// we make sure that all expectations were met
	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAuthenticateUserWithHAshWhenUserExists(t *testing.T) {
	t.Parallel()

	db, mockSQL := dbOpen()
	defer db.Close()

	fake := faker.New()
	email := fake.Internet().Email()
	password := fake.Internet().Password()
	passwordHash, _ := repositories.CreatePasswordHash(password)
	createdAt := fake.Time().Unix(time.Now())
	updatedAt := fake.Time().Unix(time.Now())
	lastSeenAt := fake.Time().Unix(time.Now())

	// Return the fake User
	mockSQL.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"email", "password", "created_at", "updated_at", "last_seen_at"}).AddRow(email, passwordHash, createdAt, updatedAt, lastSeenAt))

	userRepo, _ := repositories.NewUserRepository(db)

	success, err := userRepo.AuthenticateUser(email, password)
	assert.Nil(t, err)
	assert.True(t, success)

	// we make sure that all expectations were met
	if err := mockSQL.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}