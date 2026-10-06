$ErrorActionPreference = "Stop"
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "       Running LogiFlows Test Suite       " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

$failed = $false

# 1. Backend Go Tests
Write-Host "`n[1/4] Running Go Core Tests..." -ForegroundColor Yellow
Push-Location "$PSScriptRoot\..\backend"
try {
    go test ./...
    if ($LASTEXITCODE -ne 0) { throw "Go tests failed" }
    Write-Host "  -> Go tests PASSED" -ForegroundColor Green
} catch {
    Write-Host "  -> Go tests FAILED: $_" -ForegroundColor Red
    $failed = $true
} finally {
    Pop-Location
}

# 2. Python AI Service Tests
Write-Host "`n[2/4] Running Python AI Service Tests..." -ForegroundColor Yellow
Push-Location "$PSScriptRoot\..\ai"
try {
    if (Test-Path ".\.venv\Scripts\pytest.exe") {
        & ".\.venv\Scripts\pytest.exe"
    } else {
        pytest
    }
    if ($LASTEXITCODE -ne 0) { throw "Python tests failed" }
    Write-Host "  -> Python tests PASSED" -ForegroundColor Green
} catch {
    Write-Host "  -> Python tests FAILED: $_" -ForegroundColor Red
    $failed = $true
} finally {
    Pop-Location
}

# 3. Web Frontend Tests
Write-Host "`n[3/4] Running Web Frontend Tests..." -ForegroundColor Yellow
Push-Location "$PSScriptRoot\..\web"
try {
    npm test
    if ($LASTEXITCODE -ne 0) { throw "Web tests failed" }
    Write-Host "  -> Web tests PASSED" -ForegroundColor Green
} catch {
    Write-Host "  -> Web tests FAILED: $_" -ForegroundColor Red
    $failed = $true
} finally {
    Pop-Location
}

# 4. Mobile Flutter Tests
Write-Host "`n[4/4] Running Flutter Mobile Tests..." -ForegroundColor Yellow
Push-Location "$PSScriptRoot\..\mobile"
try {
    flutter test
    if ($LASTEXITCODE -ne 0) { throw "Flutter tests failed" }
    Write-Host "  -> Flutter tests PASSED" -ForegroundColor Green
} catch {
    Write-Host "  -> Flutter tests FAILED: $_" -ForegroundColor Red
    $failed = $true
} finally {
    Pop-Location
}

Write-Host "`n==========================================" -ForegroundColor Cyan
if ($failed) {
    Write-Host "  FAIL: One or more test suites failed!   " -ForegroundColor Red
    Write-Host "==========================================" -ForegroundColor Cyan
    exit 1
} else {
    Write-Host "  PASS: All 4 test suites passed cleanly! " -ForegroundColor Green
    Write-Host "==========================================" -ForegroundColor Cyan
    exit 0
}
