package devkit

import (
	"math"
	"testing"
)

func mustNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func mustEqual[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want != got {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func mustLen[T any](t *testing.T, got []T, want int) {
	t.Helper()
	if len(got) != want {
		t.Fatalf("want length %d, got %d", want, len(got))
	}
}

func mustEmpty[T any](t *testing.T, got []T) {
	t.Helper()
	if len(got) != 0 {
		t.Fatalf("expected empty, got %d entries", len(got))
	}
}

func mustInDelta(t *testing.T, want, got, delta float64) {
	t.Helper()
	if math.Abs(want-got) > delta {
		t.Fatalf("want %v±%v, got %v", want, delta, got)
	}
}
