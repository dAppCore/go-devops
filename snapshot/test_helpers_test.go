package snapshot

import (
	"reflect"
	"strings"
	"testing"
)

func mustNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustEqual[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want != got {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func mustDeepEqual(t *testing.T, want, got any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func mustLenMap[K comparable, V any](t *testing.T, m map[K]V, want int) {
	t.Helper()
	if len(m) != want {
		t.Fatalf("want length %d, got %d", want, len(m))
	}
}

func mustErrorContains(t *testing.T, err error, sub string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), sub) {
		t.Fatalf("expected error %q to contain %q", err.Error(), sub)
	}
}
