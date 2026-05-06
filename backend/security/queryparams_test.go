package security

import "testing"

func TestParseBoundedIntParam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		raw          string
		defaultValue int
		minValue     int
		maxValue     int
		want         int
	}{
		{
			name:         "empty uses default",
			raw:          "",
			defaultValue: 20,
			minValue:     1,
			maxValue:     50,
			want:         20,
		},
		{
			name:         "valid value is preserved",
			raw:          "5",
			defaultValue: 20,
			minValue:     1,
			maxValue:     50,
			want:         5,
		},
		{
			name:         "non integer uses default",
			raw:          "abc",
			defaultValue: 20,
			minValue:     1,
			maxValue:     50,
			want:         20,
		},
		{
			name:         "below minimum uses default",
			raw:          "0",
			defaultValue: 20,
			minValue:     1,
			maxValue:     50,
			want:         20,
		},
		{
			name:         "above maximum uses default",
			raw:          "51",
			defaultValue: 20,
			minValue:     1,
			maxValue:     50,
			want:         20,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ParseBoundedIntParam(tt.raw, tt.defaultValue, tt.minValue, tt.maxValue)
			if got != tt.want {
				t.Fatalf("ParseBoundedIntParam(%q, %d, %d, %d) = %d, want %d", tt.raw, tt.defaultValue, tt.minValue, tt.maxValue, got, tt.want)
			}
		})
	}
}
