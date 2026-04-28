$ErrorActionPreference = "Stop"

$viiperVersion = "v0.6.0"

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
    function Get-ViiperVersion($path) {
        try {
            $help = & $path --help -p 2>$null
            $match = ($help | Select-String -Pattern "Version:\s*([^\s]+)" -AllMatches | Select-Object -First 1)
            if ($match) {
                return $match.Matches[0].Groups[1].Value
            }
        }
        catch { }
        return $null
    }

    function Parse-VersionOrNull($ver) {
        if (-not $ver) { return $null }
        $clean = $ver.Trim().TrimStart('v', 'V')
        $clean = $clean.Split('-')[0]
        try { return [Version]$clean }
        catch { return $null }
    }

    $tempViiper = Join-Path $tempDir "viiper.exe"
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempViiper -ErrorAction Stop

    $newVersion = Get-ViiperVersion $tempViiper
    if (-not $newVersion) { $newVersion = "unknown" }
    Write-Host "Downloaded VIIPER version: $newVersion"
    
    $installDir = Join-Path $env:LOCALAPPDATA "VIIPER"
    $installPath = Join-Path $installDir "viiper.exe"
    $isUpdate = Test-Path $installPath
    $skipInstall = $false

    $oldVersion = "unknown"
    if ($isUpdate) {
        Write-Host "Existing VIIPER installation detected. Preserving startup/autostart configuration..."
        $oldVersionRaw = Get-ViiperVersion $installPath
        if ($oldVersionRaw) { $oldVersion = $oldVersionRaw }
        Write-Host "Installed VIIPER version: $oldVersion"

        $newV = Parse-VersionOrNull $newVersion
        $oldV = Parse-VersionOrNull $oldVersion

        if ($newVersion -eq $oldVersion -and $newVersion -ne "unknown") {
            Write-Host "Versions are identical. Skipping VIIPER install step."
            $skipInstall = $true
        }
        elseif ($newV -and $oldV -and $newV -lt $oldV) {
            Write-Host "Detected potential downgrade (installed: $oldVersion, new: $newVersion). Skipping install." -ForegroundColor Yellow
            $skipInstall = $true
        }
    }
    
    if (-not $skipInstall) {
        Write-Host "Installing binary to $installPath..."
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null

        if ($isUpdate) {
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
                Start-Sleep -Milliseconds 500
            }
        }

        Copy-Item $tempViiper $installPath -Force
    }

    Write-Host ""
    Write-Host "Checking USBIP drivers..." -ForegroundColor Cyan

    $usbipTargetVersion = [Version]"0.9.7.7"
    $usbipInstalledVersion = $null

    $usbipEntry = Get-ItemProperty "HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*" -ErrorAction SilentlyContinue |
        Where-Object { $_.DisplayName -eq 'USBip' } |
        Select-Object -First 1
    if (-not $usbipEntry) {
        $usbipEntry = Get-ItemProperty "HKLM:\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*" -ErrorAction SilentlyContinue |
            Where-Object { $_.DisplayName -eq 'USBip' } |
            Select-Object -First 1
    }
    if ($usbipEntry) {
        try { $usbipInstalledVersion = [Version]$usbipEntry.DisplayVersion } catch { }
    }

    if (-not $usbipInstalledVersion) {
        $driverPath = Join-Path $env:SystemRoot "System32\drivers\usbip2_ude.sys"
        if (Test-Path $driverPath) {
            try { $usbipInstalledVersion = [Version](Get-Item $driverPath).VersionInfo.FileVersion } catch { }
        }
    }

    $needsReboot = $false
    $needsUsbipInstall = $true

    if ($usbipInstalledVersion) {
        if ($usbipInstalledVersion -ge $usbipTargetVersion) {
            Write-Host "USBIP drivers already up to date (installed: $usbipInstalledVersion)" -ForegroundColor Green
            $needsUsbipInstall = $false
        }
        else {
            Write-Host "USBIP drivers outdated (installed: $usbipInstalledVersion, required: $usbipTargetVersion). Updating..." -ForegroundColor Yellow
        }
    }
    else {
        Write-Host "USBIP drivers not found. Installing..." -ForegroundColor Yellow
    }

    if ($needsUsbipInstall) {
        Write-Host "This requires administrator privileges." -ForegroundColor Yellow

        $usbipInstallerUrl = "https://github.com/vadimgrn/usbip-win2/releases/download/v.0.9.7.7/USBip-0.9.7.7-x64.exe"
        $usbipInstaller = Join-Path $tempDir "USBip-setup.exe"

        try {
            Write-Host "  Downloading usbip-win2 installer..." -ForegroundColor Cyan
            Invoke-WebRequest -Uri $usbipInstallerUrl -OutFile $usbipInstaller -ErrorAction Stop
            Write-Host "Installing USBIP drivers (UAC prompt will appear)..." -ForegroundColor Yellow
            Start-Process -FilePath $usbipInstaller -ArgumentList "/S" -Verb RunAs -Wait
            Write-Host "USBIP drivers installed/updated successfully" -ForegroundColor Green
            $needsReboot = $true
        }
        catch {
            Write-Host "Warning: Failed to install USBIP drivers - $($_.Exception.Message)" -ForegroundColor Yellow
            Write-Host "You may need to install usbip-win2 manually from:" -ForegroundColor Yellow
            Write-Host "  https://github.com/vadimgrn/usbip-win2/releases" -ForegroundColor Yellow
        }
    }

    if (-not $isUpdate) {
        Write-Host "Configuring system startup..."
    }
    Start-Process -WindowStyle Hidden  "$installPath" -ArgumentList "install"
    
    Write-Host "VIIPER installed successfully!" -ForegroundColor Green
    Write-Host "Binary installed to: $installPath"
    if ($isUpdate) {
        if ($skipInstall) {
            Write-Host "Binary already at correct version or newer. Skipping binary copy."
        }
        else {
            Write-Host "Update complete. Startup/autostart configuration was left unchanged."
        }
        Write-Host "VIIPER service has been restarted."
    }
    else {
        Write-Host "VIIPER server is now running and will start automatically on boot."
    }
    
    if ($needsReboot) {
        Write-Host ""
        Write-Host "IMPORTANT: A system reboot is required for USBIP drivers to function properly." -ForegroundColor Yellow
        Write-Host "Please restart your computer before using VIIPER." -ForegroundColor Yellow
    }
}
finally {
    Remove-Item -Recurse -Force $tempDir -ErrorAction SilentlyContinue
}
