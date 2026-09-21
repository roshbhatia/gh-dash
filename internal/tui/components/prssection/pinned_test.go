package prssection

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

const testPrUrl = "https://github.com/dlvhdr/gh-dash/pull/767"

func newPinnedTestModel(t *testing.T) Model {
	t.Helper()
	ctx := &context.ProgramContext{
		Config: &config.Config{
			SmartFilteringAtLaunch: true,
			Theme:                  &config.ThemeConfig{},
		},
		GHRepo: &repository.Repository{Host: "github.com", Owner: "dlvhdr", Name: "gh-dash"},
		Theme:  *theme.DefaultTheme,
		Styles: context.DefaultStyles,
		StartTask: func(task context.Task) tea.Cmd {
			return func() tea.Msg { return nil }
		},
	}
	return NewPinnedPrModel(0, ctx, "dlvhdr/gh-dash#767", testPrUrl)
}

func TestNewPinnedPrModel_SearchValueIsUrlOnly(t *testing.T) {
	m := newPinnedTestModel(t)

	require.True(t, m.IsPinned())
	// Smart filtering at launch must not prepend "repo:owner/name" to the URL.
	require.Equal(t, testPrUrl, m.SearchValue)
	require.Equal(t, testPrUrl, m.SearchBar.Value())
	require.False(t, m.IsFilteredByCurrentRemote)
	require.Equal(t, "dlvhdr/gh-dash#767", m.GetConfig().Title)
}

func TestPinnedPrModel_FetchUsesSinglePrPath(t *testing.T) {
	m := newPinnedTestModel(t)

	cmds := m.FetchNextPageSectionRows()
	// start task + fetch + loading spinner
	require.Len(t, cmds, 3)
	require.NotEmpty(t, m.LastFetch.TaskId)
}

func TestPinnedPrModel_EditingSearchUnpins(t *testing.T) {
	m := newPinnedTestModel(t)
	m.SetIsSearching(true)
	m.SearchBar.SetValue("is:open author:@me")

	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	require.False(t, m.IsPinned())
	require.Equal(t, "is:open author:@me", m.SearchValue)
}

func TestPinnedPrModel_ResubmittingUrlKeepsPin(t *testing.T) {
	m := newPinnedTestModel(t)
	m.SetIsSearching(true)

	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	require.True(t, m.IsPinned())
	require.Equal(t, testPrUrl, m.SearchValue)
}
