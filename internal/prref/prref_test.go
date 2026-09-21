package prref

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	current := &repository.Repository{Host: "github.com", Owner: "dlvhdr", Name: "gh-dash"}

	tests := []struct {
		name    string
		input   string
		current *repository.Repository
		want    Ref
	}{
		{
			name:  "full url",
			input: "https://github.com/dlvhdr/gh-dash/pull/767",
			want:  Ref{Host: "github.com", Owner: "dlvhdr", Name: "gh-dash", Number: 767},
		},
		{
			name:  "url with files subpath and fragment",
			input: "https://github.com/dlvhdr/gh-dash/pull/767/files#diff-abc",
			want:  Ref{Host: "github.com", Owner: "dlvhdr", Name: "gh-dash", Number: 767},
		},
		{
			name:  "url with query",
			input: "https://github.com/dlvhdr/gh-dash/pull/767?notification_referrer_id=x",
			want:  Ref{Host: "github.com", Owner: "dlvhdr", Name: "gh-dash", Number: 767},
		},
		{
			name:  "url without scheme",
			input: "github.com/dlvhdr/gh-dash/pull/767",
			want:  Ref{Host: "github.com", Owner: "dlvhdr", Name: "gh-dash", Number: 767},
		},
		{
			name:  "enterprise host",
			input: "https://ghe.example.com/org/repo/pull/12",
			want:  Ref{Host: "ghe.example.com", Owner: "org", Name: "repo", Number: 12},
		},
		{
			name:  "shorthand",
			input: "dlvhdr/gh-dash#767",
			want:  Ref{Host: "github.com", Owner: "dlvhdr", Name: "gh-dash", Number: 767},
		},
		{
			name:  "shorthand with host",
			input: "ghe.example.com/org/repo#5",
			want:  Ref{Host: "ghe.example.com", Owner: "org", Name: "repo", Number: 5},
		},
		{
			name:    "bare number with current repo",
			input:   "767",
			current: current,
			want:    Ref{Host: "github.com", Owner: "dlvhdr", Name: "gh-dash", Number: 767},
		},
		{
			name:    "hash number with current repo",
			input:   "#767",
			current: current,
			want:    Ref{Host: "github.com", Owner: "dlvhdr", Name: "gh-dash", Number: 767},
		},
		{
			name:    "surrounding whitespace",
			input:   "  https://github.com/dlvhdr/gh-dash/pull/767  ",
			current: current,
			want:    Ref{Host: "github.com", Owner: "dlvhdr", Name: "gh-dash", Number: 767},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input, tt.current)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParse_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		current *repository.Repository
	}{
		{name: "empty", input: ""},
		{name: "bare number without current repo", input: "767"},
		{
			name:    "bare number with zero current repo",
			input:   "767",
			current: &repository.Repository{},
		},
		{name: "issue url", input: "https://github.com/dlvhdr/gh-dash/issues/767"},
		{name: "repo url", input: "https://github.com/dlvhdr/gh-dash"},
		{name: "non numeric", input: "https://github.com/dlvhdr/gh-dash/pull/abc"},
		{name: "zero number", input: "dlvhdr/gh-dash#0"},
		{name: "shorthand missing number", input: "dlvhdr/gh-dash#"},
		{name: "garbage", input: "not a pr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input, tt.current)
			require.Error(t, err)
		})
	}
}

func TestRef_URL(t *testing.T) {
	r := Ref{Host: "github.com", Owner: "dlvhdr", Name: "gh-dash", Number: 767}
	require.Equal(t, "https://github.com/dlvhdr/gh-dash/pull/767", r.URL())
	require.Equal(t, "dlvhdr/gh-dash#767", r.String())
}
