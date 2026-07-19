package mcpserver

import "testing"

func TestNewPageOutputWithoutTotal(t *testing.T) {
	got := NewPageOutput(20, 40, 20, nil)
	if !got.HasMore {
		t.Fatalf("HasMore = false, want true")
	}
	if got.NextOffset == nil || *got.NextOffset != 60 {
		t.Fatalf("NextOffset = %#v, want 60", got.NextOffset)
	}
	if got.Total != nil {
		t.Fatalf("Total = %#v, want nil", got.Total)
	}

	got = NewPageOutput(20, 40, 7, nil)
	if got.HasMore || got.NextOffset != nil {
		t.Fatalf("short page should not have next page: %#v", got)
	}
}

func TestNewPageOutputWithTotal(t *testing.T) {
	total := 45
	got := NewPageOutput(20, 20, 20, &total)
	if !got.HasMore || got.NextOffset == nil || *got.NextOffset != 40 {
		t.Fatalf("middle page = %#v", got)
	}

	got = NewPageOutput(20, 40, 5, &total)
	if got.HasMore || got.NextOffset != nil {
		t.Fatalf("last page = %#v", got)
	}
	if got.Total == nil || *got.Total != 45 {
		t.Fatalf("Total = %#v, want 45", got.Total)
	}
}
