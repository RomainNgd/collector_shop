param(
    [string]$Tag = "local",
    [string]$Registry = "",
    [switch]$NoCache
)

$ErrorActionPreference = "Stop"

function Get-ImageName {
    param([string]$Name)

    if ([string]::IsNullOrWhiteSpace($Registry)) {
        return "collector-shop/$Name`:$Tag"
    }

    return "$Registry/$Name`:$Tag"
}

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\\..")
$apiImage = Get-ImageName "go-api"
$spaImage = Get-ImageName "collector-spa"
$cacheArgs = @()

if ($NoCache) {
    $cacheArgs += "--no-cache"
}

Write-Host "Building $apiImage"
docker build @cacheArgs -f "$repoRoot\\build\\docker\\go-api.Dockerfile" -t $apiImage $repoRoot
if ($LASTEXITCODE -ne 0) {
    throw "Docker build failed for $apiImage"
}

Write-Host "Building $spaImage"
docker build @cacheArgs -f "$repoRoot\\build\\docker\\collector-spa.Dockerfile" -t $spaImage $repoRoot
if ($LASTEXITCODE -ne 0) {
    throw "Docker build failed for $spaImage"
}

Write-Host ""
Write-Host "Images built successfully:"
Write-Host " - $apiImage"
Write-Host " - $spaImage"
