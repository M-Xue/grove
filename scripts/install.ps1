$ErrorActionPreference = "Stop"

$InstallDir = if ($env:GROVE_INSTALL_DIR) { $env:GROVE_INSTALL_DIR } else { Join-Path $HOME "AppData\Local\Programs\grove" }
$BinaryPath = Join-Path $InstallDir "grove.exe"
$ProfilePath = $PROFILE

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "go is required to install grove"
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
go build -o $BinaryPath .

$ProfileDir = Split-Path -Parent $ProfilePath
if (-not (Test-Path $ProfileDir)) {
    New-Item -ItemType Directory -Force -Path $ProfileDir | Out-Null
}
if (-not (Test-Path $ProfilePath)) {
    New-Item -ItemType File -Force -Path $ProfilePath | Out-Null
}

$InitBlock = & $BinaryPath shell-init powershell
$profileContent = Get-Content -Raw -Path $ProfilePath
if (-not $profileContent.Contains('Set-Alias grove Invoke-Grove')) {
    Add-Content -Path $ProfilePath -Value "`r`n$InitBlock"
}

Write-Output "installed grove to $BinaryPath and updated $ProfilePath"
