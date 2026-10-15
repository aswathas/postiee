# 01 — The problem

**Date.** 2026-10-14

## The problem that started it

There is no official LinkedIn MCP server. There are unofficial ones, and the shape
of them is the problem: most scrape with your session cookies, which means handing
your account — and your session — to a third party. Some route credentials through
paid intermediaries.

So an agent that needs to post to LinkedIn has three bad options and one that
requires building something.

The first commit states it:

> "Postiee: LinkedIn posting CLI with local history"

## The part nobody expected

The interesting problem is not posting. Posting is one HTTP call with one scope.

The problem is **reading your posts back**, and it is structural rather than a
matter of effort. LinkedIn's `r_member_social` permission — the one that reads your
own posts — is closed to all applicants. Not slow. Not paid. Closed.

That single fact inverts the normal shape of a posting tool. A posting tool that
cannot read its own posts back cannot:

- show you what you have posted
- search your history
- delete anything it created, because deletion needs the post id
- recover from a mistake

The first commit's own summary names the answer: "local history". Because the API
will not give the id back, the id has to be captured at the moment of creation and
kept somewhere the API cannot reach.

## What that costs if it is built wrong

The cost is not a failed feature. It is a **lost post**.

`POST /rest/posts` returns `201` with an empty body. The id arrives in the
`x-restli-id` response header. A client that reads `.Body` gets nothing, treats the
post as successful, and has no id — so the post cannot be found, cannot be deleted,
and cannot be searched. It is live on LinkedIn and unreachable forever.

The README states the stakes directly:

> "If this ID is missed, the post is gone forever."

## A second problem, discovered later

Partway through, the tool was documented as unable to delete posts. That was wrong,
and the commit that corrected it says why it was wrong in a way that matters:

> "I had conflated the two."

Reading requires the closed permission. Deleting does not — it uses
`w_member_social`, the same self-serve scope that posting already uses. The
documentation had turned a temporary absence of understanding into a permanent
statement of impossibility. See `attempts/001`.