package analytics

import (
	"reflect"
	"strings"
	"testing"
)

func TestAnalyticsTransactionAdjacentInterfacesDoNotExposeUserFinancialSnapshots(t *testing.T) {
	for _, tc := range []struct {
		name         string
		interfaceRef any
	}{
		{name: "Repository", interfaceRef: (*Repository)(nil)},
		{name: "FinancialsRepository", interfaceRef: (*FinancialsRepository)(nil)},
		{name: "DebtRepository", interfaceRef: (*DebtRepository)(nil)},
		{name: "LeaderboardRepository", interfaceRef: (*LeaderboardRepository)(nil)},
		{name: "StatsRepository", interfaceRef: (*StatsRepository)(nil)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertNoUserFinancialSnapshotMethods(t, reflect.TypeOf(tc.interfaceRef).Elem())
		})
	}
}

func TestUserFinancialMetricSnapshotRepositoryIsSeparateFromAnalyticsRepository(t *testing.T) {
	analyticsRepository := reflect.TypeOf((*Repository)(nil)).Elem()
	snapshotRepository := reflect.TypeOf((*UserFinancialMetricSnapshotRepository)(nil)).Elem()

	for i := 0; i < snapshotRepository.NumMethod(); i++ {
		method := snapshotRepository.Method(i)
		if _, ok := analyticsRepository.MethodByName(method.Name); ok {
			t.Fatalf("analytics Repository unexpectedly exposes user financial snapshot method %s", method.Name)
		}
	}
}

func assertNoUserFinancialSnapshotMethods(t *testing.T, iface reflect.Type) {
	t.Helper()
	for i := 0; i < iface.NumMethod(); i++ {
		method := iface.Method(i)
		if strings.Contains(method.Name, "UserFinancialMetricSnapshot") {
			t.Fatalf("%s unexpectedly exposes %s", iface.Name(), method.Name)
		}
	}
}
