package setup

import (
	"slices"
	"testing"

	"dappco.re/go/scm/repos"
)

func TestFilterReposByTypes_Good(t *testing.T) {
	reposList := []*repos.Repo{
		{Name: "foundation-a", Type: "foundation"},
		{Name: "module-a", Type: "module"},
		{Name: "product-a", Type: "product"},
	}

	filtered := filterReposByTypes(reposList, []string{"module", "product"})

	if len(filtered) != 2 {
		t.Fatalf("filtered length = %d, want 2", len(filtered))
	}
	if filtered[0].Name != "module-a" {
		t.Fatalf("filtered[0].Name = %q, want %q", filtered[0].Name, "module-a")
	}
	if filtered[1].Name != "product-a" {
		t.Fatalf("filtered[1].Name = %q, want %q", filtered[1].Name, "product-a")
	}
}

func TestFilterReposByTypes_EmptyFilter_Good(t *testing.T) {
	reposList := []*repos.Repo{
		{Name: "foundation-a", Type: "foundation"},
		{Name: "module-a", Type: "module"},
	}

	filtered := filterReposByTypes(reposList, nil)

	if len(filtered) != 2 {
		t.Fatalf("filtered length = %d, want 2", len(filtered))
	}
	if !slices.Equal(filtered, reposList) {
		t.Fatalf("filtered = %v, want %v", filtered, reposList)
	}
}
