# 05 — MVP scope

**Date.** 2026-10-16. Frozen after the tool shipped; reconstructed from the shipped
command surface, the first commit's own exclusion list, the README's "What it cannot
do" table, and the skill's limitations table.

## What the MVP is

postiee is a terminal client for the one thing an agent cannot otherwise do safely:
publish to LinkedIn through the official API, and keep a recoverable local record of
what it published.

The shipped command surface is ten entries, and they split cleanly.

**Posting.** `post` takes text, a file, or one image with alt text, at `PUBLIC`,
`CONNECTIONS`, or `LOGGED_IN`. One API call for the post itself, one `initializeUpload`
plus a signed-URL `PUT` for an image, and a `/v2/userinfo` round trip to resolve the
member URN at runtime rather than hardcoding it.

**Remembering.** `history` lists recorded posts newest first. `search` is full-text
over post bodies via SQLite FTS5 with Porter stemming. Both read the local database
and never touch the network.

**Getting out of it.** `forget` drops a row from local history only. `unpublish`
deletes on LinkedIn and then drops the row. Two commands, deliberately not
interchangeable, because the user asking "remove this from my history" and the user
asking "make this post go away" are different requests.

**Auth.** `auth login | logout | status`, credentials in
`~/.config/postiee/credentials.toml` at mode 600 inside a 700 directory, and
`doctor` for config, credentials and API reachability.

**Agent surface.** `--agent` on any command for JSON, compact, non-interactive,
colour-free output; the skill at `skills/postiee/`; and `postiee-pp-mcp` over stdio
or streamable HTTP as an alternative transport, with the skill recommending shell.

**Documentation.** README, a five-page Astro site on GitHub Pages, and
`SKILL.md`. `PRODUCT.md` and `DESIGN.md` record product truth and the design tokens
so they are not re-derived blind.

The architectural commitment underneath all of it is one sentence from the first
commit: the post urn is captured from `x-restli-id` and written to SQLite, because
`r_member_social` is closed to all applicants and no API can read your posts back.
Everything else — search, list, delete — is downstream of that one decision.

## What the MVP deliberately does not do

Each exclusion carries the reason. An exclusion without a reason is an oversight, and
this list exists so the next person reads a decision instead of a gap.

### No comment, react, or like

Community Management API, partner approval. Not gated by effort and not achievable at
the self-serve tier. The reasoning behind naming this at all is in `03-decision.md`: a
command that fails on first use teaches the user nothing and costs a rate-limit slot
against a 150-per-day cap, so the impossibility is stated in one sentence instead.

### No analytics, impressions, or follower counts

Separate partner-gated scopes, same category as above. The benchmark numbers in the
README are postiee's own `go test -bench` output about its local store, not LinkedIn
analytics, and the two are never conflated in the docs.

### No reading posts back from LinkedIn

`r_member_social` is closed to all applicants. This is the fact the whole tool is
built around rather than a missing feature, and it is permanent at every tier: not
slow, not paid, closed. The consequence is stated in the skill in the bluntest
available terms — if a search returns nothing, the user may be searching for something
posted before postiee existed, no tool can recover that, and postiee should say so
plainly rather than implying it is broken.

### No direct messages and no connection invitations

No API exists at any tier, not a gated one. The skill also notes that mass messaging
violates the ToS, so this is not a feature awaiting approval.

### No scheduled posts

LinkedIn's post schema has no `publishAt` field. Scheduling would mean a local timer
that fires a process, which is a different tool with a different reliability story —
and it would mean the MVP owns something that can fail silently while nobody is
watching.

### No edit of an already-published post

There is no update endpoint in the shipped surface. A post is captured once, at
creation, and the id captured with it is the only handle that will ever exist. Edit
was not built because the write path is the dangerous one here: a second remote call
against a resource that cannot be read back is exactly the class of operation this
tool is trying to make safe.

### No multi-image or video posts

One image per post on the free tier; multi-image and video need partner approval.
Shipping a flag that fails for every user who sets it would be worse than not having
the flag.

### No hard-coded person id, and no cross-app reuse

Person ids are scoped per app — the same human has a different id under a different
LinkedIn app, and reusing one 404s. The member URN is therefore resolved at runtime
through `/v2/userinfo` every time, and there is no config option to pin it, because a
pinned id that goes stale produces a 404 at the moment of posting.

### No automatic token refresh

Tokens expire at sixty days and programmatic refresh is partner-gated. The supported
path is a manual `auth login`, and the skill forbids an agent from running one because
it opens a browser the user has to approve. This is an honest ceiling rather than a
gap: nothing in the MVP could fix it.

### No credential on the command line

`unpublish` reads the token from the credential store rather than accepting
`--authorization`, even though the generated `posts delete` that does exist takes the
flag. A live token in a flag lands in shell history and in process listings. The
whole pitch of the tool is that credentials go only to `api.linkedin.com`, and
accepting a token in argv would break that in the one place a user is most likely to
paste one.

### No third-party history storage

Local SQLite, beside the credentials. A hosted database would mean storing somebody's
posts on somebody else's machine for a tool whose argument is credential locality.

### No hand-maintained fork of the generated CLI

The command surface, flags, output envelopes, MCP server and the teach/recall loop
come from a CLI Printing Press print. Systemic fixes go upstream; local edits are
recorded under `.printing-press-patches/` and regen and publish-validate fail closed
when a recorded file or call site has disappeared. Hand-maintaining the tree would
mean re-implementing all of it on every reprint.

### No agent-run authentication

The skill states it as a rule: `auth login` opens a browser and the user must
approve, so the agent asks and stops. This is a scope decision about who the commands
are for, not a missing flag.

### No CI job that downloads the installed artifact

This is the sharpest exclusion in the list because it is an omission rather than a
refusal, and it has a known cost. Installers for five platforms shipped in `43d595d`
and were never executed. `install.ps1` 404s today, because no release was ever cut
(`attempts/002`), and the docs site shipped unstyled for a commit because nothing
checked the deployed result (`a43b42e`). A workflow that builds is not a workflow that
verifies, and the MVP treated the first as the second. That is the thing the next
iteration starts with; `07-retro.md` picks it up.

## The exclusion that was wrong

One exclusion in this file is listed for honesty rather than as a standing decision:
for a while the tool documented post deletion as impossible. It was not blocked. See
`attempts/001`. The command now exists, and the reason it had been written off is
worth keeping in view — the false claim came from reasoning over the scope list rather
than from reading the API reference, and the reasoning had one bad premise in it.