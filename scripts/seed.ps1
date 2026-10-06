Write-Host "=== Seeding LogiFlows Initial Test Data & Admin ===" -ForegroundColor Cyan
Push-Location "$PSScriptRoot\..\backend"
try {
    go run cmd/seed-admin/main.go
    Write-Host "Seed completed successfully." -ForegroundColor Green
} catch {
    Write-Host "Seed failed: $_" -ForegroundColor Red
} finally {
    Pop-Location
}
