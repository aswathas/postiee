package history

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func openTemp(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "history.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestRecordAndSearch(t *testing.T) {
	ctx := context.Background()
	s := openTemp(t)

	if err := s.Record(ctx, Entry{
		URN: "urn:li:ugcPost:1", Body: "built a linkedin posting cli",
		Permalink:  "https://linkedin.com/feed/update/urn:li:ugcPost:1/",
		Visibility: "PUBLIC",
	}); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := s.Record(ctx, Entry{
		URN: "urn:li:ugcPost:2", Body: "wrote about gardening instead",
	}); err != nil {
		t.Fatalf("record: %v", err)
	}

	hits, err := s.Search(ctx, "linkedin", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) != 1 || hits[0].URN != "urn:li:ugcPost:1" {
		t.Fatalf("search hits = %+v, want only ugcPost:1", hits)
	}
}

// An empty query lists everything, newest first. This is what `history` with no
// arguments does.
func TestEmptyQueryListsNewestFirst(t *testing.T) {
	ctx := context.Background()
	s := openTemp(t)

	older := Entry{URN: "urn:li:ugcPost:old", Body: "first", CreatedAt: time.Now().Add(-time.Hour)}
	newer := Entry{URN: "urn:li:ugcPost:new", Body: "second", CreatedAt: time.Now()}
	s.Record(ctx, older)
	s.Record(ctx, newer)

	all, err := s.Search(ctx, "", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("got %d entries, want 2", len(all))
	}
	if all[0].URN != newer.URN {
		t.Fatalf("newest first violated: got %s first", all[0].URN)
	}
}

// Re-recording the same urn must not duplicate: a retried post would otherwise
// show twice in history and make the record a lie.
func TestRecordIsIdempotentPerURN(t *testing.T) {
	ctx := context.Background()
	s := openTemp(t)

	for i := 0; i < 3; i++ {
		if err := s.Record(ctx, Entry{URN: "urn:li:ugcPost:1", Body: "same post"}); err != nil {
			t.Fatalf("record %d: %v", i, err)
		}
	}
	n, err := s.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("count = %d, want 1 (upsert, not insert)", n)
	}
}

func TestRecordRejectsEmptyURN(t *testing.T) {
	s := openTemp(t)
	if err := s.Record(context.Background(), Entry{Body: "orphan"}); err == nil {
		t.Fatal("expected an error recording a post with no urn")
	}
}

func TestGetAndForget(t *testing.T) {
	ctx := context.Background()
	s := openTemp(t)
	s.Record(ctx, Entry{URN: "urn:li:ugcPost:1", Body: "delete me"})

	e, ok, err := s.Get(ctx, "urn:li:ugcPost:1")
	if err != nil || !ok {
		t.Fatalf("get: ok=%v err=%v", ok, err)
	}
	if e.Body != "delete me" {
		t.Fatalf("body = %q", e.Body)
	}

	if err := s.Forget(ctx, "urn:li:ugcPost:1"); err != nil {
		t.Fatalf("forget: %v", err)
	}
	// The FTS index must be cleaned up too, or the post still shows in search.
	hits, err := s.Search(ctx, "delete", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("search still returns %d rows after forget", len(hits))
	}
}

// Bare FTS syntax from user input must not blow up the query.
func TestSearchTolerantOfPunctuation(t *testing.T) {
	ctx := context.Background()
	s := openTemp(t)
	s.Record(ctx, Entry{URN: "urn:li:ugcPost:1", Body: "hybrid cloud work"})

	for _, q := range []string{`"`, `AND`, `NOT OR`, `a-b-c`, `()`, `--`} {
		if _, err := s.Search(ctx, q, 10); err != nil {
			t.Errorf("search(%q) errored: %v", q, err)
		}
	}
}

// Prefix matching means a partial word still finds the post.
func TestSearchMatchesPrefix(t *testing.T) {
	ctx := context.Background()
	s := openTemp(t)
	s.Record(ctx, Entry{URN: "urn:li:ugcPost:1", Body: "authentication is hard"})

	hits, err := s.Search(ctx, "auth", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("prefix search found %d hits, want 1", len(hits))
	}
}

func TestSearchRespectsLimit(t *testing.T) {
	ctx := context.Background()
	s := openTemp(t)
	for _, u := range []string{"a", "b", "c", "d"} {
		s.Record(ctx, Entry{URN: "urn:li:ugcPost:" + u, Body: "post " + u})
	}
	hits, err := s.Search(ctx, "post", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Fatalf("limit ignored: got %d, want 2", len(hits))
	}
}

func TestFTSQueryQuotesTerms(t *testing.T) {
	// Operator words must be quoted so they search as literal text instead of
	// being parsed as FTS5 syntax. `"AND"*` is a literal prefix-match on the
	// word AND; only the bare separators between terms are operators.
	got := FTSQuery(`hello AND "world`)
	if !strings.Contains(got, `"AND"`) {
		t.Errorf("operator word should be quoted as a literal term, got %q", got)
	}
	if !strings.Contains(got, `"hello"`) {
		t.Errorf("expected quoted term, got %q", got)
	}
	if !strings.Contains(got, `"world"`) {
		t.Errorf("expected quotes stripped and term quoted, got %q", got)
	}
	// A raw quote in user input must not survive into the expression.
	if strings.Contains(got, `""`) && !strings.HasPrefix(got, `""`) {
		t.Errorf("empty quoted term leaked into expression: %q", got)
	}
	if FTSQuery("   ") != `""` {
		t.Errorf("blank query should be inert, got %q", FTSQuery("   "))
	}
}
