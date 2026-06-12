function Invoke-Grove {
    param(
        [Parameter(ValueFromRemainingArguments = $true)]
        [string[]]$Arguments
    )

    $output = & grove.exe @Arguments
    if ($LASTEXITCODE -ne 0) {
        return $LASTEXITCODE
    }

    if (-not [string]::IsNullOrWhiteSpace($output)) {
        Set-Location $output
    }
}

Set-Alias grove Invoke-Grove
