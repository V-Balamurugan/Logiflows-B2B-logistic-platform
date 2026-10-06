Write-Host "=== Starting LogiFlows Development Environment ===" -ForegroundColor Cyan

if (Get-Command docker -ErrorAction SilentlyContinue) {
    Write-Host "Starting Docker Compose services (Postgres PostGIS, Redis)..." -ForegroundColor Yellow
    docker compose up -d postgres redis
    Write-Host "Infrastructure containers active." -ForegroundColor Green
} else {
    Write-Host "Docker not found; skipping container spinup." -ForegroundColor Yellow
}
