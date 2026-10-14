// Copyright 2026 aswathas and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"bytes"
	"strings"
	"testing"
)

// --dry-run must not reach the network. A guard that is added after the
// CreatePost call, or omitted entirely, publishes a real post and records it,
// which is unrecoverable: LinkedIn has no self-serve read-back to undo it.
func TestPostDryRunDoesNotPost(t *testing.T) {
	// An unusable token makes a regression loud instead of live: if the guard
	// is ever removed, the call fails on the author lookup rather than
	// publishing. Never run this test against a working credential.
	t.Setenv("POSTIEE_TOKEN", "dry-run-sentinel-not-a-real-token")

	var out bytes.Buffer
	root := RootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"post", "a post that must never go live", "--dry-run"})

	if err := root.Execute(); err != nil {
		t.Fatalf("post --dry-run returned error: %v", err)
	}
	if !strings.Contains(out.String(), "dry-run: would") {
		t.Fatalf("post --dry-run did not short-circuit, output: %q", out.String())
	}
}

func TestForgetDryRunDoesNotDelete(t *testing.T) {
	var out bytes.Buffer
	root := RootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"forget", "urn:li:ugcPost:12345", "--dry-run"})

	if err := root.Execute(); err != nil {
		t.Fatalf("forget --dry-run returned error: %v", err)
	}
	if !strings.Contains(out.String(), "dry-run: would") {
		t.Fatalf("forget --dry-run did not short-circuit, output: %q", out.String())
	}
}