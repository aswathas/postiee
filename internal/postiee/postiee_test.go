package postiee

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The two traps that silently lose data: the post id lives only in a response
// header, and the author urn is app-scoped so it must be resolved at runtime.

func TestCreatePostReadsIDFromHeaderNotBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(PostIDHeader, "urn:li:ugcPost:1234567890123456789")
		w.WriteHeader(http.StatusCreated)
		// Body deliberately empty, exactly as LinkedIn returns it.
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "t"}
	got, err := c.CreatePost(context.Background(), CreatePostRequest{
		Author: "urn:li:person:abc", Commentary: "hello",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if got.URN != "urn:li:ugcPost:1234567890123456789" {
		t.Fatalf("post id not read from header: got %q", got.URN)
	}
	if !strings.Contains(got.Permalink, "urn:li:ugcPost:1234567890123456789") {
		t.Fatalf("permalink missing urn: %q", got.Permalink)
	}
}

// Losing the id means the post can never be deleted or recorded, so a missing
// header must be a loud error, not an empty success.
func TestCreatePostErrorsWhenHeaderMissing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "t"}
	_, err := c.CreatePost(context.Background(), CreatePostRequest{
		Author: "urn:li:person:abc", Commentary: "hello",
	})
	if err == nil {
		t.Fatal("expected an error when the id header is absent")
	}
	if !strings.Contains(err.Error(), PostIDHeader) {
		t.Fatalf("error should name the missing header, got: %v", err)
	}
}

func TestCreatePostSendsDefaultsAndVersionHeader(t *testing.T) {
	var got map[string]any
	var gotVersion, gotProto string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotVersion = r.Header.Get("Linkedin-Version")
		gotProto = r.Header.Get("X-Restli-Protocol-Version")
		json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set(PostIDHeader, "urn:li:ugcPost:1")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "t"}
	if _, err := c.CreatePost(context.Background(), CreatePostRequest{
		Author: "urn:li:person:abc", Commentary: "hi",
	}); err != nil {
		t.Fatalf("create: %v", err)
	}

	if gotVersion != Version {
		t.Errorf("Linkedin-Version = %q, want %q", gotVersion, Version)
	}
	if gotProto != "2.0.0" {
		t.Errorf("X-Restli-Protocol-Version = %q, want 2.0.0", gotProto)
	}
	if got["lifecycleState"] != "PUBLISHED" {
		t.Errorf("lifecycleState = %v, want PUBLISHED (only value accepted at create)", got["lifecycleState"])
	}
	if got["visibility"] != "PUBLIC" {
		t.Errorf("visibility = %v, want PUBLIC", got["visibility"])
	}

	dist, _ := got["distribution"].(map[string]any)
	if dist["feedDistribution"] != "MAIN_FEED" {
		t.Errorf("feedDistribution = %v, want MAIN_FEED", dist["feedDistribution"])
	}
	// Empty arrays, not null: LinkedIn rejects null for these.
	if arr, ok := dist["targetEntities"].([]any); !ok || len(arr) != 0 {
		t.Errorf("targetEntities = %v, want empty array", dist["targetEntities"])
	}
}

// Omitting content entirely is what keeps a text-only post valid; sending an
// empty object is not the same thing.
func TestCreatePostOmitsContentForTextOnly(t *testing.T) {
	raw := captureBody(t, CreatePostRequest{Author: "urn:li:person:a", Commentary: "hi"})
	if strings.Contains(raw, `"content"`) {
		t.Fatalf("text-only post must omit content entirely, got: %s", raw)
	}
}

func TestCreatePostIncludesAltTextWhenProvided(t *testing.T) {
	in := CreatePostRequest{Author: "urn:li:person:a", Commentary: "hi"}
	in.Content = &PostContent{}
	in.Content.Media.ID = "urn:li:image:xyz"
	in.Content.Media.AltText = "a bar chart"

	raw := captureBody(t, in)
	if !strings.Contains(raw, `"altText":"a bar chart"`) {
		t.Fatalf("altText missing from body: %s", raw)
	}
	if !strings.Contains(raw, `"id":"urn:li:image:xyz"`) {
		t.Fatalf("image urn missing from body: %s", raw)
	}
}

func TestCreatePostRejectsEmptyCommentary(t *testing.T) {
	c := &Client{BaseURL: "http://unused", Token: "t"}
	if _, err := c.CreatePost(context.Background(), CreatePostRequest{
		Author: "urn:li:person:a", Commentary: "   ",
	}); err == nil {
		t.Fatal("expected an error for empty commentary")
	}
}

func TestCreatePostRequiresAuthor(t *testing.T) {
	c := &Client{BaseURL: "http://unused", Token: "t"}
	if _, err := c.CreatePost(context.Background(), CreatePostRequest{
		Commentary: "hi",
	}); err == nil {
		t.Fatal("expected an error when author is unset")
	}
}

// Person ids are app-scoped: the same human has a different id per app, so the
// urn has to be built from this app's own userinfo response.
func TestAuthorURNBuiltFromUserinfoSub(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/userinfo" {
			t.Errorf("expected /v2/userinfo, got %s", r.URL.Path)
		}
		w.Write([]byte(`{"sub":"782bbtaQ","name":"Aswath"}`))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "t"}
	urn, err := c.AuthorURN(context.Background())
	if err != nil {
		t.Fatalf("author urn: %v", err)
	}
	if urn != "urn:li:person:782bbtaQ" {
		t.Fatalf("urn = %q, want urn:li:person:782bbtaQ", urn)
	}
}

func TestAuthorURNRejectsResponseWithoutSub(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name":"No Id"}`))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "t"}
	if _, err := c.AuthorURN(context.Background()); err == nil {
		t.Fatal("expected an error when sub is absent")
	}
}

// Images are two steps. finalizeUpload belongs to the Videos API and must
// never appear here.
func TestUploadImageIsTwoStepsAndReturnsUrn(t *testing.T) {
	var puts int
	// The signed uploadUrl lives on its own server, so it must exist before
	// the initializeUpload handler can reference it.
	upload := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		puts++
		w.WriteHeader(http.StatusCreated)
	}))
	defer upload.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/images" || r.Method != http.MethodPost {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if got := r.URL.Query().Get("action"); got != "initializeUpload" {
			t.Errorf("action query = %q, want initializeUpload", got)
		}
		w.Write([]byte(`{"value":{"uploadUrl":"` + upload.URL + `/put","image":"urn:li:image:abc","uploadUrlExpiresAt":123}}`))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "t"}
	ctx := WithAuthor(context.Background(), "urn:li:person:abc")

	urn, err := c.UploadImage(ctx, writeTempPNG(t), "alt text")
	if err != nil {
		t.Fatalf("upload image: %v", err)
	}
	if urn != "urn:li:image:abc" {
		t.Fatalf("urn = %q, want urn:li:image:abc", urn)
	}
	if puts != 1 {
		t.Fatalf("expected exactly 1 PUT, got %d (finalizeUpload must not be called)", puts)
	}
}

func TestDeletePostIsIdempotent(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		if r.Header.Get("X-RestLi-Method") != "DELETE" {
			t.Errorf("missing X-RestLi-Method: DELETE header")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "t"}
	if err := c.DeletePost(context.Background(), "urn:li:ugcPost:123"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	// URNs in path params must be URL-encoded.
	if !strings.Contains(gotPath, "urn%3Ali%3AugcPost%3A123") {
		t.Fatalf("urn not URL-encoded in path: %s", gotPath)
	}
}

func TestDeletePostSurfacesPermissionFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"ACCESS_DENIED"}`))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "t"}
	if err := c.DeletePost(context.Background(), "urn:li:ugcPost:123"); err == nil {
		t.Fatal("expected an error on 403")
	}
}

// --- helpers ---

// captureBody posts once against a recording server and returns the exact JSON
// bytes that hit the wire.
func captureBody(t *testing.T, in CreatePostRequest) string {
	t.Helper()
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		captured = string(b)
		w.Header().Set(PostIDHeader, "urn:li:ugcPost:1")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Token: "t"}
	if _, err := c.CreatePost(context.Background(), in); err != nil {
		t.Fatalf("create: %v", err)
	}
	if captured == "" {
		t.Fatal("no request body was captured")
	}
	return captured
}

func writeTempPNG(t *testing.T) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "img.png")
	if err := os.WriteFile(f, []byte("\x89PNG\r\n\x1a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return f
}