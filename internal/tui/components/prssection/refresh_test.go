package prssection

import (
	"testing"
	"time"

	graphql "github.com/cli/shurcooL-graphql"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
)

func newRefreshTestModel(t *testing.T) Model {
	t.Helper()
	primary := &data.PullRequestData{
		Number:         42,
		Url:            "https://github.com/dlvhdr/gh-dash/pull/42",
		Title:          "old title",
		State:          "OPEN",
		ReviewDecision: "REVIEW_REQUIRED",
		Comments:       data.Comments{TotalCount: 1},
	}
	// A field only the search query provides must survive a refresh.
	primary.Commits.Nodes = append(primary.Commits.Nodes, struct {
		Commit struct {
			StatusCheckRollup struct{ State graphql.String }
		}
	}{})
	primary.Commits.Nodes[0].Commit.StatusCheckRollup.State = "SUCCESS"

	// newPinnedTestModel builds a complete section (table, columns, styles).
	m := newPinnedTestModel(t)
	m.Prs = []prrow.Data{
		{
			Primary: &data.PullRequestData{
				Number: 7,
				Url:    "https://github.com/dlvhdr/gh-dash/pull/7",
			},
		},
		{Primary: primary},
	}
	return m
}

func TestRefreshPR_UpdatesMatchingRowByUrl(t *testing.T) {
	m := newRefreshTestModel(t)
	now := time.Now()
	enriched := data.EnrichedPullRequestData{
		Url:            "https://github.com/dlvhdr/gh-dash/pull/42",
		Number:         42,
		Title:          "new title",
		State:          "OPEN",
		ReviewDecision: "APPROVED",
		UpdatedAt:      now,
		Comments:       data.CommentsWithBody{TotalCount: 3},
		Reviews:        data.Reviews{TotalCount: 2},
	}

	require.True(t, m.RefreshPR(enriched))

	row := m.Prs[1]
	require.Equal(t, "new title", row.Primary.Title)
	require.Equal(t, "APPROVED", row.Primary.ReviewDecision)
	require.Equal(t, now, row.Primary.UpdatedAt)
	require.Equal(t, 3, row.Primary.Comments.TotalCount)
	require.Equal(t, 2, row.Primary.Reviews.TotalCount)
	require.True(t, row.IsEnriched)
	require.Equal(t, enriched, row.Enriched)
	// CI rollup is not part of the single-PR payload and must be kept.
	require.Equal(t, graphql.String("SUCCESS"),
		row.Primary.Commits.Nodes[0].Commit.StatusCheckRollup.State)

	// Other rows are untouched.
	require.Equal(t, 7, m.Prs[0].Primary.Number)
	require.False(t, m.Prs[0].IsEnriched)
}

func TestRefreshPR_NoMatchingRow(t *testing.T) {
	m := newRefreshTestModel(t)
	require.False(t, m.RefreshPR(data.EnrichedPullRequestData{
		Url: "https://github.com/other/repo/pull/42", Number: 42,
	}))
	require.Equal(t, "old title", m.Prs[1].Primary.Title)
}

func TestUpdatePRDoesNotChangeSameNumberInAnotherRepository(t *testing.T) {
	m := newRefreshTestModel(t)
	m.Prs[0].Primary.Number = 42
	m.Prs[0].Primary.State = "OPEN"
	closed := true
	_, _ = m.Update(tasks.UpdatePRMsg{PrNumber: 42, PrURL: m.Prs[1].Primary.Url, IsClosed: &closed})
	require.Equal(t, "OPEN", m.Prs[0].Primary.State)
	require.Equal(t, "CLOSED", m.Prs[1].Primary.State)
}
