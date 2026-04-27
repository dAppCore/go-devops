package setup

import (
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

	mustLen(t, filtered, 2)
	mustEqual(t, "module-a", filtered[0].Name)
	mustEqual(t, "product-a", filtered[1].Name)
}

func TestFilterReposByTypes_EmptyFilter_Good(t *testing.T) {
	reposList := []*repos.Repo{
		{Name: "foundation-a", Type: "foundation"},
		{Name: "module-a", Type: "module"},
	}

	filtered := filterReposByTypes(reposList, nil)

	mustLen(t, filtered, 2)
	mustDeepEqual(t, reposList, filtered)
}
