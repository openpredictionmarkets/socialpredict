package test

import (
	"reflect"
	"socialpredict/handlers/math/outcomes/dbpm"
	test "socialpredict/tests"
	"testing"
)

func TestNetAggregateMarketPositions(t *testing.T) {
	for _, tc := range test.TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := dbpm.NetAggregateMarketPositions(tc.AggregatedPositions)
			expected := tc.NetPositions

			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("Failed %s: expected %+v, got %+v", tc.Name, expected, actual)
			}
		})
	}
}
