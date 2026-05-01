package main

import (
	"fmt"

	appruntime "socialpredict/internal/app/runtime"
	configsvc "socialpredict/internal/service/config"

	"gorm.io/gorm"
)

type startupMutationHooks struct {
	migrate      func(*gorm.DB) error
	verify       func(*gorm.DB) error
	seedUsers    func(*gorm.DB, configsvc.Service) error
	seedHomepage func(*gorm.DB, string) error
}

func runStartupMutations(db *gorm.DB, configService configsvc.Service, mode appruntime.StartupMutationMode, hooks startupMutationHooks) error {
	if mode.Writer {
		if hooks.migrate == nil || hooks.seedUsers == nil || hooks.seedHomepage == nil {
			return fmt.Errorf("startup writer hooks unavailable")
		}
		if err := hooks.migrate(db); err != nil {
			return fmt.Errorf("migrate database: %w", err)
		}
		if err := hooks.seedUsers(db, configService); err != nil {
			return fmt.Errorf("seed users: %w", err)
		}
		if err := hooks.seedHomepage(db, "."); err != nil {
			return fmt.Errorf("seed homepage: %w", err)
		}
		return nil
	}

	if hooks.verify == nil {
		return fmt.Errorf("startup schema verification hook unavailable")
	}
	if err := hooks.verify(db); err != nil {
		return fmt.Errorf("verify applied migrations: %w", err)
	}
	return nil
}
