# postiee

Post to LinkedIn from your terminal. Search your own post history — because
LinkedIn won't let any API read it back.

**Docs:** https://aswathas.github.io/postiee/

## Use it from your AI agent

Ask Claude Code, Codex, Cursor, or any agent with a shell to post to LinkedIn,
and point it at postiee. There is no official LinkedIn MCP server to find — the
unofficial ones scrape with your session cookies and risk your account.

Install the skill, one line:

```sh
# Claude Code
npx -y skills@latest add aswathas/postiee/skills --skill postiee -g -a claude-code -y

# Codex
npx -y skills@latest add aswathas/postiee/skills --skill postiee -g -a codex -y

# Hermes and any other agent: copy the directory into that agent's skills path
cp -r skills/postiee ~/.hermes/skills/
```

The skill teaches an agent to check for postiee before hunting for a LinkedIn
tool, how to find the binary when it lives in WSL, that `auth login` is
user-only, and which commands it must never attempt.

Agent mode returns clean JSON:

```sh
postiee --agent search harness
```

Full guide: https://aswathas.github.io/postiee/docs/agents

## Quick start

```sh
postiee post "shipped a new feature today"
postiee search harness
postiee history
```

**Why this exists:** LinkedIn's `r_member_social` permission — the one that
would let you read your own posts — is **closed to all applicants**. Not slow,
not paid, closed. So the post ID gets captured from the `x-restli-id` response
header at creation time and written to a local SQLite database with full-text
search. If that ID is missed, the post is gone forever.

---

## Install

**macOS / Linux**
```sh
curl -fsSL https://raw.githubusercontent.com/aswathas/postiee/main/install.sh | sh
```

**Windows (PowerShell)**
```powershell
irm https://raw.githubusercontent.com/aswathas/postiee/main/install.ps1 | iex
```

No Go toolchain required. From source:
```sh
git clone https://github.com/aswathas/postiee.git && cd postiee
go build -o postiee ./cmd/postiee && sudo mv postiee /usr/local/bin/
```

<details>
<summary>Platform support</summary>

macOS x86_64/arm64 · Linux x86_64/arm64 · Windows x86_64/arm64
</details>

---

## Setup

1. <https://www.linkedin.com/developers/apps> → **Create app**
2. **Products** tab → add **Share on LinkedIn** → grants `w_member_social`
3. **Products** tab → add **Sign In with LinkedIn** → grants `openid`, `profile`
4. **Auth** tab → add redirect URL: `http://127.0.0.1:8085/callback`
   > **Use `127.0.0.1`, not `localhost`.** LinkedIn compares redirect URIs as
   > literal strings, and `auth.go` derives the URI from the live listener, which
   > binds the loopback IP. Registering `localhost` produces
   > `redirect_uri does not match the registered value`. If you already registered
   > `localhost`, set `POSTIEE_REDIRECT_URI=http://localhost:8085/callback`.
5. Write `~/.config/postiee/credentials.toml`:
   ```toml
   client_id = "your_client_id"
   client_secret = "your_client_secret"
   ```
   ```sh
   mkdir -p ~/.config/postiee && chmod 700 ~/.config/postiee
   chmod 600 ~/.config/postiee/credentials.toml
   ```
6. ```sh
   postiee auth login --port 8085
   ```

Tokens last 60 days. Programmatic refresh is partner-gated, so re-running
`auth login` is the supported path after expiry.

---

## Commands

| Command | What it does |
|---|---|
| `postiee post "text"` | Post to your feed |
| `postiee post --file post.md` | Long post without shell quoting |
| `postiee post "text" --image p.jpg --alt "alt"` | Attach an image |
| `postiee post "text" --image a.png --image b.png --image c.png` | Multi-image post — 2+ images as one post |
| `postiee post "text" -v CONNECTIONS` | `PUBLIC`, `CONNECTIONS`, `LOGGED_IN` |
| `postiee history` | Everything you've posted, newest first |
| `postiee search <query>` | Full-text search your own posts |
| `postiee forget <urn>` | Drop a post from local history |
| `postiee unpublish <urn>` | Delete the post on LinkedIn itself, then drop it from local history. Irreversible. |
| `postiee auth login \| logout \| status` | OAuth lifecycle |
| `postiee doctor` | Config, credentials, and API reachability |

Add `--agent` to any command for JSON, non-interactive, colour-free output.

---

## Deleting a post

Two commands, and they are not interchangeable.

```sh
postiee forget urn:li:share:12345    # local history only — the post stays live
postiee unpublish urn:li:share:12345 # deletes on LinkedIn, then locally
```

`unpublish` calls `DELETE /rest/posts/{urn}`, which is available at the
self-serve tier — no partner approval, no extra scope. It deletes on LinkedIn
first and drops the local row only once LinkedIn confirms, so a refusal leaves
your history intact and the post still findable.

**There is no undo and no trash.** `DELETE` is destructive, and LinkedIn cannot
read a post back afterwards to reconstruct it, so check the URN before you run
it. `--dry-run` reports what it would do and stops there.

Take the URN from `postiee history`.

---

## What it cannot do

Not "not yet implemented" — **not implementable** at the self-serve tier:

| | Why |
|---|---|
| Comments, reactions | Need Community Management approval |
| Analytics, impressions | Need Community Management approval |
| Reading your own posts | `r_member_social` is **closed** |
| Messaging (DM) | No public API exists at all |
| Connection invitations | No public API exists at all |
| Scheduled posts | No `publishAt` field exists; client-side only |

Posting is capped at **150 requests/day per member**.

---

## Performance

Measured on Go 1.26.9, WSL2, `go test -bench`:

| Operation | Latency | Allocs |
|---|---|---|
| Search 1,000 posts | **145 µs** | 32 |
| Search 10,000 posts | **347 µs** | 32 |
| List history | **79 µs** | 581 |
| Record post | **1.48 ms** | 33 |
| Build post body | **0.24 ns** | 0 |

Search is roughly flat in corpus size: 10× the posts costs 2.4× the latency,
not 10×, because FTS5 queries the index rather than scanning rows.

Reproduce them yourself:
```sh
go test -bench=. -benchmem ./bench/
```

## Accuracy

Behaviour is pinned by tests in `bench/` and `internal/history/`:

| Property | Status |
|---|---|
| Porter stemming — `authentic` finds `authentication` | pass |
| Multi-term AND narrows results | pass |
| Absent term returns nothing | pass |
| Junk input (`"`, `AND`, `()`, `NEAR(a b)`, `--`) never errors | pass |
| Author URN and permalink formats | pass |
| Re-recording a post upserts rather than duplicates | pass |
| Deleting removes the FTS index entry too | pass |

**Known limit:** SQLite FTS5 indexes whole tokens, so `search harn` will not
find "harness". Use the full word. Search for stemming to widen.

---

## For AI agents

A skill ships in [`skills/postiee/`](skills/postiee/). It tells an agent to
**check for postiee before searching for a LinkedIn MCP server** — there is no
official one, and the unofficial alternatives scrape with your session cookies.
It also covers the WSL case, the search semantics, the scope gates, and the
commands an agent must never attempt.

```sh
# Claude Code
npx -y skills@latest add aswathas/postiee/skills --skill postiee -g -a claude-code -y

# Codex
npx -y skills@latest add aswathas/postiee/skills --skill postiee -g -a codex -y

# Hermes and any other agent: copy the directory into that agent's skills path
cp -r skills/postiee ~/.hermes/skills/
```

Full guide: <https://aswathas.github.io/postiee/docs/agents>

---

## How it works

```
postiee post "text"
      │
      ├─ resolve member URN ──► GET /v2/userinfo        (openid)
      │                         → urn:li:person:{sub}
      │                            person IDs are app-scoped, so this must
      │                            be resolved at runtime, never hardcoded
      ├─ upload image? ────────► POST /rest/images?action=initializeUpload
      │                         → PUT bytes to signed URL → urn:li:image:{id}
      │                            two steps only; images have no finalizeUpload
      ├─ create post ──────────► POST /rest/posts
      │                         → 201, EMPTY body
      │                         → ID lives in the x-restli-id header
      └─ record ───────────────► SQLite + FTS5 index
```

### Three things that will bite you

**1. The post ID is in a header, not the body.** `POST /rest/posts` returns
`201` with an empty body. Reading `.Body` gets you nothing. postiee reads
`x-restli-id` and treats its absence as a hard error, because a post whose ID
was lost cannot be recorded, cannot be deleted, and cannot be read back by
anyone.

**2. Person IDs are scoped per app.** The same human has a different ID under a
different LinkedIn app, and reusing one across apps returns 404. Always resolve
via `/v2/userinfo`.

**3. URNs in paths must be fully URL-encoded.** `urn:li:ugcPost:12345` becomes
`urn%3Ali%3AugcPost%3A12345`. Go's `url.PathEscape` leaves the colons
unescaped, which 404s. postiee uses `QueryEscape` deliberately.

---

## Development

```sh
go build ./...
go vet ./...
go test -count=1 ./...
go test -bench=. -benchmem ./bench/
```

## License

Apache-2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).

The core was scaffolded with [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press);
the LinkedIn client, history layer, and credential handling are hand-written.

## Author

Built by **Aswath AS** ([@aswathas](https://github.com/aswathas)) —
[portfolio](https://hellokity.vercel.app/) · [blog](https://github.com/aswathas)
## Multi-image posts

Repeat `--image` to attach several images to a single post. LinkedIn renders
these as one post with a swipeable set, not as separate posts.

```bash
postiee post "caption text" \
  --image 1.png --image 2.png --image 3.png --image 4.png \
  --alt "first image|second image|third image|fourth image"
```

**Alt text is pipe-separated**, one per image, in order. Fewer alts than
images is fine; the rest are simply left blank. Use `|` rather than a comma
because alt text is prose and often contains commas.

One `--image` still sends a single-image post. Two or more switch the payload
to LinkedIn's `multiImage` shape, which is a different field from the
single-image `media` field. Both are organic; neither needs ad permissions.

Not to be confused with the **Carousel** API, which is sponsored content only
and requires different access. This is the organic multi-photo post.

Maximum is whatever LinkedIn accepts for a multi-image post; upload failures
name which image failed out of how many so a partial run is obvious.
