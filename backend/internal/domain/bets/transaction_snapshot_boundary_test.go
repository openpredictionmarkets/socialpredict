package bets_test

import (
	"reflect"
	"strings"
	"testing"

	bets "socialpredict/internal/domain/bets"
)

func TestBetTransactionInterfacesDoNotExposeSnapshotMethods(t *testing.T) {
	for _, tc := range []struct {
		name         string
		interfaceRef any
	}{
		{name: "Repository", interfaceRef: (*bets.Repository)(nil)},
		{name: "BetWriter", interfaceRef: (*bets.BetWriter)(nil)},
		{name: "BetHistoryReader", interfaceRef: (*bets.BetHistoryReader)(nil)},
		{name: "MarketReader", interfaceRef: (*bets.MarketReader)(nil)},
		{name: "PositionReader", interfaceRef: (*bets.PositionReader)(nil)},
		{name: "MarketService", interfaceRef: (*bets.MarketService)(nil)},
		{name: "PlaceUnitOfWork", interfaceRef: (*bets.PlaceUnitOfWork)(nil)},
		{name: "SellUnitOfWork", interfaceRef: (*bets.SellUnitOfWork)(nil)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertNoSnapshotMethods(t, reflect.TypeOf(tc.interfaceRef).Elem())
		})
	}
}

func assertNoSnapshotMethods(t *testing.T, iface reflect.Type) {
	t.Helper()
	for i := 0; i < iface.NumMethod(); i++ {
		method := iface.Method(i)
		if strings.Contains(method.Name, "Snapshot") || strings.Contains(method.Name, "Accounting") {
			t.Fatalf("%s unexpectedly exposes read-model/cache method %s", iface.Name(), method.Name)
		}
	}
}
