# 07 — Retro

**Date.** 2026-10-16, two days after the last code commit. Written with the tool
shipped and used, not before.

## The shape of the project

Fourteen commits on one day, `5ecbb9b` through `a43b42e`, all dated 2026-10-14. The
tool was built, measured, documented, redesigned, published, corrected and
re-published in a single working session. The documentation followed over the next
two days.

That shape explains most of what is good and most of what is wrong here. What is good
is that the design work got real scrutiny — six review passes in parallel, eleven
defects found, a contrast ratio recomputed instead of re-estimated. What is wrong is
that almost nothing about the *shipped artifact* was checked by anything other than a
person remembering to look.

## What worked, and should be kept

**Measuring before claiming.** `bench/main_test.go` exists, and it found two real
bugs while being written — the `FTSQuery` prefix-match defect where the trailing `*`
sat inside the closing quote, and live account identifiers in the fixtures. Both were
caught before the commit landed rather than after. Writing the benchmark earned its
keep three times over: once for the numbers, once for the bug, once for the leak it
caused. `attempts/003` records the leak honestly rather than quietly fixing it.

**Letting the API reference overrule the reasoning.** The single most valuable moment
in the project is `0b06907`, where a documented impossibility turned out to be an
inference from a scope list with one bad premise in it. The commit says it exactly:

> "I had conflated the two."

The client had `DeletePost` implemented correctly the whole time. It was never wired
to a command, because the belief said it could not be wired. That is the failure mode
of documentation written from reasoning, and it was corrected by reading the reference
and then verifying live — post a test post, delete it with the self-serve token, get
204.

**Making failures non-destructive.** Two separate fixes have the same shape, which
suggests it is a habit rather than a coincidence. `cea3964` extended the OAuth
callback window from two minutes to fifteen because the timeout *discarded the flow
entirely rather than letting it be retried*. `0b06907` made `unpublish` delete
remotely before dropping the local row so a refusal leaves the record intact. In both
cases the question asked was not "can this fail" but "when it fails, is the state
recoverable".

**Naming what cannot be done.** The skill's table of impossible commands, with a
reason per row, exists because a command that fails on first use teaches the user
nothing and burns a rate-limit slot against a 150-per-day cap. It also stops an agent
from inventing a workaround and reaching for browser automation.

**Naming the known limit instead of hiding it.** 204 on delete is idempotent, so a
204 alone does not prove the post was there. That is in both the README and the skill.

## What went wrong, honestly

**Nothing verified the shipped artifact.** This is the theme of the retro and it
appears three times in `06-verification.md`.

The installers for five platforms shipped in `43d595d` and were never executed. On
2026-10-16, one day later, `install.ps1` was run and 404'd, because no release had
ever been cut — a release workflow existed, and a workflow that builds is not a
workflow that verifies (`attempts/002`). The docs site rendered as unstyled HTML on
every page, because GitHub Pages ran Jekyll over an Astro deploy and Jekyll excludes
directories beginning with an underscore, so the `_astro/` stylesheet 404'd. The file
was present and correct on GitHub the entire time, reachable through
`raw.githubusercontent`, 404 through Pages; only a content check found it
(`a43b42e`). And `postiee --version` reports `0.0.0-dev`, which the release ledger
says is expected for a fresh print, but which also means nothing distinguishes a dev
build from a released one at runtime.

Three failures, one cause: the project built and shipped a lot of surface that no
automated check ever touched.

**Documentation disagreed with the code, in the direction that guarantees failure.**
The setup page told users to register `localhost` for the OAuth redirect. `auth.go:241`
sends `127.0.0.1`, citing RFC 8252. Following the documentation produced exactly the
`redirect_uri does not match` failure the page's own warning described. The same wrong
host was in the agent skill, so the two documents agreed with each other and both
disagreed with the code — which is why `d525396` had to fix the skill as well as
`19caecc` fixed the page.

**And one instance of this is still in the tree.** The README's setup step 4 still
says to add `http://localhost:8085/callback`. The skill, which was corrected, says
`127.0.0.1` and explains why, and offers `POSTIEE_REDIRECT_URI` as the escape hatch
for anyone who already registered `localhost`. The README was missed. This is the same
defect `19caecc` claims to have fixed, still present in the most-read file in the
repository, two days later. It should be fixed and it is recorded here rather than
quietly patched.

**The plan was never written.** `04-plan.md` is a reconstruction, not a lock. Nothing
was scoped in advance, so `docs/decisions/` does not exist and there is no record of a
plan changing — because there was no plan to change. Everything that felt like scope
management during the build was actually just the first commit message's exclusion
list growing as problems surfaced.

**Parallel agents ran without a tree check between them.** A worktree created inside
`.claude/` was committed as an embedded repository and had to be removed in
`2e47fe6`. Six review passes ran in parallel against disjoint file sets and found
eleven real defects; the disjointness was doing real work, and still one path went
untouched.

**A limitation was written down before it was checked.** See `attempts/001`. The
generalisable form, taken from that write-up: a capability blocked by a permission you
do not have looks identical to a capability blocked by a permission nobody has, and
the only way to tell them apart is whether the API reference mentions the endpoint.

## What the next iteration should start with

**Cut a release and then run the installer.** Not before, and not on the assumption
that the workflow covers it. `attempts/002` is the largest single defect in the
project's history and it costs one command to close. If the release ledger stamps
versions externally, that is the constraint to design around, not a reason to leave
the installer pointing at nothing.

**Add one CI job that downloads the artifact the installers fetch and runs
`--version`.** Five platforms, one assertion each. This is the check whose absence
produced all three shipped-artifact failures. It is cheap, and it converts a class of
bug from "found by a person on day two" to "found by a machine on day one".

**Add a content check on the deployed docs site.** Fetch the published URL, assert
the stylesheet link resolves and that the body carries real text rather than bare
markup. The `_astro/` 404 was invisible to every check that ran.

**Fix the README redirect host and add a doc/code consistency check.** The check can
be narrow: grep the docs and the skill for `localhost:8085` and fail if
`auth.go` sends anything else. It would have caught the page and the skill in one
run, and it would catch the README today.

**Exercise the MCP server end to end, or say plainly that it is untested.** It ships
in two transports and the skill recommends shell, which means the less-preferred path
is also the less-tested one. Either path should be the verified one.

**Measure the teach/recall loop.** `AGENTS.md` sets the bar itself: judge it on
recall hit rate at a minimum denominator of fifty events, and near-zero at that
denominator means the loop is not earning its keep. That measurement has not been
made. Until it is, the loop's value is asserted rather than known — and the fix is
one `learnings stats` call.

**Write the plan before the work, next time.** The exclusion list worked well as a
scope statement; what it could not do was tell anyone, mid-build, which commitment
was load-bearing. That was only discovered afterwards, by noticing that no phase
forced an architecture change. A locked plan would have made the "no release was ever
cut" gap visible at phase 2 rather than two days later.

## The one-line version

The design was measured and the delivery was not. Benchmarks caught bugs the tests
did not look for, and review passes caught defects nobody had asked about — but
nothing in the pipeline ever asked whether the thing a user installs actually works,
and that gap is the whole retro.

## What this record does not contain

Honesty about the gaps, since that is the point of the file. There is no user
feedback in it: this was built for one person, so nothing here is validated by
anyone else's hands. There are no latency or failure numbers from real multi-user
use — only the single-machine WSL2 benchmarks. Nothing is recorded about the
distribution of real posts through the tool, because that was never measured. And
nothing is recorded about what an agent actually did when it met the skill; the
skill's contents are reasoned, not observed, which is the same gap
`AGENTS.md` already names for the learn/recall loop.