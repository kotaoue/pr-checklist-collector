package commit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v69/github"
	"github.com/kotaoue/pr-checklist-collector/commit"
)

// newTestClient returns a GitHub client and a test server whose handler is set by
// the caller. The test server is closed automatically when the test ends.
func newTestClient(t *testing.T, mux *http.ServeMux) *github.Client {
	t.Helper()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client, err := github.NewClient(nil).WithAuthToken("test-token").WithEnterpriseURLs(srv.URL+"/", srv.URL+"/")
	if err != nil {
		t.Fatalf("create test client: %v", err)
	}
	return client
}

func TestFile_Create(t *testing.T) {
	mux := http.NewServeMux()

	// GetContents returns 404 → file does not exist yet.
	mux.HandleFunc("/api/v3/repos/owner/repo/contents/out.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// CreateFile (PUT) → return minimal success response.
		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusCreated)
			resp := map[string]interface{}{
				"content": map[string]interface{}{"name": "out.json"},
				"commit":  map[string]interface{}{"sha": "abc123"},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	})

	client := newTestClient(t, mux)
	err := commit.File(context.Background(), client, "owner", "repo", "out.json", "main", "Save results", commit.Options{}, []byte(`[]`))
	if err != nil {
		t.Errorf("File() unexpected error: %v", err)
	}
}

func TestFile_Update(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v3/repos/owner/repo/contents/out.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// File already exists.
			w.WriteHeader(http.StatusOK)
			resp := map[string]interface{}{
				"name": "out.json",
				"sha":  "existing-sha",
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		// UpdateFile (PUT) → return success.
		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
			resp := map[string]interface{}{
				"content": map[string]interface{}{"name": "out.json"},
				"commit":  map[string]interface{}{"sha": "def456"},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	})

	client := newTestClient(t, mux)
	err := commit.File(context.Background(), client, "owner", "repo", "out.json", "main", "Save results", commit.Options{}, []byte(`[]`))
	if err != nil {
		t.Errorf("File() unexpected error: %v", err)
	}
}

func TestFile_WithCommitter(t *testing.T) {
	mux := http.NewServeMux()

	var gotName, gotEmail string
	mux.HandleFunc("/api/v3/repos/owner/repo/contents/out.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == http.MethodPut {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if c, ok := body["committer"].(map[string]interface{}); ok {
				gotName, _ = c["name"].(string)
				gotEmail, _ = c["email"].(string)
			}
			w.WriteHeader(http.StatusCreated)
			resp := map[string]interface{}{
				"content": map[string]interface{}{"name": "out.json"},
				"commit":  map[string]interface{}{"sha": "abc123"},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	})

	client := newTestClient(t, mux)
	opts := commit.Options{CommitterName: "bot", CommitterEmail: "bot@example.com"}
	err := commit.File(context.Background(), client, "owner", "repo", "out.json", "main", "Save results", opts, []byte(`[]`))
	if err != nil {
		t.Errorf("File() unexpected error: %v", err)
	}
	if gotName != "bot" {
		t.Errorf("committer name = %q, want %q", gotName, "bot")
	}
	if gotEmail != "bot@example.com" {
		t.Errorf("committer email = %q, want %q", gotEmail, "bot@example.com")
	}
}
