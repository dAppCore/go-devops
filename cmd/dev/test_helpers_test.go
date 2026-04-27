package dev

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

func mustContains(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("expected %q to contain %q", s, sub)
	}
}

func mustContainsString(t *testing.T, haystack []string, needle string) {
	t.Helper()
	for _, s := range haystack {
		if s == needle {
			return
		}
	}
	t.Fatalf("expected %v to contain %q", haystack, needle)
}

func mustNotContains(t *testing.T, s, sub string) {
	t.Helper()
	if strings.Contains(s, sub) {
		t.Fatalf("expected %q to not contain %q", s, sub)
	}
}

func mustTrue(t *testing.T, cond bool) {
	t.Helper()
	if !cond {
		t.Fatal("expected true")
	}
}

func mustFalse(t *testing.T, cond bool) {
	t.Helper()
	if cond {
		t.Fatal("expected false")
	}
}

func mustLen[T any](t *testing.T, got []T, want int) {
	t.Helper()
	if len(got) != want {
		t.Fatalf("want length %d, got %d", want, len(got))
	}
}

func mustNotNil(t *testing.T, v any) {
	t.Helper()
	if v == nil {
		t.Fatal("expected non-nil")
	}
}
