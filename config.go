package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kotaoue/pr-checklist-collector/formatter"
	"github.com/kotaoue/pr-checklist-collector/parser"
)

// errSkip is returned when the action should exit successfully without doing any work
// (e.g. the PR title does not match the configured pattern).
var errSkip = errors.New("no matching PR title; skipping")

// dateTokenRe matches user-friendly date tokens (longest alternative first so
// that yyyy is matched before yy at each position).
var dateTokenRe = regexp.MustCompile(`yyyy|yy|mm|dd`)

// dateMarkerRe matches {…} placeholders in an output_file path.
var dateMarkerRe = regexp.MustCompile(`\{([^}]*)\}`)

// config holds all runtime parameters derived from environment variables.
type config struct {
	owner           string
	repo            string
	token           string
	date            string
	checksKey       string
	checks          []formatter.Check
	outputFile      string
	baseBranch      string
	commitUserName  string
	commitUserEmail string
	prTitlePattern  *regexp.Regexp // nil means accept all PR titles
}

// configFromEnv builds a config from environment variables set by the GitHub Actions runner.
func configFromEnv() (*config, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is required")
	}

	rawRepo := os.Getenv("GITHUB_REPOSITORY")
	parts := strings.SplitN(rawRepo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("GITHUB_REPOSITORY must be in owner/repo format, got %q", rawRepo)
	}

	outputFile := expandDateMarkers(os.Getenv("INPUT_OUTPUT_FILE"))
	if outputFile == "" {
		return nil, fmt.Errorf("INPUT_OUTPUT_FILE is required")
	}

	// Compile the optional PR title pattern before reading the event payload so
	// that a bad pattern is caught early and clearly.
	var prTitlePattern *regexp.Regexp
	if raw := os.Getenv("INPUT_PR_TITLE_PATTERN"); raw != "" {
		var err error
		prTitlePattern, err = regexp.Compile(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid INPUT_PR_TITLE_PATTERN %q: %w", raw, err)
		}
	}

	prBody, baseBranch, prTitle, err := readPREvent()
	if err != nil {
		return nil, err
	}

	// If a title pattern is configured, skip PRs that don't match.
	if prTitlePattern != nil && !prTitlePattern.MatchString(prTitle) {
		return nil, fmt.Errorf("%w: title %q does not match %q", errSkip, prTitle, prTitlePattern.String())
	}

	checks := parser.ParseBody(prBody)
	if len(checks) == 0 {
		return nil, fmt.Errorf("no checklist items found in pull request body")
	}

	checksKey := os.Getenv("INPUT_CHECKS_KEY")
	if checksKey == "" {
		checksKey = "checks"
	}

	commitUserName := os.Getenv("INPUT_COMMIT_USER_NAME")
	if commitUserName == "" {
		commitUserName = "github-actions[bot]"
	}
	commitUserEmail := os.Getenv("INPUT_COMMIT_USER_EMAIL")
	if commitUserEmail == "" {
		commitUserEmail = "github-actions[bot]@users.noreply.github.com"
	}

	return &config{
		owner:           parts[0],
		repo:            parts[1],
		token:           token,
		date:            time.Now().Format("2006-01-02"),
		checksKey:       checksKey,
		checks:          checks,
		outputFile:      outputFile,
		baseBranch:      baseBranch,
		commitUserName:  commitUserName,
		commitUserEmail: commitUserEmail,
		prTitlePattern:  prTitlePattern,
	}, nil
}

// prEventPayload holds the fields we need from the pull_request event JSON.
type prEventPayload struct {
	PullRequest struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		Base  struct {
			Ref string `json:"ref"`
		} `json:"base"`
	} `json:"pull_request"`
}

// readPREvent reads the GitHub event payload from GITHUB_EVENT_PATH and returns
// the pull request body, the base branch ref, and the PR title.
func readPREvent() (body string, baseBranch string, title string, err error) {
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return "", "", "", fmt.Errorf("GITHUB_EVENT_PATH is not set")
	}

	data, err := os.ReadFile(eventPath)
	if err != nil {
		return "", "", "", fmt.Errorf("read event file: %w", err)
	}

	var payload prEventPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", "", "", fmt.Errorf("parse event payload: %w", err)
	}

	if payload.PullRequest.Base.Ref == "" {
		return "", "", "", fmt.Errorf("pull_request.base.ref is empty in event payload")
	}

	return payload.PullRequest.Body, payload.PullRequest.Base.Ref, payload.PullRequest.Title, nil
}

// toGoTimeLayout converts a user-friendly date pattern to Go's reference-time
// layout string. Supported tokens (case-insensitive):
//
//	yyyy → 2006  (4-digit year)
//	yy   → 06    (2-digit year)
//	mm   → 01    (2-digit month)
//	dd   → 02    (2-digit day)
//
// Go reference-time tokens (2006, 01, 02) are also accepted as-is.
func toGoTimeLayout(pattern string) string {
	return dateTokenRe.ReplaceAllStringFunc(strings.ToLower(pattern), func(token string) string {
		switch token {
		case "yyyy":
			return "2006"
		case "yy":
			return "06"
		case "mm":
			return "01"
		case "dd":
			return "02"
		}
		return token
	})
}

// expandDateMarkers replaces {pattern} placeholders in path with the current
// date formatted according to pattern. Pattern may use yyyy/yy/mm/dd tokens or
// Go reference-time tokens (2006, 01, 02). Parts of path outside {} are left
// unchanged, so a literal "2006-01-02" in the path stays as-is.
func expandDateMarkers(path string) string {
	now := time.Now()
	return dateMarkerRe.ReplaceAllStringFunc(path, func(match string) string {
		inner := match[1 : len(match)-1]
		return now.Format(toGoTimeLayout(inner))
	})
}
