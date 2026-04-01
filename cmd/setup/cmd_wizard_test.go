package setup

import (
	"testing"

	"dappco.re/go/core/scm/repos"
	"github.com/stretchr/testify/require"
)

func TestFilterReposByTypes_Good(t *testing.T) {
	reposList := []*repos.Repo{
		{Name: "foundation-a", Type: "foundation"},
		{Name: "module-a", Type: "module"},
		{Name: "product-a", Type: "product"},
	}

	filtered := filterReposByTypes(reposList, []string{"module", "product"})

	require.Len(t, filtered, 2)
	require.Equal(t, "module-a", filtered[0].Name)
	require.Equal(t, "product-a", filtered[1].Name)
}

func TestFilterReposByTypes_EmptyFilter_Good(t *testing.T) {
	reposList := []*repos.Repo{
		{Name: "foundation-a", Type: "foundation"},
		{Name: "module-a", Type: "module"},
	}

	filtered := filterReposByTypes(reposList, nil)

	require.Len(t, filtered, 2)
	require.Equal(t, reposList, filtered)
}
