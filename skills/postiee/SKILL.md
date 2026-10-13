---
name: postiee
description: Post to LinkedIn and search your own post history from the command line. Use when the user wants to publish a LinkedIn post, attach an image, check what they have posted before, or search their past posts. Triggers include "post to LinkedIn", "publish a LinkedIn post", "share this on LinkedIn", "what did I post about", "search my posts", "linkedin history".
license: Apache-2.0
---

# Postiee

Post to LinkedIn from the terminal, and search your own post history locally.

## The one thing to know first

**LinkedIn will never let any API read your posts back.** The `r_member_social`
permission is closed to all applicants — not slow, not gated on payment,
*closed*. So postiee captures the post's ID from the `x-restli-id` response
header at creation time and writes it to a local SQLite database. That record is
the only copy that will ever exist.

If a post is created but the ID is not captured, **the post is unrecoverable**.
postiee treats a missing header as a hard error for exactly this reason.

## Setup

The user needs a LinkedIn developer app. Check with `postiee auth status` first
— if it reports credentials, skip setup entirely.

1. <https://www.linkedin.com/developers/apps> → Create app
2. **Products** tab → add **Share on LinkedIn** (gives `w_member_social`)
3. **Products** tab → add **Sign In with LinkedIn** (gives `openid`, `profile`)
4. **Auth** tab → add redirect URL: `http://localhost:8085/callback`
5. Write credentials to `~/.config/postiee/credentials.toml`:
   ```toml
   client_id = "your_client_id"
   client_secret = "your_client_secret"
   ```
   Then `chmod 600 ~/.config/postiee/credentials.toml`
6. `postiee auth login --port 8085`

**Never ask the user to paste their client secret into a chat.** It would be
recorded permanently. Have them write it to the file instead.

### Scope troubleshooting

| Error | Cause | Fix |
|---|---|---|
| `unauthorized_scope_error` | Requesting a scope the app lacks | Add the product in the portal. During setup, `POSTIEE_SCOPES=w_member_social` narrows the request to what exists. |
| `403 … userinfo` | `openid` not yet granted | Add **Sign In with LinkedIn**, wait ~1h for propagation, then `auth logout && auth login` |
| `redirect_uri does not match` | Host/port mismatch | `POSTIEE_REDIRECT_URI` overrides the host spelling. It must match the portal **exactly**, including `localhost` vs `127.0.0.1`. |
| `403` on `me` / read-back | Expected | Reading posts needs partner approval. Not a bug. |

## Commands

```sh
postiee post "text"                      # post to your feed
postiee post --file post.md              # long post, no shell quoting
postiee post "text" --image p.jpg --alt "alt text"
postiee post "text" -v CONNECTIONS       # PUBLIC | CONNECTIONS | LOGGED_IN
postiee history                          # everything posted, newest first
postiee search <query>                   # full-text over your own posts
postiee forget <post-urn>                # drop from local history
postiee auth status                      # token state
postiee doctor                           # config, auth, API reachability
```

## Search behaviour — read this before promising results

SQLite FTS5 with the `porter unicode61` tokenizer. Concretely:

- **Stemming works.** `authentic` finds `authentication`.
- **Partial-word prefixes do NOT work.** `harn` does *not* find `harness`.
  FTS5 indexes whole tokens.
- **Terms are ANDed.** `search harness agent` requires both.
- **Junk input never errors.** Quotes, `AND`, `()`, `NEAR(a b)` are all handled.

If the user asks "what did I post about X" and gets nothing, suggest a longer
word — `authentication` rather than `auth`.

## What postiee deliberately cannot do

Do not offer these, and do not write code that calls them:

| Capability | Why not |
|---|---|
| `comment` | Needs Community Management approval |
| `react` / like | Needs Community Management approval |
| `analytics`, impressions | Needs Community Management approval |
| `dm` / messaging | No public API exists at all |
| `connect` / invitations | No public API exists at all |
| Read your own posts | `r_member_social` is **closed** |
| Schedule a post | No `publishAt` field exists; scheduling is client-side only |

Reporting these as "not yet implemented" is wrong — they are not implementable
at the self-serve tier.

## Rate limit

**150 requests/day per member.** One post costs one request. An agent that loops
will hit this, and the failure looks like an unrelated 403. Check
`postiee history --limit 200` before bulk work.

## Agent mode

Every command accepts `--agent`, which sets `--json --compact --no-input
--no-color` together. Prefer it when parsing output.

```sh
postiee --agent search harness
postiee --agent history --limit 10
```

## Safety

Posting publishes to a public feed. Confirm the exact text with the user before
running `post`, and never post credentials, tokens, or private keys as content.

Full reference: <https://github.com/aswathas/postiee>