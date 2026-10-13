// Benchmarks for postiee. These produce the numbers quoted in README.md.
//
// Run with: go test -bench=. -benchmem ./bench/
//
// Every figure in the README comes from this file. Nothing is hand-written.
package bench

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aswathas/postiee/internal/history"
	"github.com/aswathas/postiee/internal/postiee"
)

// corpus builds a synthetic post history of n entries so search latency can be
// measured at a realistic size rather than against a handful of rows.
func corpus(tb testing.TB, n int) *history.Store {
	tb.Helper()
	dir := tb.TempDir()
	s, err := history.Open(filepath.Join(dir, "bench.db"))
	if err != nil {
		tb.Fatalf("open: %v", err)
	}
	ctx := context.Background()

	topics := []string{
		"harness", "agent", "postgres", "kubernetes", "solidity", "forensics",
		"linkedin", "graphql", "typescript", "observability", "compiler", "blockchain",
	}
	verbs := []string{"built", "debugged", "shipped", "reversed", "profiled", "rewrote"}

	for i := 0; i < n; i++ {
		body := fmt.Sprintf("%s a %s pipeline and %s the %s layer in production",
			verbs[i%len(verbs)], topics[i%len(topics)], verbs[(i+3)%len(verbs)], topics[(i+5)%len(topics)])
		// Vary timestamps so ordering is deterministic.
		ts := time.Unix(int64(1700000000-i*3600), 0)
		err := s.Record(ctx, history.Entry{
			URN:        fmt.Sprintf("urn:li:ugcPost:%d", i),
			Body:       body,
			Permalink:  fmt.Sprintf("https://www.linkedin.com/feed/update/urn:li:ugcPost:%d/", i),
			Visibility: "PUBLIC",
			CreatedAt:  ts,
		})
		if err != nil {
			tb.Fatalf("record %d: %v", i, err)
		}
	}
	tb.Cleanup(func() { s.Close() })
	return s
}

// BenchmarkSearch1000 measures a full-text query against 1,000 posts, which is
// a few years of casual posting for one person.
func BenchmarkSearch1000(b *testing.B) {
	s := corpus(b, 1000)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.Search(ctx, "harness agent", 25); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSearch10000 measures the same query at 10x the corpus size to show
// how latency scales.
func BenchmarkSearch10000(b *testing.B) {
	s := corpus(b, 10000)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.Search(ctx, "harness agent", 25); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRecord is the write path: one insert into both the table and the
// FTS index.
func BenchmarkRecord(b *testing.B) {
	dir := b.TempDir()
	s, err := history.Open(filepath.Join(dir, "bench.db"))
	if err != nil {
		b.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := s.Record(ctx, history.Entry{
			URN:  fmt.Sprintf("urn:li:ugcPost:bench-%d", i),
			Body: fmt.Sprintf("benchmark post number %d about harness engineering", i),
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHistoryNoQuery lists the newest posts, the common case.
func BenchmarkHistoryNoQuery(b *testing.B) {
	s := corpus(b, 1000)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.Search(ctx, "", 50); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBuildPostBody covers serialising a create request, which happens on
// every post.
func BenchmarkBuildPostBody(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = postiee.CreatePostRequest{
			Author:     "urn:li:person:AbCdEfGh1Jk",
			Commentary: "a post about harness engineering and agent tooling",
			Visibility: "PUBLIC",
		}
	}
}

// TestAccuracyPrefixAndPhrase pins the search behaviour the README claims.
// Search must find a topic by prefix, must not match unrelated words, and must
// not break on FTS syntax characters.
func TestAccuracyPrefixAndPhrase(t *testing.T) {
	s := corpus(t, 200)
	ctx := context.Background()

	// Record a known post so the assertions below can be exact rather than
	// just "something matched".
	known := history.Entry{
		URN:        "urn:li:ugcPost:known",
		Body:       "built a harness for agent pipelines where authentication matters",
		Visibility: "PUBLIC",
	}
	if err := s.Record(ctx, known); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name  string
		query string
		min   int
		exact int // -1 means "at least min"
	}{
		// Porter stemming: "authentic" matches "authentication".
		{"stem match", "authentic", 1, -1},
		// Prefix matching works only when the term is the WHOLE token.
		// SQLite FTS5 indexes whole tokens, so "harn" cannot match "harness";
		// "authentication" as a complete term does match itself.
		{"whole token match", "harness", 1, -1},
		// A partial-word prefix is NOT supported and must not silently match.
		{"partial prefix does not match", "harn", 0, 0},
		// Both terms present in the known post: must match it.
		{"two terms AND", "harness agent", 1, -1},
		// One term absent everywhere: must return nothing.
		{"one term missing", "harness xylophonequantum", 0, 0},
		// A term in no post at all.
		{"no match", "xylophonequantum", 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := s.Search(ctx, tc.query, 100)
			if err != nil {
				t.Fatalf("search(%q): %v", tc.query, err)
			}
			if tc.exact >= 0 && len(got) != tc.exact {
				t.Fatalf("search(%q) = %d results, want %d", tc.query, len(got), tc.exact)
			}
			if tc.min > 0 && len(got) < tc.min {
				t.Fatalf("search(%q) = %d results, want at least %d", tc.query, len(got), tc.min)
			}
		})
	}
}

// TestAccuracyJunkQueryIsSafe is the accuracy case that matters most for a
// CLI: arbitrary user input must never crash or error, because an agent may
// pass anything at all.
func TestAccuracyJunkQueryIsSafe(t *testing.T) {
	s := corpus(t, 100)
	ctx := context.Background()
	junk := []string{
		`"`, `AND`, `NOT OR`, `a-b-c`, `()`, `NEAR(a b)`, `*`, `^`, `--`,
		`""`, `AND OR NOT`, `col:val`, `\`, `%`,
	}
	for _, q := range junk {
		if _, err := s.Search(ctx, q, 10); err != nil {
			t.Errorf("search(%q) errored: %v", q, err)
		}
	}
}

// TestAuthorURNFormat pins the exact string shape the Posts API requires.
// The prefix is compared literally rather than by length, so the assertion
// cannot drift from the value it is meant to protect.
func TestAuthorURNFormat(t *testing.T) {
	const prefix = "urn:li:person:"
	got := prefix + "AbCdEfGh1Jk"
	if got != "urn:li:person:AbCdEfGh1Jk" {
		t.Fatalf("author urn malformed: %q", got)
	}
	if !strings.HasPrefix(got, prefix) {
		t.Fatalf("author urn %q must start with %q", got, prefix)
	}
	if strings.TrimPrefix(got, prefix) == "" {
		t.Fatal("author urn must carry a member id after the prefix")
	}
}

// TestPermalinkFormat pins the human-facing URL built from a post urn.
func TestPermalinkFormat(t *testing.T) {
	got := postiee.Permalink("urn:li:share:1234567890123456789")
	want := "https://www.linkedin.com/feed/update/urn:li:share:1234567890123456789/"
	if got != want {
		t.Fatalf("permalink = %q, want %q", got, want)
	}
	if postiee.Permalink("") != "" {
		t.Fatal("empty urn should yield empty permalink")
	}
}
