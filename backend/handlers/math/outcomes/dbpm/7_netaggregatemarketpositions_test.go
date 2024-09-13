package dbpm

import (
	"reflect"
	"testing"
)

func TestNetAggregateMarketPositions(t *testing.T) {
	for _, tc := range TestCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := NetAggregateMarketPositions(tc.AggregatedPositions)
			expected := tc.NetPositions

			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("Failed %s: expected %+v, got %+v", tc.Name, expected, actual)
			}
		})
	}
}
