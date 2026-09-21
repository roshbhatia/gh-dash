// Package prref parses user-supplied references to a GitHub pull request.
//
// Supported forms:
//
//	https://github.com/owner/repo/pull/123            (any host, extra path/query/fragment ignored)
//	github.com/owner/repo/pull/123
//	owner/repo#123
//	#123 or 123                                        (requires a current repository)
package prref

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
)

const DefaultHost = "github.com"

// Ref identifies a single pull request.
type Ref struct {
	Host   string
	Owner  string
	Name   string
	Number int
}

// URL returns the canonical HTML URL of the pull request.
func (r Ref) URL() string {
	return fmt.Sprintf("https://%s/%s/%s/pull/%d", r.Host, r.Owner, r.Name, r.Number)
}

// NameWithOwner returns "owner/name".
func (r Ref) NameWithOwner() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

// String returns a short human readable form, e.g. "owner/name#123".
func (r Ref) String() string {
	return fmt.Sprintf("%s#%d", r.NameWithOwner(), r.Number)
}

// Parse turns a user-supplied reference into a Ref. current is used to resolve
// bare numbers and may be nil or the zero value when no repository is known.
func Parse(input string, current *repository.Repository) (Ref, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return Ref{}, errors.New("empty pull request reference")
	}

	if strings.Contains(s, "://") || strings.Contains(s, "/pull/") {
		return parseURL(s)
	}

	if nameWithOwner, num, ok := strings.Cut(s, "#"); ok && strings.Contains(nameWithOwner, "/") {
		return parseShorthand(nameWithOwner, num)
	}

	numStr := strings.TrimPrefix(s, "#")
	n, err := parseNumber(numStr)
	if err != nil {
		return Ref{}, fmt.Errorf(
			"%q is not a pull request URL, owner/repo#number, or number", input)
	}
	if current == nil || *current == (repository.Repository{}) {
		return Ref{}, fmt.Errorf(
			"cannot resolve pull request #%d: not inside a GitHub repository; pass a full URL or owner/repo#%d",
			n,
			n,
		)
	}
	host := current.Host
	if host == "" {
		host = DefaultHost
	}
	return Ref{Host: host, Owner: current.Owner, Name: current.Name, Number: n}, nil
}

// parseShorthand handles "owner/repo#123" and "host/owner/repo#123".
func parseShorthand(nameWithOwner, num string) (Ref, error) {
	parts := strings.Split(strings.Trim(nameWithOwner, "/"), "/")
	var host, owner, name string
	switch len(parts) {
	case 2:
		host, owner, name = DefaultHost, parts[0], parts[1]
	case 3:
		host, owner, name = parts[0], parts[1], parts[2]
	default:
		return Ref{}, fmt.Errorf("%q is not in the form owner/repo#number", nameWithOwner+"#"+num)
	}
	if owner == "" || name == "" {
		return Ref{}, fmt.Errorf("%q is not in the form owner/repo#number", nameWithOwner+"#"+num)
	}
	n, err := parseNumber(num)
	if err != nil {
		return Ref{}, fmt.Errorf("%q is not a valid pull request number", num)
	}
	return Ref{Host: host, Owner: owner, Name: name, Number: n}, nil
}

func parseURL(s string) (Ref, error) {
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return Ref{}, fmt.Errorf("invalid pull request URL %q: %w", s, err)
	}
	if u.Hostname() == "" {
		return Ref{}, fmt.Errorf("invalid pull request URL %q: missing host", s)
	}

	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	// owner/repo/pull/123[/...]
	if len(segments) < 4 || segments[2] != "pull" {
		if len(segments) >= 4 && segments[2] == "issues" {
			return Ref{}, fmt.Errorf("%q points to an issue, only pull requests are supported", s)
		}
		return Ref{}, fmt.Errorf(
			"%q is not a pull request URL (expected https://host/owner/repo/pull/number)", s)
	}
	n, err := parseNumber(segments[3])
	if err != nil {
		return Ref{}, fmt.Errorf("%q does not contain a valid pull request number", s)
	}
	return Ref{Host: u.Hostname(), Owner: segments[0], Name: segments[1], Number: n}, nil
}

func parseNumber(s string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid number %q", s)
	}
	return n, nil
}
