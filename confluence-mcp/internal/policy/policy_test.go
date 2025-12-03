package policy

import "testing"

func TestAllowedSpacesMergesAndDedupes(t *testing.T) {
	store := NewStore([]string{"A", "B", "A"}, map[string][]string{
		"user": []string{"B", "C"},
	})

	got := store.AllowedSpaces("user")
	expected := []string{"A", "B", "C"}
	if len(got) != len(expected) {
		t.Fatalf("unexpected length: %v", got)
	}
	for i, v := range expected {
		if got[i] != v {
			t.Fatalf("expected %s at %d, got %s", v, i, got[i])
		}
	}
}

func TestAllowedSpacesFallsBackToDefault(t *testing.T) {
	store := NewStore([]string{"BASE"}, map[string][]string{})
	got := store.AllowedSpaces("unknown")
	if len(got) != 1 || got[0] != "BASE" {
		t.Fatalf("unexpected spaces: %v", got)
	}
}
