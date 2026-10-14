# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Two confirmed audiences, and the site must serve both without splitting into two products.

- **The undecided developer.** Arrives from a LinkedIn post, a link shared in chat, or a search. Does not know postiee exists. Needs the closed-permission argument before they will care. Writes the home page.
- **The committed developer.** Landed from GitHub or the README. Has decided. Wants install steps, commands, and the API gotchas that cost them an afternoon otherwise. Reads the docs.

The author is the current primary user of both surfaces.

## Product Purpose

postiee posts to LinkedIn from a terminal and keeps a searchable local record of every post made through it.

Success means a user can post once, and find that post again months later without LinkedIn's help. Secondary success: an AI agent can post on the user's behalf without the user opening LinkedIn.

## Positioning

LinkedIn's `r_member_social` permission — the one that exists so an app can read your own posts — is closed to all applicants. No API can read back what it wrote.

Therefore a post's ID, captured at creation time from the `x-restli-id` response header, is the only handle that will ever exist on that post. The local SQLite history is not a convenience feature; it is the only surviving copy.

A competitor could add comments or analytics. It could not truthfully copy the premise, because the premise is a permission LinkedIn has closed.

## Operating Context

- Install is one line per platform; the user creates their own LinkedIn developer app because scopes are granted per application.
- Setup is roughly ten minutes the first time and is dominated by LinkedIn app creation, not by installing.
- Daily use is one to a few commands.
- Post history is a personal archive that accumulates over years and is expected to stay searchable.
- Common workflow: draft an idea, hand it to an AI agent, agent posts text plus one image.

## Capabilities and Constraints

Works self-serve:

- Create a post: text, one image with alt text, or a body read from a file
- Visibility: `PUBLIC`, `CONNECTIONS`, `LOGGED_IN`
- Local history with full-text search (SQLite FTS5, porter stemming)
- OAuth login, logout, status
- Agent surfaces: `--agent` JSON mode, MCP server, installable skill

Hard constraints:

- **150 API requests per day per member.** One post is one request.
- Access tokens live 60 days; programmatic refresh is partner-only, so re-login is required on expiry.
- Person IDs are app-scoped and must be resolved at runtime, never hardcoded.
- Image upload is exactly two calls; there is no finalize step.
- URNs in paths need full percent-encoding; Go's `url.PathEscape` does not do this.
- Search matches whole tokens only. Partial-word prefixes do not work and are documented rather than worked around.

Confirmed unavailable, and must never be presented as forthcoming without a LinkedIn policy change:

- Comments, reactions, analytics — partner approval
- Messaging, DMs, connection invitations — no public API exists at any tier
- Running ads — separate Marketing API and approval
- Deleting a post on LinkedIn — no API exists
- Reading posts back from LinkedIn, or resolving another member's ID

## Brand Commitments

Name: **postiee**. Author is credited as Aswath AS. License is Apache-2.0.

Voice commitment from the author: plain-spoken and short. The author explicitly rejected marketing gloss and wants the site's copy written the way they would say it out loud.

The author considers honesty about limitations a feature, not a liability. The scope table is load-bearing content, not a disclaimer.

## Evidence on Hand

- Benchmarks measured on the author's machine via `bench/main_test.go`: search 1,000 posts 144 µs / 32 allocs; search 10,000 posts 343 µs / 32 allocs; list history 81 µs / 581 allocs; record a post 1.45 ms / 33 allocs. Reproducible with `go test -bench=. -benchmem ./bench/`.
- Test suite covering stemming, multi-term AND, junk input, upsert, index deletion, header parsing, two-call image upload, and URN encoding.
- Live-verified end-to-end against the real LinkedIn API, including a real post.
- Three SVG diagrams in `docs-site/src/components/`: PostLifecycle, ScopeGate, CliSurface.

Absences future work must not fabricate: no testimonials, no user counts, no performance claims for the LinkedIn API round trip (never measured end-to-end), no marketing logos, no pricing.

## Product Principles

1. **Ship only what genuinely works.** A command that 403s on first use is worse than an absent command, because it teaches the user nothing and wastes their time. The omission list is longer than the feature list and every omission states its reason.
2. **Never scrape.** Official API only. Credentials go to `api.linkedin.com` and nowhere else.
3. **The local record is the product.** Anything that risks losing a post ID is a defect, because there is no recovery path.
4. **Measure, do not assert.** Every number published comes from a test the user can rerun.
5. **State limits plainly.** Naming what the tool cannot do is a trust mechanism, not an apology.

## Accessibility & Inclusion

The author's stated requirement is that the site meets WCAG AA contrast at minimum, keeps a visible focus ring, has a skip link, respects `prefers-reduced-motion`, and never drops body text below 16px on mobile.

Honoring `prefers-reduced-motion` is treated as non-negotiable rather than polish: motion is decorative here, and decorative motion excludes people who get nausea from it.