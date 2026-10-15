# Attempt 002 — A release workflow that has never produced a release

**Opened.** 2026-10-14, commit `43d595d`.
**Status: open. Found 2026-10-16, one day after it was built.**

**Hypothesis.** Nobody should need a Go toolchain to install this. A release workflow
that builds all five platform targets, plus `install.sh` and `install.ps1` that
download a prebuilt binary, makes installation one line on every supported platform.

**How we tested it.** The workflow file was added. The platform detection was
"verified against real uname output". The installers were committed. Nothing
downloaded an artifact, because that was never done.

**What actually happened.** Running the documented Windows installer today:

```powershell
irm https://raw.githubusercontent.com/aswathas/postiee/main/install.ps1 | iex
```

```
Downloading postiee for Windows/x86_64 (latest)...
Invoke-WebRequest : Not Found
```

The workflow exists. There are no releases. `gh release list` on this repository
returns empty, and `git tag` returns empty.

So the one-line install — the headline promise, printed in the README under "No Go
toolchain required" — fails for every platform, and fails on the *first* thing it
does, before anything else could go wrong. The script got as far as choosing
`Windows/x86_64 (latest)` and then found nothing tagged "latest".

This was not found by reading the code. It was found by running the documented
command on the one machine that had neither Go nor a prebuilt binary.

**Outcome.** Open. Nothing fixed.

**Why it is still open.** Cutting a release is the owner's call, not an
engineering one — it attaches a tag and starts a public artifact. That decision has
not been made.

### What is definitely wrong and safe to fix

The README states "No Go toolchain required" and links installers that cannot work.
Either the workflow is triggered and a release cut, or the README says
`git clone` and `go build`. What must not stay is a documented path that has never
been executed.

### What would change this decision

Nothing about the design. The design is right — a prebuilt binary is the correct
answer for a CLI whose users are being asked to run one line in a PowerShell prompt.
The failure is that the last step of the pipeline was never pulled.

### What this attempt is a reminder of

**A workflow that has never run is not a tested path.** It is a file. The
verification that mattered was one `irm | iex` on a clean machine, and it would have
taken a minute when the installer was written.

The general failure: infrastructure gets reviewed for whether it *looks* right.
Nobody asks what happens on the first invocation. This repository caught its own
logic bugs with benchmarks and an eleven-point accessibility sweep, and shipped an
installer that had never been executed.

**Related.** The local build does work, and the binary that ends up installed is
useful. That is a separate matter and does not make the documented path work — see
`06-verification.md`.