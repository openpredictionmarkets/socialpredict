package main

import (
	"errors"
	"reflect"
	"testing"

	appruntime "socialpredict/internal/app/runtime"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/logger"

	"gorm.io/gorm"
)

func TestRunStartupMutationsWriterRunsMigrationsThenSeeds(t *testing.T) {
	var calls []string
	hooks := startupMutationHooks{
		migrate: func(*gorm.DB) error {
			calls = append(calls, "migrate")
			return nil
		},
		verify: func(*gorm.DB) error {
			calls = append(calls, "verify")
			return nil
		},
		seedUsers: func(*gorm.DB, configsvc.Service) error {
			calls = append(calls, "seedUsers")
			return nil
		},
		seedHomepage: func(*gorm.DB, string) error {
			calls = append(calls, "seedHomepage")
			return nil
		},
	}

	err := runStartupMutations(nil, nil, appruntime.StartupMutationMode{Writer: true}, hooks)
	if err != nil {
		t.Fatalf("runStartupMutations returned error: %v", err)
	}

	want := []string{"migrate", "seedUsers", "seedHomepage"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestRunStartupMutationsNonWriterOnlyVerifies(t *testing.T) {
	var calls []string
	hooks := startupMutationHooks{
		migrate: func(*gorm.DB) error {
			calls = append(calls, "migrate")
			return nil
		},
		verify: func(*gorm.DB) error {
			calls = append(calls, "verify")
			return nil
		},
		seedUsers: func(*gorm.DB, configsvc.Service) error {
			calls = append(calls, "seedUsers")
			return nil
		},
		seedHomepage: func(*gorm.DB, string) error {
			calls = append(calls, "seedHomepage")
			return nil
		},
	}

	err := runStartupMutations(nil, nil, appruntime.StartupMutationMode{}, hooks)
	if err != nil {
		t.Fatalf("runStartupMutations returned error: %v", err)
	}

	want := []string{"verify"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestRunStartupMutationsWriterStopsBeforeSeedsOnMigrationFailure(t *testing.T) {
	migrationErr := errors.New("migration failed")
	var calls []string
	hooks := startupMutationHooks{
		migrate: func(*gorm.DB) error {
			calls = append(calls, "migrate")
			return migrationErr
		},
		seedUsers: func(*gorm.DB, configsvc.Service) error {
			calls = append(calls, "seedUsers")
			return nil
		},
		seedHomepage: func(*gorm.DB, string) error {
			calls = append(calls, "seedHomepage")
			return nil
		},
	}

	err := runStartupMutations(nil, nil, appruntime.StartupMutationMode{Writer: true}, hooks)
	if !errors.Is(err, migrationErr) {
		t.Fatalf("expected migration error, got %v", err)
	}

	want := []string{"migrate"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestRunStartupMutationsNonWriterReturnsVerificationFailure(t *testing.T) {
	verifyErr := errors.New("schema not ready")
	hooks := startupMutationHooks{
		verify: func(*gorm.DB) error {
			return verifyErr
		},
	}

	err := runStartupMutations(nil, nil, appruntime.StartupMutationMode{}, hooks)
	if !errors.Is(err, verifyErr) {
		t.Fatalf("expected verification error, got %v", err)
	}
}

func TestStartupIncompatibilityFieldsUseLoggerRuntimeVocabulary(t *testing.T) {
	fields := startupIncompatibilityFields("LoadDBConfigFromEnv")
	want := []logger.Field{
		logger.Event(logger.EventStartupIncompatibility),
		logger.Operation("LoadDBConfigFromEnv"),
		logger.ErrorType(logger.EventStartupIncompatibility),
	}

	if !reflect.DeepEqual(fields, want) {
		t.Fatalf("fields = %+v, want %+v", fields, want)
	}
}

func TestStartupMigrationFailureFieldsUseLoggerRuntimeVocabulary(t *testing.T) {
	fields := startupMigrationFailureFields("RunStartupMutations")
	want := []logger.Field{
		logger.Event(logger.EventStartupMigrationFailed),
		logger.Operation("RunStartupMutations"),
		logger.ErrorType(logger.EventStartupMigrationFailed),
	}

	if !reflect.DeepEqual(fields, want) {
		t.Fatalf("fields = %+v, want %+v", fields, want)
	}
}
