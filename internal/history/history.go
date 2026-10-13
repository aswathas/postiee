// Package history keeps a local, searchable record of every post made through
// postiee.
//
// This is the point of the tool. LinkedIn's `r_member_social` permission is
// CLOSED — not accepting requests — so the API will never let you read your own
// posts back. LinkedIn's own data export is a static archive on a 72-hour
// download window, desktop only. So the post urn captured from the x-restli-id
// header at create time is the only record that will ever exist, and it has to
// land somewhere durable and searchable.
package history

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	_ "modernc.org/sqlite" // pure-Go driver: no cgo, so cross-compiles cleanly
)

// Entry is one post postiee created.
type Entry struct {
	URN        string    `json:"urn"`
	Body       string    `json:"body"`
	Permalink  string    `json:"permalink"`
	MediaURN   string    `json:"media_urn,omitempty"`
	Visibility string    `json:"visibility"`
	CreatedAt  time.Time `json:"created_at"`
}

// Store is the SQLite-backed history.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) the history database at the given path.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("creating history directory: %w", err)
	}
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// DefaultPath is where history lives unless overridden.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "postiee-history.db"
	}
	return filepath.Join(home, ".config", "postiee", "history.db")
}

func (s *Store) migrate() error {
	// FTS5 virtual table for full-text search over post bodies. An external
	// content table would complicate writes for no gain at this size; a plain
	// standalone FTS index is enough and keeps writes transactional.
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS posts (
			post_urn    TEXT PRIMARY KEY,
			body        TEXT NOT NULL,
			permalink   TEXT NOT NULL DEFAULT '',
			media_urn   TEXT NOT NULL DEFAULT '',
			visibility  TEXT NOT NULL DEFAULT 'PUBLIC',
			created_at  INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS posts_created_at ON posts(created_at DESC)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS posts_fts USING fts5(
			post_urn UNINDEXED,
			body,
			tokenize = 'porter unicode61'
		)`,
	}
	for _, q := range stmts {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("migrating history: %w", err)
		}
	}
	return nil
}

// Record writes a post into history. Safe to call twice for the same urn: the
// primary key makes it an upsert, so a retried post does not duplicate.
func (s *Store) Record(ctx context.Context, e Entry) error {
	if e.URN == "" {
		return fmt.Errorf("cannot record a post with no urn")
	}
	ts := e.CreatedAt
	if ts.IsZero() {
		ts = time.Now()
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO posts (post_urn, body, permalink, media_urn, visibility, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(post_urn) DO UPDATE SET
		   body = excluded.body,
		   permalink = excluded.permalink,
		   media_urn = excluded.media_urn,
		   visibility = excluded.visibility`,
		e.URN, e.Body, e.Permalink, e.MediaURN, e.Visibility, ts.Unix(),
	); err != nil {
		return fmt.Errorf("recording post: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO posts_fts (post_urn, body) VALUES (?, ?)`, e.URN, e.Body,
	); err != nil {
		return fmt.Errorf("indexing post: %w", err)
	}

	return tx.Commit()
}

// Search runs a full-text query over post bodies, newest first. An empty query
// lists everything, which is what `history` with no arguments does.
func (s *Store) Search(ctx context.Context, query string, limit int) ([]Entry, error) {
	if limit <= 0 {
		limit = 50
	}

	var (
		rows *sql.Rows
		err  error
	)
	if strings.TrimSpace(query) == "" {
		rows, err = s.db.QueryContext(ctx,
			`SELECT post_urn, body, permalink, media_urn, visibility, created_at
			 FROM posts ORDER BY created_at DESC LIMIT ?`, limit)
	} else {
		// Quote the query so user input is treated as a phrase, not FTS syntax.
		// Without this a stray quote or a bare AND breaks the parse.
		rows, err = s.db.QueryContext(ctx,
			`SELECT p.post_urn, p.body, p.permalink, p.media_urn, p.visibility, p.created_at
			 FROM posts_fts f
			 JOIN posts p ON p.post_urn = f.post_urn
			 WHERE posts_fts MATCH ?
			 ORDER BY p.created_at DESC LIMIT ?`, FTSQuery(query), limit)
	}
	if err != nil {
		return nil, fmt.Errorf("searching history: %w", err)
	}
	defer rows.Close()

	var out []Entry
	for rows.Next() {
		var e Entry
		var ts int64
		if err := rows.Scan(&e.URN, &e.Body, &e.Permalink, &e.MediaURN, &e.Visibility, &ts); err != nil {
			return nil, err
		}
		e.CreatedAt = time.Unix(ts, 0)
		out = append(out, e)
	}
	return out, rows.Err()
}

// Get returns a single post by urn, or ok=false when absent.
func (s *Store) Get(ctx context.Context, urn string) (Entry, bool, error) {
	var e Entry
	var ts int64
	err := s.db.QueryRowContext(ctx,
		`SELECT post_urn, body, permalink, media_urn, visibility, created_at
		 FROM posts WHERE post_urn = ?`, urn).
		Scan(&e.URN, &e.Body, &e.Permalink, &e.MediaURN, &e.Visibility, &ts)
	if err == sql.ErrNoRows {
		return Entry{}, false, nil
	}
	if err != nil {
		return Entry{}, false, err
	}
	e.CreatedAt = time.Unix(ts, 0)
	return e, true, nil
}

// Forget removes a post and its index entry. Used after a delete.
func (s *Store) Forget(ctx context.Context, urn string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM posts WHERE post_urn = ?`, urn); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM posts_fts WHERE post_urn = ?`, urn); err != nil {
		return err
	}
	return nil
}

// Count returns how many posts are recorded.
func (s *Store) Count(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts`).Scan(&n)
	return n, err
}

// Close releases the database handle.
func (s *Store) Close() error { return s.db.Close() }

// FTSQuery turns free user input into a safe FTS5 MATCH expression.
//
// Bare input is risky: a quote, a hyphen, or the word AND/NOT/OR can produce a
// syntax error that surfaces as a confusing "malformed MATCH expression".
//
// Each term is emitted as a quoted literal, which stops FTS5 from parsing it as
// syntax. Reserved words ("AND", "NOT", "OR") are the reason quoting matters:
// unquoted they are operators. Prefix matching needs the `*` outside the
// closing quote — `"post"*` is a valid prefix match under FTS5, while a bare
// `AND*` is a syntax error — so quoting the term and appending the star gives
// both safety and prefix behaviour.
func FTSQuery(q string) string {
	fields := strings.Fields(q)
	if len(fields) == 0 {
		return `""`
	}
	terms := make([]string, 0, len(fields))
	for _, f := range fields {
		// Keep only characters that cannot alter the MATCH grammar.
		var b strings.Builder
		for _, r := range f {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
				b.WriteRune(r)
			}
		}
		cleaned := b.String()
		if cleaned == "" {
			continue
		}
		terms = append(terms, `"`+cleaned+`"*`)
	}
	if len(terms) == 0 {
		return `""`
	}
	return strings.Join(terms, " AND ")
}
