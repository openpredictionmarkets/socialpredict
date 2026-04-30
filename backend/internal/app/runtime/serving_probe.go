package runtime

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// ServingProbe owns process liveness and request-serving readiness checks.
type ServingProbe struct {
	db        *gorm.DB
	readiness *Readiness
}

func NewServingProbe(db *gorm.DB, readiness *Readiness) ServingProbe {
	return ServingProbe{db: db, readiness: readiness}
}

func (p ServingProbe) Live() bool {
	return true
}

func (p ServingProbe) Ready(ctx context.Context) error {
	if p.readiness == nil || !p.readiness.Ready() {
		return fmt.Errorf("startup readiness gate closed")
	}
	if err := CheckDBReadiness(ctx, p.db); err != nil {
		return err
	}
	return nil
}
