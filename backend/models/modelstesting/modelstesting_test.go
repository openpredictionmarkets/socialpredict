package modelstesting

import "testing"

func TestNewFakeDB(t *testing.T) {
	db := NewFakeDB(t)
	if db == nil {
		t.Error("Failed to create fake db")
	}
}
