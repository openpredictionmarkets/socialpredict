package runtime

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type dbPoolWaitMeasurement struct {
	WaitCount    int64
	WaitDuration time.Duration
	Elapsed      time.Duration
}

func TestMeasureDBPoolWaitDifferentiatesSaturation(t *testing.T) {
	saturated := measureDBPoolWait(t, 1, 1, 1)
	if saturated.WaitCount == 0 {
		t.Fatalf("expected saturated pool to record at least one wait, got %+v", saturated)
	}
	if saturated.WaitDuration <= 0 {
		t.Fatalf("expected saturated pool to record positive wait duration, got %+v", saturated)
	}

	headroom := measureDBPoolWait(t, 2, 1, 1)
	if headroom.WaitCount != 0 {
		t.Fatalf("expected pool with spare connection capacity not to wait, got %+v", headroom)
	}
}

func BenchmarkDBPoolWaitUnderSaturation(b *testing.B) {
	for _, tc := range []struct {
		name      string
		maxOpen   int
		heldConns int
		waiters   int
	}{
		{
			name:      "max_open_1_saturated",
			maxOpen:   1,
			heldConns: 1,
			waiters:   1,
		},
		{
			name:      "max_open_2_headroom",
			maxOpen:   2,
			heldConns: 1,
			waiters:   1,
		},
	} {
		b.Run(tc.name, func(b *testing.B) {
			var totalWaits int64
			var totalWaitDuration time.Duration

			for range b.N {
				measurement := measureDBPoolWait(b, tc.maxOpen, tc.heldConns, tc.waiters)
				totalWaits += measurement.WaitCount
				totalWaitDuration += measurement.WaitDuration
			}

			b.ReportMetric(float64(totalWaits)/float64(b.N), "pool_waits/op")
			b.ReportMetric(float64(totalWaitDuration.Nanoseconds())/float64(b.N), "pool_wait_ns/op")
		})
	}
}

func measureDBPoolWait(tb testing.TB, maxOpenConns, heldConns, waiters int) dbPoolWaitMeasurement {
	tb.Helper()
	if maxOpenConns < 1 {
		tb.Fatalf("maxOpenConns must be positive, got %d", maxOpenConns)
	}
	if heldConns > maxOpenConns {
		tb.Fatalf("heldConns %d exceeds maxOpenConns %d", heldConns, maxOpenConns)
	}

	db, sqlDB := newMeasurementDB(tb, maxOpenConns)
	defer func() {
		if err := CloseDB(db); err != nil {
			tb.Fatalf("close measurement db: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	held := acquireMeasurementConns(tb, ctx, sqlDB, heldConns)
	before := sqlDB.Stats()

	start := time.Now()
	var wg sync.WaitGroup
	errs := make(chan error, waiters)
	for range waiters {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := sqlDB.Conn(ctx)
			if err != nil {
				errs <- err
				return
			}
			errs <- conn.Close()
		}()
	}

	if heldConns == maxOpenConns && waiters > 0 {
		waitForPoolWait(tb, ctx, sqlDB, before.WaitCount)
	}
	for _, conn := range held {
		if err := conn.Close(); err != nil {
			tb.Fatalf("release held connection: %v", err)
		}
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			tb.Fatalf("measurement waiter: %v", err)
		}
	}

	after := sqlDB.Stats()
	return dbPoolWaitMeasurement{
		WaitCount:    after.WaitCount - before.WaitCount,
		WaitDuration: after.WaitDuration - before.WaitDuration,
		Elapsed:      time.Since(start),
	}
}

func newMeasurementDB(tb testing.TB, maxOpenConns int) (*gorm.DB, *sql.DB) {
	tb.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		tb.Fatalf("open measurement sqlite db: %v", err)
	}
	if err := ConfigureDBPool(db, DBConfig{
		MaxOpenConns: maxOpenConns,
		MaxIdleConns: maxOpenConns,
	}); err != nil {
		tb.Fatalf("configure measurement db pool: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		tb.Fatalf("measurement sql db: %v", err)
	}
	return db, sqlDB
}

func acquireMeasurementConns(tb testing.TB, ctx context.Context, db *sql.DB, count int) []*sql.Conn {
	tb.Helper()

	conns := make([]*sql.Conn, 0, count)
	for range count {
		conn, err := db.Conn(ctx)
		if err != nil {
			tb.Fatalf("hold measurement connection: %v", err)
		}
		conns = append(conns, conn)
	}
	return conns
}

func waitForPoolWait(tb testing.TB, ctx context.Context, db *sql.DB, previousWaitCount int64) {
	tb.Helper()

	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		if db.Stats().WaitCount > previousWaitCount {
			return
		}

		select {
		case <-ctx.Done():
			tb.Fatalf("timed out waiting for pool wait: %v", ctx.Err())
		case <-ticker.C:
		}
	}
}
