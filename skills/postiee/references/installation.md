# Installing postiee

## One-liners

**macOS / Linux**
```sh
curl -fsSL https://raw.githubusercontent.com/aswathas/postiee/main/install.sh | sh
```

**Windows (PowerShell)**
```powershell
irm https://raw.githubusercontent.com/aswathas/postiee/main/install.ps1 | iex
```

Both download a prebuilt binary. No Go toolchain needed.

## What the installers do

- Detect OS and CPU architecture
- Prefer `/usr/local/bin` when writable, else `~/.local/bin` (macOS/Linux)
- Prefer `%USERPROFILE%\.local\bin` on Windows — no admin rights required
- Warn when the install directory is not on `PATH`
- Install atomically so a concurrent run cannot leave a half-written binary

If the install directory is not on `PATH`, add it:

```sh
# macOS / Linux, ~/.zshrc or ~/.bashrc
export PATH="$HOME/.local/bin:$PATH"
```

```powershell
# Windows, PowerShell profile
$env:Path += ";$HOME\.local\bin"
```

## Install from source

Requires Go 1.26.9 or newer.

```sh
git clone https://github.com/aswathas/postiee.git
cd postiee
go build -o postiee ./cmd/postiee
sudo mv postiee /usr/local/bin/   # or: mv postiee ~/.local/bin/
```

## Pin a version

```sh
curl -fsSL https://raw.githubusercontent.com/aswathas/postiee/main/install.sh | POSTIEE_VERSION=v1.0.0 sh
```

```powershell
$env:POSTIEE_VERSION="v1.0.0"; irm https://raw.githubusercontent.com/aswathas/postiee/main/install.ps1 | iex
```

## Supported platforms

| OS | Architectures | Asset |
|---|---|---|
| macOS | x86_64, arm64 | `postiee_Darwin_{arch}.tar.gz` |
| Linux | x86_64, arm64 | `postiee_Linux_{arch}.tar.gz` |
| Windows | x86_64, arm64 | `postiee_Windows_{arch}.zip` |

## Verify

```sh
postiee --version
postiee doctor
```

`doctor` checks config, credentials, and that `api.linkedin.com` is reachable.
Credentials present but "not verified" is normal until a command makes a real
request.

## Setup after install

1. <https://www.linkedin.com/developers/apps> → Create app
2. **Products** → add **Share on LinkedIn**
3. **Products** → add **Sign In with LinkedIn**
4. **Auth** → redirect URL `http://localhost:8085/callback`
5. Write `~/.config/postiee/credentials.toml`:

   ```toml
   client_id = "..."
   client_secret = "..."
   ```

6. Lock it down — the secret lives in this file:
   ```sh
   mkdir -p ~/.config/postiee && chmod 700 ~/.config/postiee
   chmod 600 ~/.config/postiee/credentials.toml
   ```
7. ```sh
   postiee auth login --port 8085
   ```

## Where things are stored

| | macOS / Linux | Windows |
|---|---|---|
| Credentials | `~/.config/postiee/` | `%USERPROFILE%\.config\postiee\` |
| Token | `~/.local/share/postiee/` | `%LOCALAPPDATA%\postiee\` |
| Post history | `~/.config/postiee/history.db` | `%USERPROFILE%\.config\postiee\history.db` |

None of these are ever committed. The repository's `.gitignore` blocks
`credentials.toml`, `config.toml`, and `*.db` as a second line of defence.