param(
    [string]$Version,
    [string]$InputJson,
    [string]$OutputJson
)

$ver = $Version.TrimStart('v')
$major = 0
$minor = 0
$patch = 0
$build = 0

if ($ver -match '^(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:[.-](\d+))?') {
    $major = [int]$Matches[1]
    if ($Matches[2]) {
        $minor = [int]$Matches[2]
    }
    if ($Matches[3]) {
        $patch = [int]$Matches[3]
    }
    if ($Matches[4]) {
        $build = [int]$Matches[4]
    }
}

$json = Get-Content $InputJson -Raw | ConvertFrom-Json
$json.FixedFileInfo.FileVersion.Major = $major
$json.FixedFileInfo.FileVersion.Minor = $minor
$json.FixedFileInfo.FileVersion.Patch = $patch
$json.FixedFileInfo.FileVersion.Build = $build
$json.FixedFileInfo.ProductVersion.Major = $major
$json.FixedFileInfo.ProductVersion.Minor = $minor
$json.FixedFileInfo.ProductVersion.Patch = $patch
$json.FixedFileInfo.ProductVersion.Build = $build
$json.StringFileInfo.FileVersion = "$major.$minor.$patch.$build"
$json.StringFileInfo.ProductVersion = $ver

$json | ConvertTo-Json -Depth 10 | Set-Content $OutputJson
