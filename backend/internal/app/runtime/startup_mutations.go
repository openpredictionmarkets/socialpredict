package runtime

import (
	"fmt"

	configsvc "socialpredict/internal/service/config"

	"gorm.io/gorm"
)

type StartupMutationHooks struct {
	Migrate      func(*gorm.DB) error
	Verify       func(*gorm.DB) error
	SeedUsers    func(*gorm.DB, configsvc.Service) error
	SeedHomepage func(*gorm.DB, string) error
}

func RunStartupMutations(db *gorm.DB, configService configsvc.Service, mode StartupMutationMode, hooks StartupMutationHooks) error {
	if mode.Writer {
		if hooks.Migrate == nil || hooks.SeedUsers == nil || hooks.SeedHomepage == nil {
			return fmt.Errorf("startup writer hooks unavailable")
		}
		if err := hooks.Migrate(db); err != nil {
			return fmt.Errorf("migrate database: %w", err)
		}
		if err := hooks.SeedUsers(db, configService); err != nil {
			return fmt.Errorf("seed users: %w", err)
		}
		if err := hooks.SeedHomepage(db, "."); err != nil {
			return fmt.Errorf("seed homepage: %w", err)
		}
		return nil
	}

	if hooks.Verify == nil {
		return fmt.Errorf("startup schema verification hook unavailable")
	}
	if err := hooks.Verify(db); err != nil {
		return fmt.Errorf("verify applied migrations: %w", err)
	}
	return nil
}
