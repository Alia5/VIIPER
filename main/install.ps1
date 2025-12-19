$ErrorActionPreference = "Stop"

$viiperVersion = "dev-snapshot"

$repo = "Alia5/VIIPER"
$apiUrl = "https://api.github.com/repos/$repo/releases/tags/$viiperVersion"

Write-Host "Fetching VIIPER release: $viiperVersion..."
$releaseData = Invoke-RestMethod -Uri $apiUrl -ErrorAction Stop
$version = $releaseData.tag_name

if (-not $version) {
    Write-Host "Error: Could not fetch VIIPER release" -ForegroundColor Red
    exit 1
}

Write-Host "Version: $version"

$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else {
    Write-Host "Error: Only 64-bit Windows is supported" -ForegroundColor Red
    exit 1
}

if ((Get-CimInstance Win32_ComputerSystem).SystemType -match "ARM") {
    $arch = "arm64"
}

$binaryName = "viiper-windows-$arch.exe"
$downloadUrl = "https://github.com/$repo/releases/download/$version/$binaryName"

Write-Host "Downloading from: $downloadUrl"
$tempDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }

try {
    $tempViiper = Join-Path $tempDir "viiper.exe"
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempViiper -ErrorAction Stop
    
    $installDir = Join-Path $env:LOCALAPPDATA "VIIPER"
    $installPath = Join-Path $installDir "viiper.exe"
    $isUpdate = Test-Path $installPath
    
    Write-Host "Installing binary to $installPath..."
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null

    # On update, preserve the user's previous autostart preference.
    # Autostart is controlled by `viiper install`/`viiper uninstall` (registry Run key).
    # If we ran `install` again here we'd re-enable autostart for users who disabled it.
    # So: if the binary already exists, just replace it.
    if ($isUpdate) {
        Write-Host "Existing VIIPER installation detected (update). Preserving startup/autostart configuration..."
        # If the old binary is running, Windows will usually refuse to overwrite it.
        # Stop any processes whose ExecutablePath matches the install path.
        $procs = Get-CimInstance Win32_Process -ErrorAction SilentlyContinue |
        Where-Object { $_.ExecutablePath -eq $installPath }
        if ($procs) {
            Write-Host "Stopping running VIIPER instance(s) so the binary can be updated..."
            foreach ($p in $procs) {
                try {
                    Stop-Process -Id $p.ProcessId -Force -ErrorAction SilentlyContinue
                }
                catch { }
            }
        }
    }

    Copy-Item $tempViiper $installPath -Force

    if (-not $isUpdate) {
        Write-Host "Configuring system startup..."
        & $installPath install
    }
    
    Write-Host "VIIPER installed successfully!" -ForegroundColor Green
    Write-Host "Binary installed to: $installPath"
    if ($isUpdate) {
        Write-Host "Update complete. Startup/autostart configuration was left unchanged."
        Write-Host "If VIIPER was running, restart it to use the updated binary."
    }
    else {
        Write-Host "VIIPER server is now running and will start automatically on boot."
    }
}
finally {
    Remove-Item -Recurse -Force $tempDir -ErrorAction SilentlyContinue
}
