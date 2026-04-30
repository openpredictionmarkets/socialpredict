package runtime

import (
	"context"
	"testing"

	"socialpredict/models/modelstesting"
)

func TestServingProbeLivenessIsIndependentOfReadiness(t *testing.T) {
	probe := NewServingProbe(nil, nil)

	if !probe.Live() {
		t.Fatalf("expected serving probe to report live process")
	}
}

func TestServingProbeReadinessRequiresStartupGateAndDatabase(t *testing.T) {
	ctx := context.Background()
	db := modelstesting.NewFakeDB(t)
	readiness := NewReadiness()
	probe := NewServingProbe(db, readiness)

	if err := probe.Ready(ctx); err == nil {
		t.Fatalf("expected readiness to fail while startup gate is closed")
	}

	readiness.MarkReady()
	if err := probe.Ready(ctx); err != nil {
		t.Fatalf("expected readiness to pass with open startup gate and reachable database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close sql db: %v", err)
	}
	if err := probe.Ready(ctx); err == nil {
		t.Fatalf("expected readiness to fail when database is unavailable")
	}
}
