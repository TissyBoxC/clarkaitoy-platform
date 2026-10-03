package repository

import (
	"reflect"
	"testing"
)

func TestNonNilStrings(t *testing.T) {
	t.Run("preserves nil as empty array", func(t *testing.T) {
		got := nonNilStrings(nil)
		if got == nil {
			t.Fatal("expected non-nil empty slice")
		}
		if len(got) != 0 {
			t.Fatalf("expected empty slice, got %v", got)
		}
	})

	t.Run("preserves populated values", func(t *testing.T) {
		want := []string{"model-a", "model-b"}
		got := nonNilStrings(want)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("expected %v, got %v", want, got)
		}
	})
}
