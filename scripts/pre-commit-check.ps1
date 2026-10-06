$ErrorActionPreference = "Stop"
Write-Host "=== LogiFlows Pre-Commit Quality & Security Verification ===" -ForegroundColor Cyan

# 1. Secret Scanner
Write-Host "`n[1/3] Scanning for accidentally staged secrets / raw keys..." -ForegroundColor Yellow
$stagedFiles = git diff --cached --name-only
$secretPatterns = @(
    "BEGIN PRIVATE KEY",
    "BEGIN RSA PRIVATE KEY",
    "password\s*=\s*['`"][^'`"]+['`"]",
    "JWT_SECRET\s*=\s*['`"][a-zA-Z0-9_\-]+['`"]",
    "CLOUDINARY_API_SECRET\s*=\s*[a-zA-Z0-9_\-]+"
)

$secretFound = $false
foreach ($f in $stagedFiles) {
    if ($f -match "\.env$" -and $f -notmatch "\.env\.example$") {
        Write-Host "CRITICAL ERROR: Staged private environment file: $f" -ForegroundColor Red
        $secretFound = $true
    }
}

if ($secretFound) {
    Write-Host "Commit blocked: Secrets or private credentials detected in staged changes." -ForegroundColor Red
    exit 1
} else {
    Write-Host "  -> Secret scan clean." -ForegroundColor Green
}

# 2. Syntax & Linters / Formatter Check
Write-Host "`n[2/3] Verifying code hygiene..." -ForegroundColor Yellow
Push-Location "$PSScriptRoot\..\backend"
try {
    go vet ./...
    Write-Host "  -> Go vet passed." -ForegroundColor Green
} catch {
    Write-Host "  -> Go vet failed: $_" -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
}

# 3. Unit tests of touched modules / core suites
Write-Host "`n[3/3] Running module tests..." -ForegroundColor Yellow
Push-Location "$PSScriptRoot\..\backend"
try {
    go test ./...
    Write-Host "  -> Go tests passed." -ForegroundColor Green
} catch {
    Write-Host "  -> Go tests failed: $_" -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
}

Write-Host "`nAll pre-commit checks PASSED." -ForegroundColor Green
exit 0
