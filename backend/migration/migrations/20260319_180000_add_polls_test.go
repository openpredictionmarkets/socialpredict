package migrations_test

import (
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestAddPollsMigration_TablesExist(t *testing.T) {
	db := modelstesting.NewFakeDB(t)

	m := db.Migrator()

	if !m.HasTable(&models.Poll{}) {
		t.Fatal("expected polls table to exist after migration")
	}
	if !m.HasTable(&models.PollVote{}) {
		t.Fatal("expected poll_votes table to exist after migration")
	}

	for _, col := range []string{"CreatorUsername", "Question", "IsClosed"} {
		if !m.HasColumn(&models.Poll{}, col) {
			t.Fatalf("expected polls.%s column to exist", col)
		}
	}
	for _, col := range []string{"PollID", "Username", "Vote"} {
		if !m.HasColumn(&models.PollVote{}, col) {
			t.Fatalf("expected poll_votes.%s column to exist", col)
		}
	}
}
