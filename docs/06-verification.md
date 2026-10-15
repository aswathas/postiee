# 06 — Verification

**Date.** 2026-10-14, with additions 2026-10-16.

## What passed

**Benchmarks.** `bench/main_test.go` produces every number quoted in the README, and
`go test -bench=. -benchmem ./bench/` reproduces them. Measured on Go 1.26.9, WSL2:

| Operation | Latency | Allocs |
|---|---|---|
| Search 1,000 posts | 145 µs | 32 |
| Search 10,000 posts | 347 µs | 32 |
| List history | 79 µs | 581 |
| Record post | 1.48 ms | 33 |

Search is roughly flat in corpus size — 10× the posts costs 2.4× the latency, not
10× — because FTS5 queries the index rather than scanning rows. That is the property
worth having, and it is measured rather than asserted.

**Accuracy tests.** Porter stemming (`authentic` finds `authentication`), multi-term
AND narrowing, absent terms returning nothing, junk input (`"`, `AND`, `()`, `NEAR(a
b)`, `--`) never erroring, urn and permalink formats, re-recording upserting rather
than duplicating, deletion removing the FTS index entry too.

**Deletion, verified live.** Post a test post, delete it with the self-serve token,
get 204. This is the one check that turned a documented impossibility into a shipped
command.

**Eleven accessibility and layout defects, found and fixed.** From `19caecc`, six
`impeccable` passes ran in parallel against disjoint file sets. Every finding named
there is real:

- Setup docs told users to register `localhost`, but `auth.go:241` sends `127.0.0.1`,
  citing RFC 8252. **Following the documentation produced exactly the failure the
  page's own warning described.**
- The Windows token path in the docs was invented; `paths.go` has no Windows branch.
- Hero headline wrapped to three lines at every viewport — `16ch` max-width against a
  20-character line — then orphaned a word at 375px.
- The terminal's last output animated in at t=0, before the first command was typed.
  `nth-of-type` counts divs, so position 6 is an output and the selector matched
  nothing.
- `prefers-reduced-motion` set `animation-duration` but not `animation-delay`, so
  `backwards` fill-mode kept the hero invisible for ~550ms.
- Horizontal page scroll at 320px on every docs page.
- Touch targets under 44px across nav, copy button and sidebar.
- SVG diagrams used GitHub's light palette on a black canvas; one line overflowed its
  viewBox because SVG text does not wrap.
- Four unused CSS tokens, dead declarations, a dead `#troubleshooting` anchor, and a
  "guaranteed to work" table containing a row that is not.

That first one is the important one. The documentation and the code disagreed, and
the documentation was wrong in the direction that guaranteed a user failure.

**The 204 idempotency limit is documented rather than hidden.** Deleting an
already-deleted post also returns 204, so 204 alone does not prove the post was
there. Both the README and the skill say so.

## What failed

**The documented installer does not work.** `install.ps1` 404s because no release has
ever been cut. See `attempts/002`. Found 2026-10-16, one day after the installer
shipped, by running it.

**A parallel agent left a nested git repository in the tree.** A worktree created
inside `.claude/` was committed as an embedded repository. Removed from the index and
ignored in `2e47fe6`.

**Two-minute OAuth timeout was shorter than a human login.** `cea3964` extended it to
fifteen:

> "A LinkedIn login can involve a password, a 2FA prompt, and an email code. Two
> minutes is not enough for a human to get through it, and the timeout discarded the
> flow entirely rather than letting it be retried."

The second half is the real defect: the failure was destructive rather than
recoverable.

**A claimed permanent limitation was wrong.** See `attempts/001`.

**Live account ids leaked into benchmark fixtures.** Caught before the commit landed.
The values remain in git history. See `attempts/003`.

## Still untested

**The one-line install, on every platform.** Never executed. There is no CI job that
downloads the artifact the installers point at, which is why a workflow with no
release went unnoticed.

**The generated tree under a reprint.** The `.printing-press-patches/` records are
validated at regen and publish time, but no reprint has been performed since the
patches were added. The fail-closed rule is documented and unexercised.

**Cross-platform behaviour.** Go builds are verified by the workflow's existence, not
by running the binary on macOS, Linux, or a second Windows machine. The binary in
active use was built on WSL2 and is Linux. `postiee --version` reports `0.0.0-dev`,
which the release ledger says is expected for a fresh print — but it also means
nothing distinguishes a dev build from a released one at runtime.

**The MCP server.** `postiee-pp-mcp` exists in both transports. Neither has been
exercised end to end. The skill recommends shell access over MCP, which means the
less-preferred path is also the less-tested one.

**Rate limit behaviour.** The 150-requests-per-day cap is documented and nothing
tests what happens at it.

**Concurrency.** Two agents posting through the same local SQLite history database is
not tested, and FTS5 writes take a lock.

**Long-term token refresh.** Tokens expire at sixty days and programmatic refresh is
partner-gated, so the supported path is a manual `auth login`. No test covers the
expiry boundary, because it is sixty days away and has not been reached yet.