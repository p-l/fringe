package mocks

import (
	"os"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/jmoiron/sqlx"
	"github.com/p-l/fringe/internal/repos"
	"modernc.org/ql"
)

// NewMockUserRepository returns an actual repos.UserRepository setup with fake data in a temporary directory.
func NewMockUserRepository(t *testing.T) *repos.UserRepository {
	t.Helper()

	fake := faker.New()

	tempDir, err := os.MkdirTemp("", "fringe_test_*")
	if err != nil {
		t.Fatalf("NewMockUserRepository: Could not create temp directory for ")
	}

	// Initialize Database connexion
	ql.RegisterDriver()

	connexion, err := sqlx.Open("ql", tempDir+"/db")
	if err != nil {
		t.Fatalf("NewMockUserRepository: could not connect to database: %v", err)
	}

	userRepo, err := repos.NewUserRepository(connexion)
	if err != nil {
		t.Fatalf("NewMockUserRepository: Could not initate user repository: %v", err)
	}

	// Create 5 fake users
	for i := 0; i < 5; i++ {
		_, err = userRepo.CreateUser(fake.Internet().Email(), fake.Internet().Password())
		if err != nil {
			t.Fatalf("NewMockUserRepository: Could not initate user repository: %v", err)
		}
	}

	t.Cleanup(func() {
		connexion.Close()
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Fatalf("NewMockUserRepository.Cleanup could clean up temp database (%s): %v", tempDir, err)
		}
	})

	return userRepo
}
