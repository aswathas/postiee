# Attempt 003 — Benchmarks seeded from the live account

**Opened.** 2026-10-14, while writing the benchmark suite in `43d595d`.
**Closed.** 2026-10-14, commit `43d595d`. Fixed before the commit landed.

**Hypothesis.** A benchmark needs realistic data. The most realistic data available
is the output of the tool running against a real account — real member ids, real
post urns, real permalinks.

**How we tested it.** `bench/main_test.go` generates every number quoted in the
README. To generate them faithfully, the fixtures were populated from what the live
account had actually produced.

**What actually happened.** Real member ids and real post urns were committed to the
repository as test fixtures.

The commit message records the fix in one line:

> "Member ids and post urns from the live account had leaked into benchmark
> fixtures. Replaced with placeholders before this commit."

**Outcome.** Fixed before merge. Placeholders replaced the real values.

### The thing that is still true about this

The values are gone from `HEAD`. **They are not gone from history.** A repository
that was public before this commit may have had them indexed, and anyone with the
old commit can read them.

Whether that matters is a judgement about what a member id and a post urn are worth:

- A member id is app-scoped, not a LinkedIn-wide identity. `03-decision.md` records
  that person ids are scoped per app, so the same human has a different id under a
  different app. On its own it identifies an app, not a person.
- A post urn identifies a post. If that post is still public, it is already public by
  definition.

So the exposure is small. It is recorded anyway, because "it turned out to be low
risk" is a conclusion reached after the fact, and the next person seeding fixtures
should know it happened once and why it was treated as a defect at the time.

### What this attempt is a reminder of

**Realistic test data is the most likely place for real data to leak.** Test
fixtures get less scrutiny than source because they do not run in production. A
credential, a member id, a customer name, an email address — all of them are most
likely to survive in a fixture file.

The general rule: fixtures are data, and data has provenance. If it came from a
live system, it does not belong in version control until it has been through a
substitution pass.

### The positive half

The same commit that leaked the ids also found a real bug while measuring:

> "FTSQuery emitted quoted terms whose trailing `*` did not prefix-match, so
> partial-word search silently returned nothing. The star must sit outside the
> closing quote."

Writing the benchmark was worth it twice over: once for the number, once for the
bug, and once for the leak it caused.