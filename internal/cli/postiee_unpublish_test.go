package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aswathas/postiee/internal/postiee"
)

// LinkedIn returns 204 for DELETE whether or not the post existed, so a 204 is
// the only success signal. If the handler does not see the required headers, the
// request is malformed and LinkedIn rejects it in production.
func TestDeletePostSendsRequiredHeaders(t *testing.T) {
	var gotMethod, gotRestLi, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotRestLi = r.Header.Get("X-RestLi-Method")
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := &postiee.Client{BaseURL: srv.URL, Token: "token"}
	if err := c.DeletePost(context.Background(), "urn:li:share:12345"); err != nil {
		t.Fatalf("DeletePost() error = %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	// Without this header LinkedIn treats the call as malformed.
	if gotRestLi != "DELETE" {
		t.Errorf("X-RestLi-Method = %q, want DELETE", gotRestLi)
	}
	// urn:li:share:12345 must arrive percent-encoded; a bare colon 404s.
	if strings.Contains(gotPath, ":") {
		t.Errorf("path %q contains an unencoded colon", gotPath)
	}
	if !strings.Contains(gotPath, "%3A") {
		t.Errorf("path %q is not percent-encoded", gotPath)
	}
}

// A 403 from LinkedIn must surface as an error rather than a silent success,
// otherwise unpublish would remove the local record for a post still live.
func TestDeletePostFailsLoudlyOn403(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "ACCESS_DENIED", http.StatusForbidden)
	}))
	defer srv.Close()

	c := &postiee.Client{BaseURL: srv.URL, Token: "token"}
	err := c.DeletePost(context.Background(), "urn:li:share:12345")
	if err == nil {
		t.Fatal("DeletePost() returned nil on 403; the post would still be live")
	}
	if !strings.Contains(err.Error(), "delete failed") {
		t.Errorf("error = %v, want it to mention the failed delete", err)
	}
}
