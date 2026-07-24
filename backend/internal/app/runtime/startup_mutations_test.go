package runtime

import (
	"errors"
	"testing"

	configsvc "socialpredict/internal/service/config"

	"gorm.io/gorm"
)

func TestRunStartupMutationsWriterRunsMigrateAndSeeds(t *testing.T) {
	calls := []string{}
	hooks := StartupMutationHooks{
		Migrate:      func(*gorm.DB) error { calls = append(calls, "migrate"); return nil },
		Verify:       func(*gorm.DB) error { calls = append(calls, "verify"); return nil },
		SeedUsers:    func(*gorm.DB, configsvc.Service) error { calls = append(calls, "seedUsers"); return nil },
		SeedHomepage: func(*gorm.DB, string) error { calls = append(calls, "seedHomepage"); return nil },
	}

	err := RunStartupMutations(nil, configsvc.NewStaticService(nil), StartupMutationMode{Writer: true}, hooks)
	if err != nil {
		t.Fatalf("RunStartupMutations returned error: %v", err)
	}
	want := []string{"migrate", "seedUsers", "seedHomepage"}
	if len(calls) != len(want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("calls = %#v, want %#v", calls, want)
		}
	}
}

func TestRunStartupMutationsReaderVerifiesOnly(t *testing.T) {
	calls := []string{}
	hooks := StartupMutationHooks{
		Migrate:      func(*gorm.DB) error { calls = append(calls, "migrate"); return nil },
		Verify:       func(*gorm.DB) error { calls = append(calls, "verify"); return nil },
		SeedUsers:    func(*gorm.DB, configsvc.Service) error { calls = append(calls, "seedUsers"); return nil },
		SeedHomepage: func(*gorm.DB, string) error { calls = append(calls, "seedHomepage"); return nil },
	}

	err := RunStartupMutations(nil, configsvc.NewStaticService(nil), StartupMutationMode{Writer: false}, hooks)
	if err != nil {
		t.Fatalf("RunStartupMutations returned error: %v", err)
	}
	if len(calls) != 1 || calls[0] != "verify" {
		t.Fatalf("calls = %#v, want verify only", calls)
	}
}

func TestRunStartupMutationsWrapsHookErrors(t *testing.T) {
	hooks := StartupMutationHooks{Verify: func(*gorm.DB) error { return errors.New("boom") }}
	err := RunStartupMutations(nil, configsvc.NewStaticService(nil), StartupMutationMode{}, hooks)
	if err == nil {
		t.Fatalf("RunStartupMutations returned nil error")
	}
}
