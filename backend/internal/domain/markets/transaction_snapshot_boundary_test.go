package markets_test

import (
	"reflect"
	"testing"

	markets "socialpredict/internal/domain/markets"
)

func TestMarketTransactionRepositoryInterfacesDoNotExposeAccountingSnapshots(t *testing.T) {
	for _, tc := range []struct {
		name         string
		interfaceRef any
	}{
		{name: "Repository", interfaceRef: (*markets.Repository)(nil)},
		{name: "MarketWriteRepository", interfaceRef: (*markets.MarketWriteRepository)(nil)},
		{name: "MarketPositionRepository", interfaceRef: (*markets.MarketPositionRepository)(nil)},
		{name: "MarketBetRepository", interfaceRef: (*markets.MarketBetRepository)(nil)},
		{name: "ResolutionRepository", interfaceRef: (*markets.ResolutionRepository)(nil)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertInterfaceDoesNotExposeSnapshotMethods(t, reflect.TypeOf(tc.interfaceRef).Elem())
		})
	}
}

func TestMarketAccountingSnapshotRepositoryIsSeparateFromTransactionRepository(t *testing.T) {
	transactionRepository := reflect.TypeOf((*markets.Repository)(nil)).Elem()
	snapshotRepository := reflect.TypeOf((*markets.MarketAccountingSnapshotRepository)(nil)).Elem()

	for i := 0; i < snapshotRepository.NumMethod(); i++ {
		method := snapshotRepository.Method(i)
		if _, ok := transactionRepository.MethodByName(method.Name); ok {
			t.Fatalf("transaction Repository unexpectedly exposes snapshot method %s", method.Name)
		}
	}
}

func assertInterfaceDoesNotExposeSnapshotMethods(t *testing.T, iface reflect.Type) {
	t.Helper()
	for _, methodName := range []string{
		"GetMarketAccountingSnapshot",
		"UpsertMarketAccountingSnapshot",
	} {
		if _, ok := iface.MethodByName(methodName); ok {
			t.Fatalf("%s unexpectedly exposes %s", iface.Name(), methodName)
		}
	}
}
