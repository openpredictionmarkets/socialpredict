package marketshandlers

import (
	"context"
	"errors"
	"testing"
)

func TestIsRequestCanceled(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "context canceled", err: context.Canceled, want: true},
		{name: "deadline exceeded", err: context.DeadlineExceeded, want: true},
		{name: "wrapped canceled", err: errors.Join(errors.New("wrapped"), context.Canceled), want: true},
		{name: "other error", err: errors.New("boom"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRequestCanceled(tt.err); got != tt.want {
				t.Fatalf("isRequestCanceled(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
