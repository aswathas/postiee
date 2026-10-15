# 03 — Decision

**Date.** 2026-10-14

## A missing post id is a hard error, not a warning

**Chosen.** `POST /rest/posts` returns `201` with an empty body. The id is in the
`x-restli-id` header. If that header is absent, the command fails.

### What ruled out the alternative

Logging a warning and continuing.

There is no recovery. The post is live on LinkedIn. Without the id it cannot be
recorded, cannot be deleted, and cannot be read back — by postiee or by anyone else,
because no API can read posts back at all. The local database is the only copy of
that id that will ever exist.

A warning would have produced a tool that occasionally loses posts silently, and the
user would find out weeks later from the other direction: a post they cannot delete.

### What would change this decision

Nothing at this tier. If LinkedIn ever returned the post body on creation, the
header would become a belt-and-braces check rather than the only source.

## Deletion exists, and it is remote-first

**Chosen.** `unpublish` calls `DELETE /rest/posts/{encoded-urn}`, which returns 204
and uses `w_member_social` — the same self-serve scope posting already uses. It
deletes on LinkedIn first and only then drops the local row.

### What ruled out the alternative

The alternative was documenting deletion as impossible, which is what shipped first
and was wrong. See `attempts/001`.

The ordering within the command is the second decision, and it is not arbitrary.
Remote-first means a refusal — an expired token, a revoked scope — leaves the local
record intact, so the post is still findable and still deletable once auth is fixed.
Local-first would leave a post live on LinkedIn with no local record, which is the
exact unrecoverable state this tool exists to prevent.

### The known limit, recorded honestly

204 is idempotent: deleting an already-deleted post also returns 204. So a 204 alone
does not prove the post was there. The command cannot distinguish "deleted it" from
"it was already gone".

### What would change this decision

Nothing. There is no undo and LinkedIn offers no restore, which is why the agent
skill forbids running `unpublish` without an explicit request naming the post.

## Credentials never touch a command line

**Chosen.** `~/.config/postiee/credentials.toml`, mode 600, inside a directory at
mode 700. `unpublish` reads the token from the store rather than accepting
`--authorization`.

### What ruled out the alternative

The generated `posts delete` command takes `--authorization` on the command line.
That is the documented, generated, working path — and it puts a live token into shell
history and into process listings, where it persists.

### What would change this decision

Nothing on this platform. If a future environment guaranteed no shell history and no
shared process table, the flag would be acceptable. Neither is guaranteed.

## Commands an agent must never attempt are named explicitly

**Chosen.** The skill carries a table of things that are not unimplemented but
impossible at any access tier, with the reason for each.

### What ruled out the alternative

Leaving them undocumented and letting an agent discover them by failing.

The skill's reason is about teaching:

> "These are not unimplemented. They are impossible at any access tier, and a command
> that fails on first use teaches the user nothing."

A wrong answer that costs a failed API call teaches the agent nothing and costs the
user a rate-limit slot against a 150-per-day cap. Naming the impossibility is cheaper
and is the only version that is true.

### What would change this decision

If LinkedIn opened any of these scopes at the self-serve tier, the row moves to a
command and the table shrinks. Messaging and connection invitations are listed as
having no API at all, not as gated ones.

## The core is generated; local edits stay narrow

**Chosen.** The CLI is a `postiee-pp-cli` produced by
[CLI Printing Press](https://github.com/mvanhorn/cli-printing-press). Systemic fixes
go upstream. Any local modification to the generated tree is recorded under
`.printing-press-patches/`, and regen and publish-validate **fail closed** when a
recorded file or call site has disappeared.

### What ruled out the alternative

Hand-maintaining the whole tree.

The generated tree covers the command surface, flags, output envelopes, the MCP
server, and a self-capturing teach/recall loop over a local SQLite store. Hand-
maintaining that means re-implementing it on every reprint.

The fail-closed rule is what makes this safe. A customization that is dropped by a
reprint cannot ship as though it were still applied — the validation refuses.

### What would change this decision

If the hand-written parts — the LinkedIn client, the history layer, credential
handling — grew to dominate the tree, the balance would shift toward owning it.