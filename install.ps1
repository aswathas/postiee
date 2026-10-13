# Postiee installer for Windows.
#
#   irm https://raw.githubusercontent.com/aswathas/postiee/main/install.ps1 | iex
#
# Downloads the prebuilt binary for your platform. No Go toolchain required.
# To a file instead of running it:
#
#   irm https://raw.githubusercontent.com/aswathas/postiee/main/install.ps1 -OutFile install.ps1
#   .\install.ps1

[CmdletBinding()]
param(
    # Install somewhere other than the default.
    [string]$InstallDir,

    # Pin a release tag. Defaults to the latest published release.
    [string]$Version = 'latest'
)

$ErrorActionPreference = 'Stop'

$repo = 'aswathas/postiee'
$bin  = 'postiee'

function Write-Step  { param($m) Write-Host $m -ForegroundColor Cyan }
function Write-Ok    { param($m) Write-Host $m -ForegroundColor Green }
function Write-Note  { param($m) Write-Host $m -ForegroundColor Yellow }

# ---- platform --------------------------------------------------------------
$osLabel = switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'x86_64'; break }
    'ARM64' { 'arm64';  break }
    default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

# ---- install directory -----------------------------------------------------
if (-not $InstallDir) {
    # Prefer a per-user install: no admin rights, and it works on locked-down
    # machines where Program Files is not writable.
    $localBin = Join-Path $HOME '.local\bin'
    if (-not (Test-Path $localBin)) { New-Item -ItemType Directory -Force -Path $localBin | Out-Null }
    $InstallDir = $localBin
}

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($userPath -notlike "*$InstallDir*") {
    Write-Note "$InstallDir is not on your PATH — adding it for the current user."
    [Environment]::SetEnvironmentVariable('Path', "$userPath;$InstallDir", 'User')
    $env:Path = "$env:Path;$InstallDir"
}

# ---- download --------------------------------------------------------------
if ($Version -eq 'latest') {
    $url = "https://github.com/$repo/releases/latest/download/${bin}_Windows_${osLabel}.zip"
    Write-Step "Downloading $bin for Windows/$osLabel (latest)..."
} else {
    $url = "https://github.com/$repo/releases/download/$Version/${bin}_Windows_${osLabel}.zip"
    Write-Step "Downloading $bin for Windows/$osLabel ($Version)..."
}

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("postiee-" + [System.Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Force -Path $tmp | Out-Null

try {
    $zip = Join-Path $tmp 'postiee.zip'
    Invoke-WebRequest -Uri $url -OutFile $zip -UseBasicParsing

    Expand-Archive -Path $zip -DestinationPath $tmp -Force

    $exe = Get-ChildItem -Path $tmp -Filter "$bin.exe" -Recurse | Select-Object -First 1
    if (-not $exe) { throw "Archive did not contain $bin.exe" }

    $target = Join-Path $InstallDir "$bin.exe"
    Copy-Item -Path $exe.FullName -Destination $target -Force

    Write-Ok "installed $bin -> $target"
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}

Write-Host ''
Write-Host 'next:' -ForegroundColor Cyan
Write-Host '  1. create a LinkedIn app at https://www.linkedin.com/developers/apps'
Write-Host "     Products -> add 'Share on LinkedIn' and 'Sign In with LinkedIn'"
Write-Host '     Auth -> add redirect URL: http://localhost:8085/callback'
Write-Host '  2. write your credentials to %USERPROFILE%\.config\postiee\credentials.toml:'
Write-Host '       client_id = "..."'
Write-Host '       client_secret = "..."'
Write-Host '  3. postiee auth login --port 8085'
Write-Host ''