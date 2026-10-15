# 04 — Plan

**Date.** 2026-10-14.

## A note on how this file came to exist

Normally a plan is written before the work and locked, so that any later change has
to be argued for in `decisions/`. This plan is written after the fact. It is a
reconstruction of the eight phases the fourteen commits in `5ecbb9b` through
`a43b42e` actually followed, and it is labelled as executed rather than as intended.

The honest reason is that the plan was never written down. What existed instead was
a scope statement in the first commit message and a growing list of things that had
to be true before the tool could be handed to an agent. `docs/decisions/` does not
exist in this repository, which means no locked phase was ever formally changed —
there was nothing to change, because nothing was ever locked. That is a process
finding, and it is recorded in `07-retro.md`.

Every phase below is grounded in commit messages. Where the ordering in the log was
incidental rather than planned, this file says so.

## Phase 1 — the CLI itself

**Delivered by `5ecbb9b`.** A Go CLI that posts to LinkedIn through the official
Community Management API and keeps a local, searchable record of every post.

Six commands in the first commit: `post`, `history`, `search`, `forget`, `auth`,
`doctor`. The architecture that everything else depends on is in this commit: the
post urn is captured from the `x-restli-id` response header at creation and written
to SQLite with FTS5 over the body, because `r_member_social` is closed and no API
can read posts back.

The commit also states the first exclusion list in the project's own words —
"comment, react, analytics, dm, connect" — and the reason for each category. That
list became the table the agent skill now carries.

## Phase 2 — distribution, measurement, and the README

**Delivered by `43d595d`.** This is where the tool stopped working only for its
author and started working for anyone else.

Four things, in one commit because they were one phase: `install.sh` and
`install.ps1` with a release workflow building five platform targets; the agent
skill at `skills/postiee/`; the benchmark suite in `bench/main_test.go`; and the
README.

The benchmarks were not decoration. They produced every number the README quotes,
and writing them found two real bugs before the commit landed — a prefix-match
defect where the `*` sat inside the closing quote in `FTSQuery`, and live account ids
in the fixtures. See `attempts/003`.

## Phase 3 — the docs site

**Delivered by `f4494cd`, `48fdaa1`, `d767d35`.** Five Astro pages, static, on
GitHub Pages.

This phase cost more than the plan expected. `f4494cd` is a build fix as much as a
delivery commit: a literal JSON example in a `<pre>` block started with `{`, which
Astro parsed as the start of a JSX expression. `48fdaa1` moved the site off
Cloudflare Pages onto GitHub Pages, which meant real work — `BASE_URL` carries no
trailing slash, so naive concatenation produced `/postieeabout`. `d767d35` links it
from the README.

The decision to build a docs site at all, against a README-only alternative, is in
`02-options.md`.

## Phase 4 — making login survivable

**Delivered by `cea3964`.** The OAuth callback window went from two minutes to
fifteen.

This is the smallest commit in the history and the clearest defect it fixed. Two
minutes is not enough for a password, a 2FA prompt and an email code, and the
timeout discarded the flow rather than letting it be retried. The failure was
destructive, not recoverable — which is the same class of bug as dropping a local
row before a remote delete, and it was fixed the same way, by making the failure
non-destructive.

## Phase 5 — design, and the eleven defects review found

**Delivered by `10c75cd` and `19caecc`.** `PRODUCT.md` for product truth, `DESIGN.md`
for the token set and settled decisions, and a full site rebuild around eight
decisions.

`10c75cd` recomputed a contrast ratio rather than re-estimating it: `--fg-3` was
documented at 4.6:1 and computes to 4.14:1, so it failed AA. `19caecc` records six
`impeccable` passes run against disjoint file sets and eleven defects, including a
setup page that told users to register `localhost` while `auth.go:241` sends
`127.0.0.1`. Following that page produced precisely the failure the page's own
warning described. All eleven are enumerated in `06-verification.md`.

## Phase 6 — the agent-facing surface

**Delivered by `d525396` and `99f52af`.** The skill rewritten around what an agent
actually needs, and the README restructured to lead with agent installation.

The trigger is a behaviour, not a preference: an agent asked to post to LinkedIn
searches for a LinkedIn MCP server, finds nothing official, and either scrapes with
the user's session cookies or gives up. The skill now intercepts that search. It
covers finding the binary including the WSL case, passing `--agent` for parseable
output, the rule that `auth login` is user-only, and an explicit table of what the
tool cannot do so an agent says "not possible" instead of writing a workaround.

`d525396` also fixed the same wrong redirect host the docs carried — the skill and
the setup page had agreed with each other and both disagreed with the code.

## Phase 7 — deletion, which was never blocked

**Delivered by `0b06907`.** `unpublish`.

This phase did not exist in any plan. The tool had shipped documenting deletion as
impossible, and that documentation was wrong. `DELETE /rest/posts/{encoded-urn}`
returns 204 and needs `w_member_social`, the same self-serve scope posting already
uses; only *reading* needs the closed `r_member_social`. Verified live before
shipping. `unpublish` reads the token from the credential store rather than taking
`--authorization`, and deletes remotely before dropping the local row.

`attempts/001` records the false claim; `03-decision.md` records the decision and
what would reopen it.

## Phase 8 — shipping the site correctly

**Delivered by `a43b42e`.** A one-line `.nojekyll` file, added last.

GitHub Pages ran Jekyll over the Astro deploy, and Jekyll excludes any directory
beginning with an underscore. Astro emits its bundled CSS to `_astro/`, so the
stylesheet 404'd and every page rendered as unstyled HTML. The asset was present and
correct on GitHub the whole time — reachable through `raw.githubusercontent`, 404
through Pages. Only a content check found it.

This phase is in the plan because its lesson is a plan lesson: nothing verified that
the deployed artifact was actually usable, on any platform, for any phase of the
project.

## Two phases that were not phases

`49fa36d` removes an accidentally committed message file. `2e47fe6` removes a nested
git repository a parallel agent created inside `.claude/` and adds it to
`.gitignore`. Neither delivers scope. Both are recorded here because both are
recorded in `06-verification.md` as failures, and the second one is evidence that
work in this project was running in parallel without the tree being checked between
passes.

## What the plan got right, retrospectively

The order was correct on the one axis that mattered: build the tool, then make it
installable, then measure it, then document it, then make it safe for an agent to
drive. Nothing in phases 3 through 7 forced a change to phase 1's architecture. The
local-history design held from `5ecbb9b` to `a43b42e` without a rewrite.

## What the plan did not cover

Nothing checked the shipped artifact on the platform it shipped for. Phase 2 built
installers for five platforms; no phase ever ran one. Phase 8 exists because a
static file was missing from a deploy. `attempts/002` is the same gap with a larger
consequence: the documented installer 404s, because no release was ever cut.