# Attempt 001 — Documenting post deletion as impossible

**Opened.** 2026-10-14, during the first pass of the README and agent skill.
**Closed.** 2026-10-14, commit `0b06907`. Abandoned — the claim was false.

**Hypothesis.** Deleting a post is not possible through LinkedIn's API at the
self-serve tier. It should be documented as a limitation, alongside comments,
reactions and analytics, so users do not waste time looking for it.

**How we tested it.** Reasoning from the scope list. `r_member_social` is closed, so
reading posts is impossible. Deletion needs a post id, and if the id cannot be read
back, deletion is impossible too.

**What actually happened.** The reasoning had one bad premise in it, and the commit
that fixed it names it precisely:

> "I had conflated the two."

Checking against LinkedIn's Posts API reference:

- `DELETE /rest/posts/{encoded-urn}` returns **204**
- It requires **`w_member_social`** — the same self-serve scope that posting already
  uses
- Only **reading** posts needs `r_member_social`, which is the closed one

Deletion never needed to read the post. It needs the id, and the id comes from the
`x-restli-id` header at creation time — which the tool already captures. The
dependency chain that made deletion look impossible did not exist.

It was then verified live: post a test post, delete it with the self-serve token,
get 204.

**Outcome.** Abandoned. The limitation was struck from the documentation and
`unpublish` shipped.

### Why it matters that it was wrong in this specific direction

The commit says it better than it can be restated:

> "That was wrong, and it was wrong in the direction that would have shipped a
> missing capability as a permanent one."

The client already had `DeletePost` — correctly implemented, with the
`X-RestLi-Method` header and the URN encoding. It was never wired to a command. So
the work was done; only the belief was wrong, and the belief was what stood between
a working feature and a documented impossibility.

This is the specific failure mode of documenting limitations from reasoning rather
than from checking. A limitation nobody challenges looks exactly like a fact.

### What this attempt is a reminder of

**"Impossible at any tier" is a claim that requires evidence, and this one had
none.** Every other row in that table is grounded in an API reference or a partner
gate. This row was grounded in an inference, and the inference was wrong.

Before writing that something cannot be done, check whether the thing you are
relying on is actually a dependency of it. Here the answer was no: the id was
already captured at creation, so the closed reading permission was never on the path.

### The general form

A capability blocked by a permission you do not have looks identical to a capability
blocked by a permission nobody has. The difference is whether the API reference
mentions the endpoint at all. Check the reference. A limitation found by reading a
scope list is a hypothesis wearing a fact's clothes.