# Postiee CLI

Self-serve LinkedIn posting surface for a personal developer app.

Scope of this spec is deliberately limited to what the "Share on LinkedIn"
product grants a self-serve app (scope `w_member_social`): create a post as
the authenticated member, and delete one. Nothing else.

NOT INCLUDED, because the API does not grant it at any self-serve tier:
comments, reactions, analytics, reading your own posts (`r_member_social`
is a closed permission), messaging, connection invitations, third-party
profile reads. Commands for those must never appear in `--help`; a command
that 403s on first use is worse than an absent one.

Two create endpoints are modelled because LinkedIn's own docs disagree:
`/v2/ugcPosts` is the endpoint the Share-on-LinkedIn product documents
against, but the Posts API doc states it "replaces the ugcPosts API". The
client tries `/v2/ugcPosts` first and falls back to `/rest/posts`.

## Install

The recommended path installs both the `postiee-pp-cli` binary and the `pp-postiee` agent skill (Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, and other agents supported by the upstream [`skills`](https://github.com/vercel-labs/skills) CLI) in one shot:

```bash
npx -y @mvanhorn/printing-press-library install postiee
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press-library install postiee --cli-only
```

For skill only — installs the skill into the same agents as the default command above, but skips the CLI binary (use this to update or reinstall just the skill):

```bash
npx -y @mvanhorn/printing-press-library install postiee --skill-only
```

To constrain the skill install to one or more specific agents (repeatable — agent names match the [`skills`](https://github.com/vercel-labs/skills) CLI):

```bash
npx -y @mvanhorn/printing-press-library install postiee --agent claude-code
npx -y @mvanhorn/printing-press-library install postiee --agent claude-code --agent codex
```

### Without Node (Go fallback)

If `npx` isn't available (no Node, offline), install the CLI directly via Go (requires Go 1.26.9 or newer):

```bash
go install github.com/mvanhorn/printing-press-library/library/social-and-messaging/postiee/cmd/postiee-pp-cli@latest
```

This installs the CLI only — no skill.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/postiee-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

Install the CLI binary first. The installer writes binaries to a per-user managed bin directory by default: `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows.

```bash
npx -y @mvanhorn/printing-press-library install postiee --cli-only
```

Then install the focused Hermes skill.

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-postiee --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-postiee --force
```

Restart the Hermes session or gateway if the newly installed skill is not visible immediately.

## Install for OpenClaw
Install both the CLI binary and the focused OpenClaw skill. The installer defaults binaries to a per-user bin directory (`$HOME/.local/bin` on macOS/Linux, `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows):

```bash
npx -y @mvanhorn/printing-press-library install postiee --agent openclaw
```

Restart the OpenClaw session or gateway if the newly installed skill is not visible immediately.

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/postiee-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.
3. Fill in `POSTIEE_MEMBER_OAUTH` when Claude Desktop prompts you.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


```bash
go install github.com/mvanhorn/printing-press-library/library/social-and-messaging/postiee/cmd/postiee-pp-mcp@latest
```

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "postiee": {
      "command": "postiee-pp-mcp",
      "env": {
        "POSTIEE_MEMBER_OAUTH": "<your-key>"
      }
    }
  }
}
```

</details>

## Quick Start

### 1. Install

See [Install](#install) above.

### 2. Set Up Credentials

Get your access token from your API provider's developer portal, then store it:

```bash
echo "$TOKEN" | postiee-pp-cli auth set-token
```

Or set it via environment variable:

```bash
export POSTIEE_MEMBER_OAUTH="your-token-here"
```

### 3. Verify Setup

```bash
postiee-pp-cli doctor
```

This checks your configuration and credentials.

### 4. Try Your First Command

```bash
postiee-pp-cli userinfo --authorization example-value
```

## Usage

Run `postiee-pp-cli --help` for the full command reference and flag list.

## Paths & environment variables

This CLI separates local files into four path kinds:

| Kind | Contents |
|------|----------|
| `config` | User-editable settings such as `config.toml` and saved profiles |
| `data` | Durable local data: `credentials.toml`, `data.db`, cookies, browser-session proof files, and other auth sidecars |
| `state` | Runtime state such as persisted queries, jobs, and `teach.log` |
| `cache` | Regenerable HTTP/cache files |

Each kind resolves independently. The ladder is:

1. Per-kind env var: `POSTIEE_CONFIG_DIR`, `POSTIEE_DATA_DIR`, `POSTIEE_STATE_DIR`, or `POSTIEE_CACHE_DIR`
2. `--home <dir>` for this invocation
3. `POSTIEE_HOME` for a flat relocated root
4. XDG env vars: `XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`
5. Platform defaults matching existing installs

For containers and agent sandboxes, prefer a single relocated root:

```bash
export POSTIEE_HOME=/srv/postiee
postiee-pp-cli doctor
```

Under `POSTIEE_HOME=/srv/postiee`, the four dirs resolve to `/srv/postiee/config`, `/srv/postiee/data`, `/srv/postiee/state`, and `/srv/postiee/cache`.

MCP servers do not receive CLI flags from the host. Put relocation in the host `env` block:

```json
{
  "mcpServers": {
    "postiee": {
      "command": "postiee-pp-mcp",
      "env": {
        "POSTIEE_HOME": "/srv/postiee"
      }
    }
  }
}
```

Precedence matters in fleets: an ambient per-kind variable such as `POSTIEE_DATA_DIR` overrides an explicit `--home` for that kind. Use `POSTIEE_HOME` or the per-kind variables for durable fleet relocation; treat `--home` as the weaker per-invocation lever.

Relocation is one-way. Unsetting `POSTIEE_HOME` does not move files back to platform defaults, and `doctor` cannot find credentials left under a former root. Move the files manually before unsetting relocation variables.

Existing installs keep working because the platform-default rung matches the legacy layout. On the first auth write, stored secrets leave `config.toml` and are consolidated into `credentials.toml` under the data directory. Run `postiee-pp-cli doctor --fail-on warn` to check path and credential-location warnings in automation.

## Commands

### images

Manage images

- **`postiee-pp-cli images`** - Two steps only. There is NO finalizeUpload for images — that step
belongs to the Videos API. The `image` field in this response is
already the final urn; the only remaining work is the PUT of the bytes.
`SYNCHRONOUS_UPLOAD` is not supported. Constraints: JPG/GIF/PNG, under
36,152,320 pixels, GIF up to 250 frames.

### posts

Manage posts

- **`postiee-pp-cli posts create-member`** - Fallback for apps that hold Community Management access. `lifecycleState`
accepts PUBLISHED only at creation; DRAFT and PUBLISH_REQUESTED are
response-only. There is no publishAt or scheduledAt field anywhere in
the schema, so scheduling is client-side or impossible.
- **`postiee-pp-cli posts delete`** - Idempotent. Returns 204 whether or not the post existed. Batch delete
is not supported. UNVERIFIED whether member-authored posts are
deletable with a self-serve app; the docs state no author restriction
but never confirm it for members.

### userinfo

Manage userinfo

- **`postiee-pp-cli userinfo`** - Returns `sub`, the app-scoped member identifier. Person IDs are unique
per application, so this MUST be called at runtime to build the author
URN. Never hardcode a person ID.


### Self-learning loop

This CLI caches per-question discovery so repeat queries skip the walk and structurally similar queries get answered via entity substitution. The loop also self-captures: every invocation is journaled locally, and failed-flag corrections plus fresh teaches surface as candidates on the next `recall` for confirm/reject judgment. Agents call `recall` before discovery and fire `teach &` after answering. See the `## Automatic learning` section in `SKILL.md` for the full protocol.

- **`postiee-pp-cli recall <query>`** - Look up cached resources for a query before running discovery
- **`postiee-pp-cli teach`** - Record a query -> resource mapping (silent on success, safe to background with `&`)
- **`postiee-pp-cli learnings list`** - Inspect taught rows
- **`postiee-pp-cli learnings forget <query>`** - Undo a teach
- **`postiee-pp-cli learnings candidates`** - List auto-captured candidates awaiting confirm/reject
- **`postiee-pp-cli learnings stats`** - Local loop metrics: recall hit rate, teach-to-reuse, playbook resolution, candidate counts
- **`postiee-pp-cli teach-pattern`** - Install a query/resource template up front
- **`postiee-pp-cli teach-lookup`** - Add an entity mapping (e.g. country code, team alias) for pattern substitution

Pass `--no-learn` or set `POSTIEE_NO_LEARN=true` to disable the loop for deterministic flows.

The local store's schema version stamp is one-way: once this version of `postiee-pp-cli` opens the database, older binaries refuse it with a version error — upgrade the binary rather than downgrading.

## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
postiee-pp-cli userinfo --authorization example-value

# JSON for scripting and agents
postiee-pp-cli userinfo --authorization example-value --json
# Filter to specific fields
postiee-pp-cli userinfo --authorization example-value --json --select email,email_verified,family_name

# Dry run — show the request without sending
postiee-pp-cli userinfo --authorization example-value --dry-run

# Agent mode — JSON + compact + no prompts in one flag
postiee-pp-cli userinfo --authorization example-value --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select <field>[,<field>...]` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Explicit retries** - add `--idempotent` to create retries and add `--ignore-missing` to delete retries when a no-op success is acceptable
- **Explicit confirmation** - `--agent` does not imply `--yes`; pass `--yes` separately only after the target, arguments, and side effects are clear
- **Piped input** - write commands can accept structured input when their help lists `--stdin`
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `6` partial failure, `7` rate limited, `10` config error.

## Health Check

```bash
postiee-pp-cli doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Run `postiee-pp-cli doctor` to see the resolved config, data, state, and cache directories. The platform-default config path is `~/.config/linkedin-member-posting-pp-cli/config.toml`; `--home`, `POSTIEE_HOME`, and per-kind env vars can relocate it.

Static request headers can be configured under `headers`; per-command header overrides take precedence.

Environment variables:

| Name | Kind | Required | Description |
| --- | --- | --- | --- |
| `POSTIEE_MEMBER_OAUTH` | per_call | Yes | Set to your API credential. |

### agentcookie (optional)

If you use agentcookie to sync secrets across machines, this CLI auto-adopts agentcookie-managed credentials with no extra setup. When the daemon writes to this CLI's config, `postiee-pp-cli doctor` reports `agentcookie: detected` and `auth-status` labels the source as `agentcookie`. Skip this section if you don't use agentcookie - the CLI works the same as any other.

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `postiee-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $POSTIEE_MEMBER_OAUTH`
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

---

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
