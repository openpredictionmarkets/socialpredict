package users_test

import (
	"reflect"
	"strings"
	"testing"

	users "socialpredict/internal/domain/users"
)

func TestUserTransactionInterfacesDoNotExposeFinancialSnapshotMethods(t *testing.T) {
	for _, tc := range []struct {
		name         string
		interfaceRef any
	}{
		{name: "Repository", interfaceRef: (*users.Repository)(nil)},
		{name: "UserBalanceRepository", interfaceRef: (*users.UserBalanceRepository)(nil)},
		{name: "UserReader", interfaceRef: (*users.UserReader)(nil)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertNoFinancialSnapshotMethods(t, reflect.TypeOf(tc.interfaceRef).Elem())
		})
	}
}

func assertNoFinancialSnapshotMethods(t *testing.T, iface reflect.Type) {
	t.Helper()
	for i := 0; i < iface.NumMethod(); i++ {
		method := iface.Method(i)
		if strings.Contains(method.Name, "FinancialMetricSnapshot") || strings.Contains(method.Name, "FinancialSnapshot") {
			t.Fatalf("%s unexpectedly exposes read-model/cache method %s", iface.Name(), method.Name)
		}
	}
}
