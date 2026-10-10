$ErrorActionPreference = "Stop"

$viiperVersion = "v0.9.0"

$repo = "Alia5/VIIPER"
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/tags/$viiperVersion"

$arch = "amd64"
if ((Get-CimInstance Win32_ComputerSystem).SystemType -match "ARM") {
    $arch = "arm64"
}

$installer = Join-Path ([System.IO.Path]::GetTempPath()) "viiper.exe"
Invoke-WebRequest -Uri "https://github.com/$repo/releases/download/$($release.tag_name)/viiper-windows-$arch.exe" -OutFile $installer

& $installer install --quiet
