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

# grove owns this file outright: rewrite it on every run. It defines the shell
# wrapper that turns the path grove prints on stdout into a directory change,
# so $PROFILE only ever needs the stable source line below.
$InitFile = Join-Path $InstallDir "init.ps1"
$InitContent = @"
function Invoke-Grove {
    param(
        [Parameter(ValueFromRemainingArguments = `$true)]
        [string[]]`$Arguments
    )

    `$output = & "$BinaryPath" @Arguments
    if (`$LASTEXITCODE -ne 0) {
        return `$LASTEXITCODE
    }

    if (-not [string]::IsNullOrWhiteSpace(`$output)) {
        Set-Location `$output
    }
}

Set-Alias grove Invoke-Grove
"@
Set-Content -Path $InitFile -Value $InitContent

$SourceLine = ". `"$InitFile`""
$profileContent = Get-Content -Raw -Path $ProfilePath
if ($null -eq $profileContent) { $profileContent = '' }
if (-not $profileContent.Contains($SourceLine)) {
    Add-Content -Path $ProfilePath -Value "`r`n$SourceLine"
}

Write-Output "installed grove to $BinaryPath, wrote $InitFile, and updated $ProfilePath"
