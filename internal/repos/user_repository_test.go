package repos_test

import (
	"log"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jaswdr/faker"
	"github.com/jmoiron/sqlx"
	"github.com/p-l/fringe/internal/repos"
	"github.com/stretchr/testify/assert"
)

func userTableColumns() []string {
	return []string{"email", "name", "picture", "password", "created_at", "profile_updated_at", "password_updated_at", "last_seen_at"}
}

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

func TestUserRepository_CreateUser(t *testing.T) {
	t.Parallel()

	t.Run("Returns new user when creating", func(t *testing.T) {
		t.Parallel()

		db, mockSQL := dbOpen()
		defer db.Close()

		mockSQL.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(userTableColumns()))

		mockSQL.ExpectBegin()
		mockSQL.ExpectPrepare("INSERT INTO users").WillBeClosed()
		mockSQL.ExpectExec("").WillReturnResult(sqlmock.NewResult(1, 1))
		mockSQL.ExpectCommit()

		userRepo, _ := repos.NewUserRepository(db)
		fake := faker.New()

		email := fake.Internet().Email()
		password := fake.Internet().Password()
		name := fake.Person().Name()
		picture := fake.Internet().URL()

		user, err := userRepo.Create(email, name, picture, password)
		if err != nil {
			t.Errorf("Creating user failed: %v", err)
		}

		// Create without an existing user must return a valid user with its hashed password
		// The created user will have the provided password hash and will validate
		assert.NotNil(t, user)

		if user != nil {
			assert.Equal(t, email, user.Email)
			assert.Equal(t, name, user.Name)
			assert.Equal(t, picture, user.Picture)
			assert.NotEqual(t, password, user.PasswordHash)
			assert.True(t, user.PasswordMatch(password))
		}

		// we make sure that all expectations were met
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Don't change record when trying to create existing user", func(t *testing.T) {
		t.Parallel()

		db, mockSQL := dbOpen()
		defer db.Close()

		fake := faker.New()
		email := fake.Internet().Email()
		name := fake.Person().Name()
		picture := fake.Internet().URL()
		password := fake.Internet().Password()
		passwordHash, _ := repos.CreatePasswordHash(password)
		createdAt := fake.Time().Unix(time.Now())
		profileUpdatedAt := fake.Time().Unix(time.Now())
		passwordUpdatedAt := fake.Time().Unix(time.Now())
		lastSeenAt := fake.Time().Unix(time.Now())

		newPassword := fake.Internet().Password()

		// Return the fake User
		mockSQL.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows(userTableColumns()).AddRow(email, name, picture, passwordHash, createdAt, profileUpdatedAt, passwordUpdatedAt, lastSeenAt))

		userRepo, _ := repos.NewUserRepository(db)

		user, err := userRepo.Create(email, name, picture, newPassword)
		assert.ErrorIs(t, err, repos.ErrUserAlreadyExist)

		// CreatUser will return the existing user without modifying it
		assert.NotNil(t, user)

		if user != nil {
			assert.Equal(t, email, user.Email)
			// The retrieved user is expected to contain the fake password Hash and match
			assert.True(t, user.PasswordMatch(password))
			assert.Equal(t, createdAt, user.CreatedAt)
			assert.Equal(t, profileUpdatedAt, user.ProfileUpdatedAt)
			assert.Equal(t, passwordUpdatedAt, user.PasswordUpdatedAt)
			assert.Equal(t, lastSeenAt, user.LastSeenAt)
		}

		// we make sure that all expectations were met
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestUserRepository_UserWithEmail(t *testing.T) {
	t.Parallel()

	t.Run("Returns unmodified user when found", func(t *testing.T) {
		t.Parallel()

		db, mockSQL := dbOpen()
		defer db.Close()

		fake := faker.New()
		email := fake.Internet().Email()
		name := fake.Person().Name()
		picture := fake.Internet().URL()
		password := fake.Internet().Password()
		passwordHash, _ := repos.CreatePasswordHash(password)
		createdAt := fake.Time().Unix(time.Now())
		profileUpdatedAt := fake.Time().Unix(time.Now())
		passwordUpdatedAt := fake.Time().Unix(time.Now())
		lastSeenAt := fake.Time().Unix(time.Now())

		// Return the fake User
		mockSQL.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows(userTableColumns()).AddRow(email, name, picture, passwordHash, createdAt, profileUpdatedAt, passwordUpdatedAt, lastSeenAt))

		userRepo, _ := repos.NewUserRepository(db)

		user, err := userRepo.FindByEmail(email)
		assert.Nil(t, err)
		assert.NotNil(t, user)

		if user != nil {
			assert.Equal(t, email, user.Email)
			assert.Equal(t, name, user.Name)
			assert.Equal(t, picture, user.Picture)
			assert.Equal(t, passwordHash, user.PasswordHash)
			assert.Equal(t, createdAt, user.CreatedAt)
			assert.Equal(t, profileUpdatedAt, user.ProfileUpdatedAt)
			assert.Equal(t, passwordUpdatedAt, user.PasswordUpdatedAt)
			assert.Equal(t, lastSeenAt, user.LastSeenAt)
		}

		// we make sure that all expectations were met
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Returns error when user is not found", func(t *testing.T) {
		t.Parallel()

		db, mockSQL := dbOpen()
		defer db.Close()

		fake := faker.New()
		email := fake.Internet().Email()

		mockSQL.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows(userTableColumns()))

		userRepo, _ := repos.NewUserRepository(db)

		user, err := userRepo.FindByEmail(email)
		assert.ErrorIs(t, err, repos.ErrUserNotFound)
		assert.Nil(t, user)

		// we make sure that all expectations were met
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestUserRepository_UpdateUserPassword(t *testing.T) {
	t.Parallel()
	t.Run("Fails with unknown user", func(t *testing.T) {
		t.Parallel()

		db, mockSQL := dbOpen()
		defer db.Close()

		fake := faker.New()
		email := fake.Internet().Email()

		newPassword := fake.Internet().Password()

		mockSQL.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows(userTableColumns()))

		userRepo, _ := repos.NewUserRepository(db)

		success, err := userRepo.UpdatePassword(email, newPassword)
		assert.NotNil(t, err)
		assert.False(t, success)

		// we make sure that all expectations were met
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Updates known user", func(t *testing.T) {
		t.Parallel()

		db, mockSQL := dbOpen()
		defer db.Close()

		fake := faker.New()
		email := fake.Internet().Email()
		name := fake.Person().Name()
		picture := fake.Internet().URL()
		passwordHash, _ := repos.CreatePasswordHash(fake.Internet().Password())
		createdAt := fake.Time().Unix(time.Now())
		profileUpdatedAt := fake.Time().Unix(time.Now())
		passwordUpdatedAt := fake.Time().Unix(time.Now())
		lastSeenAt := fake.Time().Unix(time.Now())
		newPassword := fake.Internet().Password()

		mockSQL.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows(userTableColumns()).AddRow(email, name, picture, passwordHash, createdAt, profileUpdatedAt, passwordUpdatedAt, lastSeenAt))
		mockSQL.ExpectBegin()
		// Ensure that password, updated_at AND last_seen_at are updated
		mockSQL.ExpectPrepare("UPDATE users SET password = .*, password_updated_at = .* WHERE").WillBeClosed()
		mockSQL.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
		mockSQL.ExpectCommit()

		userRepo, _ := repos.NewUserRepository(db)

		success, err := userRepo.UpdatePassword(email, newPassword)
		assert.Nil(t, err)
		assert.True(t, success)

		// we make sure that all expectations were met
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestUserRepository_TouchUser(t *testing.T) {
	t.Parallel()
	t.Run("Fails with unknown user", func(t *testing.T) {
		t.Parallel()

		db, mockSQL := dbOpen()
		defer db.Close()

		fake := faker.New()
		email := fake.Internet().Email()

		mockSQL.ExpectBegin()
		mockSQL.ExpectPrepare("UPDATE users").WillBeClosed()
		mockSQL.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 0))
		mockSQL.ExpectCommit()

		userRepo, _ := repos.NewUserRepository(db)

		err := userRepo.Seen(email)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, repos.ErrUserNotFound)

		// we make sure that all expectations were met
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Updates only known user", func(t *testing.T) {
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

		userRepo, _ := repos.NewUserRepository(db)

		err := userRepo.Seen(email)
		assert.Nil(t, err)

		// we make sure that all expectations were met
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestUserRepository_AuthenticateUser(t *testing.T) {
	t.Parallel()
	t.Run("Returns UserNotFound when not found", func(t *testing.T) {
		t.Parallel()

		db, mockSQL := dbOpen()
		defer db.Close()

		fake := faker.New()
		email := fake.Internet().Email()
		password := fake.Internet().Password()

		// Return the fake User
		mockSQL.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(userTableColumns()))

		userRepo, _ := repos.NewUserRepository(db)

		success, err := userRepo.Authenticate(email, password)
		assert.ErrorIs(t, err, repos.ErrUserNotFound)
		assert.False(t, success)

		// we make sure that all expectations were met
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("User with valid hash authenticates and update last seen", func(t *testing.T) {
		t.Parallel()

		db, mockSQL := dbOpen()
		defer db.Close()

		fake := faker.New()
		email := fake.Internet().Email()
		name := fake.Person().Name()
		picture := fake.Internet().URL()
		password := fake.Internet().Password()
		passwordHash, _ := repos.CreatePasswordHash(password)
		createdAt := fake.Time().Unix(time.Now())
		profileUpdatedAt := fake.Time().Unix(time.Now())
		passwordUpdatedAt := fake.Time().Unix(time.Now())
		lastSeenAt := fake.Time().Unix(time.Now())

		// Return the fake User
		mockSQL.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows(userTableColumns()).AddRow(email, name, picture, passwordHash, createdAt, profileUpdatedAt, passwordUpdatedAt, lastSeenAt))
		mockSQL.ExpectBegin()
		// Ensures target "last_seen_at" column
		mockSQL.ExpectPrepare("UPDATE users SET last_seen_at = .* WHERE").WillBeClosed()
		mockSQL.ExpectExec("").WillReturnResult(sqlmock.NewResult(0, 1))
		mockSQL.ExpectCommit()

		userRepo, _ := repos.NewUserRepository(db)

		success, err := userRepo.Authenticate(email, password)
		assert.Nil(t, err)
		assert.True(t, success)

		// we make sure that all expectations were met
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestUserRepository_AllUsers(t *testing.T) {
	t.Parallel()
	t.Run("Return error when no users are in the db", func(t *testing.T) {
		t.Parallel()

		db, mockSQL := dbOpen()
		defer db.Close()

		// Return the fake User
		mockSQL.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows(userTableColumns()))

		userRepo, err := repos.NewUserRepository(db)
		assert.NoError(t, err)

		users, err := userRepo.AllUsers(0, 0)
		assert.ErrorIs(t, err, repos.ErrUserNotFound)
		assert.Nil(t, users)

		// we make sure that all expectations were met
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Return all users if there are more than users db", func(t *testing.T) {
		t.Parallel()

		db, mockSQL := dbOpen()
		defer db.Close()

		fake := faker.New()
		email := fake.Internet().Email()
		name := fake.Person().Name()
		picture := fake.Internet().URL()
		password := fake.Internet().Password()
		passwordHash, _ := repos.CreatePasswordHash(password)
		createdAt := fake.Time().Unix(time.Now())
		profileUpdatedAt := fake.Time().Unix(time.Now())
		passwordUpdatedAt := fake.Time().Unix(time.Now())
		lastSeenAt := fake.Time().Unix(time.Now())

		// Return the fake User
		mockSQL.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows(userTableColumns()).AddRow(email, name, picture, passwordHash, createdAt, profileUpdatedAt, passwordUpdatedAt, lastSeenAt))

		userRepo, err := repos.NewUserRepository(db)
		assert.NoError(t, err)

		users, err := userRepo.AllUsers(0, 0)
		assert.NoError(t, err)
		assert.NotNil(t, users)
		assert.Len(t, users, 1)
		assert.Equal(t, email, users[0].Email)

		// we make sure that all expectations were met
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("follow request limit", func(t *testing.T) {
		t.Parallel()

		db, mockSQL := dbOpen()
		defer db.Close()

		fake := faker.New()
		createdAt := fake.Time().Unix(time.Now())
		profileUpdatedAt := fake.Time().Unix(time.Now())
		passwordUpdatedAt := fake.Time().Unix(time.Now())
		lastSeenAt := fake.Time().Unix(time.Now())

		// BD with 3 users
		mockRows := sqlmock.NewRows(userTableColumns())
		mockRows.AddRow(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), fake.Internet().Password(), createdAt, profileUpdatedAt, passwordUpdatedAt, lastSeenAt)
		mockRows.AddRow(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), fake.Internet().Password(), createdAt, profileUpdatedAt, passwordUpdatedAt, lastSeenAt)
		mockSQL.ExpectQuery("SELECT").WithArgs(2, 0).WillReturnRows(mockRows)

		userRepo, err := repos.NewUserRepository(db)
		assert.NoError(t, err)

		users, err := userRepo.AllUsers(2, 0)
		assert.NoError(t, err)
		assert.NotNil(t, users)
		assert.Len(t, users, 2)

		// we make sure that all expectations were met
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("page 0 and 1 are the same", func(t *testing.T) {
		t.Parallel()

		db, mockSQL := dbOpen()
		defer db.Close()

		fake := faker.New()
		createdAt := fake.Time().Unix(time.Now())
		profileUpdatedAt := fake.Time().Unix(time.Now())
		passwordUpdatedAt := fake.Time().Unix(time.Now())
		lastSeenAt := fake.Time().Unix(time.Now())

		firstQueryRows := sqlmock.NewRows(userTableColumns())
		firstQueryRows.AddRow(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), fake.Internet().Password(), createdAt, profileUpdatedAt, passwordUpdatedAt, lastSeenAt)
		firstQueryRows.AddRow(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), fake.Internet().Password(), createdAt, profileUpdatedAt, passwordUpdatedAt, lastSeenAt)
		mockSQL.ExpectQuery("SELECT").WithArgs(2, 0).WillReturnRows(firstQueryRows)

		userRepo, err := repos.NewUserRepository(db)
		assert.NoError(t, err)

		users, err := userRepo.AllUsers(2, 1)
		assert.NoError(t, err)
		assert.NotNil(t, users)
		assert.Len(t, users, 2)

		secondQueryRows := sqlmock.NewRows(userTableColumns())
		secondQueryRows.AddRow(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), fake.Internet().Password(), createdAt, profileUpdatedAt, passwordUpdatedAt, lastSeenAt)
		secondQueryRows.AddRow(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), fake.Internet().Password(), createdAt, profileUpdatedAt, passwordUpdatedAt, lastSeenAt)
		mockSQL.ExpectQuery("SELECT").WithArgs(2, 0).WillReturnRows(secondQueryRows)
		users, err = userRepo.AllUsers(2, 0)
		assert.NoError(t, err)
		assert.NotNil(t, users)
		assert.Len(t, users, 2)

		// we make sure that all expectations were met
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
