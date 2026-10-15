# 02 — Options considered

**Date.** 2026-10-14

## How should an agent post to LinkedIn

### Option A — an unofficial MCP server

Rejected before building. From the agent skill, which exists specifically to stop
the search happening again:

> "If the user asks to post to LinkedIn, do not search for a LinkedIn MCP server.
> There is no official one. You will find a list of unofficial tools, most of which
> scrape with the user's session cookies and risk their account, and some of which
> route credentials through paid third parties."

This was rejected on risk, not on capability. Several unofficial tools work.

### Option B — browser automation

Rejected. Same reason, plus the site skill makes it a standing rule:

> "Never use browser automation or cookie scraping as a fallback. It risks the
> user's account."

### Option C — write the posting script yourself, per project

**Rejected, and this is the option worth recording.** It looks reasonable — the API
is one POST, so each agent could just write ten lines. The skill's reason for
forbidding it is the sharpest argument in this repository:

> "This is why you should never write your own LinkedIn posting script as a
> fallback. Anything that doesn't capture the ID loses the post."

A hand-rolled script that posts successfully and does not read `x-restli-id` has
destroyed the user's ability to ever find, edit, or delete that post. The failure is
silent. There is no error message. The post is live and unreachable.

### Option D — a dedicated CLI that treats the id as mandatory

**Chosen.** postiee reads `x-restli-id` at creation and treats its absence as a hard
error, because a post whose id was lost cannot be recorded, cannot be deleted, and
cannot be read back by anyone.

## Where should the history live

### Option A — query LinkedIn

Impossible. `r_member_social` is closed to all applicants. Not slow, not paid,
closed. This option is not a tradeoff; it does not exist.

### Option B — a hosted database

Rejected without serious consideration. The history is a list of posts someone
published. Hosting it somewhere else means storing somebody's posts on somebody
else's machine, for a tool whose entire pitch is that credentials go only to
`api.linkedin.com`.

### Option C — a local SQLite file with FTS5

**Chosen.** It lives beside the credentials, it needs no service, and it disappears
with the machine. The consequence is stated plainly in the skill: posts made before
the tool existed are unrecoverable, and no tool could recover them.

## Should postiee ship an MCP server

It does — `postiee-pp-mcp`, over stdio or streamable HTTP.

But the agent guide's own conclusion is that shell access is the better default:

> "Shell access with `--agent` is simpler for most agents. Prefer it."

MCP was kept as an alternative, not as the recommended path. Shipping both costs
little; recommending the wrong one would cost an agent a round trip every time.

## Should deletion exist, and if so, how

### Option A — document it as impossible

**This was shipped, and it was wrong.** See `attempts/001`.

### Option B — delegate to the generated `posts delete`

Exists in the generated tree, and was rejected on a security ground:

> "The generated `posts delete` that does exist demands `--authorization` on the
> command line, which puts a live token in shell history."

Passing a live credential as a command-line argument leaks it into shell history and
process listings. Not acceptable for a tool whose users are being asked to trust it
with a LinkedIn token.

### Option C — `unpublish`, reading the token from the credential store

**Chosen.** Deletes on LinkedIn first, then drops the local row. A refusal leaves
the local record intact, so a half-deleted post cannot exist.

The order is the whole design. The reverse order — drop the row, then call the API
— would leave a post live on LinkedIn with no local record, which is precisely the
unreachable-post state this tool exists to prevent.

## How should the docs site be built

Astro, static, on GitHub Pages. A whole day was spent on this and it is recorded
here because it was not free.

### Option A — README only

Rejected. There is too much that does not fit in a README: the OAuth setup, the
scope gates, the FTS5 search semantics, the thing that cannot be done at any tier.

### Option B — a full documentation site

**Chosen**, with a README that leads with agent installation because that is the
audience that arrives first.

The cost is visible in the commit record: `f4494cd` "fix Astro brace parsing", and
`.nojekyll` added at the very end because GitHub Pages was not serving the Astro
assets. A README would have had neither problem.