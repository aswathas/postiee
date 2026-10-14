# Why I built a CLI to post on LinkedIn — and how it actually works

## The problem

I have things I want to post. I almost never post them.

That's not a content problem. I have plenty to say. It's a friction problem.

Posting on LinkedIn means: open a browser. Find the box. Paste text. Wait for the image to upload. Check the character counter. Hit post.

That's six steps. Each one is small. Together they're enough that the whole thing quietly dies, and the idea goes into a folder called "things I meant to post."

I know I'm not special here. Everyone has that folder.

The interesting part is *why* the obvious tools don't fix it.

## Why the obvious tools don't work

I went looking for something that posts to LinkedIn from the command line. Here's what exists:

- **Scheduling dashboards** (Buffer, Hootsuite, Taplio). Hootsuite starts around $99/user/month to schedule posts on a platform that gives you a scheduler for free.
- **Data export.** LinkedIn will give you an archive of your posts. It's a one-time request, desktop only, with a 72-hour download window. Not searchable. Not ongoing.
- **Browser automation wrappers.** Some tools drive a real browser session for you. That works until it doesn't — one login prompt and your account is in someone else's hands.

None of these make posting *one step*. They add steps.

Then I found something that changed the whole picture.

## The detail that made this necessary

LinkedIn has a permission called `r_member_social`.

Its only job is to let an app read your own posts back to you.

It is **closed to all applicants**. Not a slow review queue. Not a paid tier. Closed, permanently, to everyone including individual developers building their own tools.

Think about what that means. You post something. Then you want to find it later. And there is **no API on earth — including mine — that can read it back.**

So the post ID that LinkedIn hands you at creation time is the only handle that will ever exist for that post. If you don't capture it at the exact moment of creation, that post is gone forever. Unsearchable. Unlinkable. Unfindable through any official route.

That turned a convenience feature into the whole product:

**A post you made through postiee is stored on your own machine. That's the only copy that will ever exist.**

---

## What I built

A single command-line tool called postiee.

Post:

```sh
postiee post "shipped a new thing today"
```

Find it later:

```sh
postiee search harness
postiee history
```

That's the core loop. One command to post, one to find it.

Underneath:

- A single static binary, written in Go
- Runs on macOS, Linux, and Windows — both Intel and ARM
- No runtime. No daemon. No database server. No dependencies
- Standard OAuth login against LinkedIn
- Local storage is SQLite with FTS5 full-text search
- JSON output mode for tools that parse it
- An MCP server and an installable skill for AI agents

---

## How it works under the hood

When you run `postiee post "text"`, four things happen in order.

### 1. It figures out who you are

```sh
GET /v2/userinfo
```

This returns something like `urn:li:person:abc123`.

There's a trap here. **LinkedIn person IDs are different for every app.** The same human has a different ID under a different LinkedIn application. If you hardcode an ID, or reuse one across apps, LinkedIn returns a 404.

So postiee always resolves this at runtime.

### 2. It uploads the image, if there is one

Two steps, no more:

```sh
POST /rest/images?action=initializeUpload
PUT <signed URL>        # the actual bytes
```

The result is `urn:li:image:{id}`.

Images have no finalize step. That's why postiee's test suite asserts this upload performs **exactly two API calls**. Three would mean something got refactored that shouldn't have been.

### 3. It creates the post

```sh
POST /rest/posts
→ 201 Created
→ empty body
```

And here's the second trap, the one that defines this whole project.

**The response body is empty.** The post ID does not live in the JSON. It lives in a response header called `x-restli-id`.

Most integrations read `.Body` and get nothing back. postiee reads the header, and treats a missing header as a hard error — because a post whose ID was lost cannot be recorded, cannot be searched, and cannot be recovered by anyone, ever.

### 4. It writes it to your machine

Post ID, text, author URN, permalink, timestamp — into SQLite, with an FTS5 index for search.

That's it. That's the whole flow.

### One more trap worth knowing

If a post URN appears in a URL path, it has to be fully encoded:

```
urn:li:ugcPost:12345
→ urn%3Ali%3AugcPost%3A12345
```

Go's `url.PathEscape` leaves colons unescaped, which causes a 404. postiee uses `QueryEscape` deliberately, and has a test pinning it.

---

## Setting it up

Honestly, this is the easy part.

### Step 1 — Install

macOS or Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/aswathas/postiee/main/install.sh | sh
```

Windows, in PowerShell:

```powershell
irm https://raw.githubusercontent.com/aswathas/postiee/main/install.ps1 | iex
```

No Go toolchain needed. If you'd rather build from source:

```sh
git clone https://github.com/aswathas/postiee.git && cd postiee
go build -o postiee ./cmd/postiee
```

### Step 2 — Create a LinkedIn app

This is the part that takes about ten minutes, and it's mostly clicking.

1. Go to https://www.linkedin.com/developers/apps and create an app
2. On the **Products** tab, add **Share on LinkedIn** → this grants `w_member_social` (posting)
3. On the **Products** tab, add **Sign In with LinkedIn** → this grants `openid` and `profile`
4. On the **Auth** tab, add this redirect URL: `http://localhost:8085/callback`

You create your own app because LinkedIn grants permissions per application. There's no way around it.

### Step 3 — Write your credentials

```toml
# ~/.config/postiee/credentials.toml
client_id = "your_client_id"
client_secret = "your_client_secret"
```

```sh
mkdir -p ~/.config/postiee && chmod 700 ~/.config/postiee
chmod 600 ~/.config/postiee/credentials.toml
```

Those permissions matter. The file holds your secret.

### Step 4 — Log in

```sh
postiee auth login --port 8085
```

Your browser opens, you approve, done.

Access tokens last **60 days**. LinkedIn reserves automatic refresh for partners, so after that you run `auth login` again. If you're still signed in, it takes seconds.

### Step 5 — Check it works

```sh
postiee doctor
```

This reports where it found your config, whether a token is loaded, and whether LinkedIn's API is reachable. If something's broken, run this first.

---

## Using it with an AI agent

This is the part I'm happiest about, and honestly the reason I'd recommend postiee to a developer.

I'm already working inside an AI coding agent. Claude Code, Codex, Cursor, whatever. Those tools can read files and run commands.

If posting to LinkedIn is one command, then an agent can post.

postiee ships a skill that documents the command surface, the search behaviour, and the things an agent should never try:

```sh
# Claude Code
npx -y skills@latest add aswathas/postiee/skills --skill postiee -g -a claude-code -y

# Codex
npx -y skills@latest add aswathas/postiee/skills --skill postiee -g -a codex -y
```

Or use it as an MCP server, which works with any MCP-capable client.

The skill matters more than it sounds. It tells the agent things like: search is whole-word only, `forget` doesn't delete the LinkedIn post, posting spends your daily budget. An agent without that context will confidently do the wrong thing. An agent with it won't.

### Agent mode

Every command takes `--agent`, which turns on JSON output, compact formatting, no colour, and no interactive prompts:

```sh
postiee --agent search harness
```

That's it for anything parsing the output.

---

## What you get

| Command | What it does |
|---|---|
| `postiee post "text"` | Post to your feed |
| `postiee post --file post.md` | Post from a file — no shell quoting problems |
| `postiee post "text" --image p.png --alt "alt text"` | Post with one image |
| `postiee post "text" -v CONNECTIONS` | Visibility: `PUBLIC`, `CONNECTIONS`, `LOGGED_IN` |
| `postiee history` | Everything you've posted, newest first |
| `postiee search <query>` | Full-text search your own posts |
| `postiee forget <urn>` | Drop a post from your local record |
| `postiee auth` | Login, logout, status |
| `postiee doctor` | Diagnose config, credentials, connectivity |

Add `--agent` to any of them for JSON.

### Search that actually searches

`search` isn't a grep over a text file. It's a real index:

- **Porter stemming** — `authentic` finds `authentication`
- **Multiple terms AND together** — more words narrow the result, they don't widen it
- **Junk input never crashes it** — quotes, `AND`, parentheses, `NEAR(a b)`, `--` all return empty instead of erroring

Measured on my machine:

| Operation | Latency |
|---|---|
| Search 1,000 posts | 145 µs |
| Search 10,000 posts | 347 µs |
| List history | 79 µs |
| Record a post | 1.48 ms |

Ten times the posts costs 2.4× the latency, not 10×, because FTS5 queries the index rather than scanning rows. You can reproduce these yourself:

```sh
go test -bench=. -benchmem ./bench/
```

### Behaviour is pinned by tests

Seven properties are locked in by tests, so a change that breaks one fails the build instead of quietly changing behaviour:

| Property | Status |
|---|---|
| Porter stemming works | pass |
| Multi-term AND narrows results | pass |
| An absent term returns nothing | pass |
| Junk input never errors | pass |
| Post ID read from the header, not the body | pass |
| Re-recording upserts instead of duplicating | pass |
| Deleting removes the search index entry too | pass |

---

## Why this is worth using

**Posting stops being a project.** One command. No browser, no character counter, no image upload UI.

**Your post history becomes yours.** Searchable, offline, permanent. Not a 72-hour download link.

**Agents can do it for you.** The workflow that used to be "open LinkedIn, paste, hope" becomes "tell the agent to post this."

**It's local and private.** Your archive lives in a SQLite file on your machine. Nothing syncs anywhere.

**You can see everything it can't do.** That matters more than it sounds — see below.

---

## What it doesn't do

I'd rather be clear about this now than have you find out in an afternoon.

| | Why |
|---|---|
| **Read posts you made in the LinkedIn app** | `r_member_social` is closed. No API can do this. Search only covers posts made through postiee. |
| **Delete a post on LinkedIn** | No endpoint exists at any tier. `forget` removes your *local record* and leaves the post live. |
| **Comments, reactions** | Need Community Management API approval |
| **Analytics, impressions, followers** | Need separate analytics scopes, also gated |
| **Scheduled posts** | No `publishAt` field exists in the API. Scheduling would be a local timer pretending to be a feature. |
| **Multi-image or video posts** | One image works self-serve. Everything beyond needs partner approval. |
| **Messaging, DMs** | No public API exists at all |
| **Connection invitations** | No public API exists at all |
| **Another member's profile** | Person IDs are app-scoped, so you can't resolve anyone else |

### Why so many holes

LinkedIn's API wasn't built for individuals posting to their own feed. It was built for marketing platforms managing thousands of company pages.

The features that would genuinely help you — reading your posts, seeing how they performed, reacting to them — are exactly the features that require becoming a vetted partner.

So postiee ships only what works. A `comment` command that throws a 403 the first time you try it teaches you nothing and wastes an afternoon. An absent command tells you the truth immediately.

If LinkedIn ever opens one of these, the architecture already supports it. The local history stores the post URN, which is exactly what a comments table would need as a join key.

### Practical limits

- **150 API requests per day, per member.** One post uses one. Nothing is shared with anyone else. Searching and browsing history are local and cost nothing.
- **Tokens last 60 days.** Re-run `auth login` after that.
- **Search is whole-word only.** `search harn` will not find "harness", because SQLite FTS5 indexes complete tokens. Documented rather than worked around, because a search that quietly returns nothing is worse than one that tells you where it stops.

---

## Try it

- **Docs and install:** https://aswathas.github.io/postiee/
- **Source:** https://github.com/aswathas/postiee

Apache-2.0. macOS, Linux, Windows.

If you also have a folder of things you meant to post — give it one command instead.