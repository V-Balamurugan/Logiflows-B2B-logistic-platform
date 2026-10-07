Write-Host "=== LogiFlows Development Environment Tool Check ===" -ForegroundColor Cyan

function Test-ExecutableVersion {
    param (
        [string]$Name,
        [scriptblock]$CheckBlock,
        [string]$InstallHint = ""
    )
    Write-Host -NoNewline "Checking $Name... "
    try {
        $ver = & $CheckBlock 2>&1 | Out-String
        if ($LASTEXITCODE -eq 0 -or $ver) {
            $first = ($ver.Trim() -split "`n")[0]
            Write-Host "FOUND: $first" -ForegroundColor Green
            return $true
        } else {
            Write-Host "NOT FOUND" -ForegroundColor Yellow
            if ($InstallHint) { Write-Host "  -> Install hint: $InstallHint" -ForegroundColor DarkGray }
            return $false
        }
    } catch {
        Write-Host "NOT FOUND" -ForegroundColor Yellow
        if ($InstallHint) { Write-Host "  -> Install hint: $InstallHint" -ForegroundColor DarkGray }
        return $false
    }
}

Test-ExecutableVersion -Name "Git" -CheckBlock { git --version } -InstallHint "winget install --id Git.Git -e --source winget"
Test-ExecutableVersion -Name "GitHub CLI (gh)" -CheckBlock { gh --version } -InstallHint "winget install --id GitHub.cli -e --source winget"
Test-ExecutableVersion -Name "Docker" -CheckBlock { docker --version } -InstallHint "Install Docker Desktop: https://www.docker.com/products/docker-desktop/"
Test-ExecutableVersion -Name "Docker Compose" -CheckBlock { docker compose version } -InstallHint "Included with Docker Desktop"
Test-ExecutableVersion -Name "Go (stable)" -CheckBlock { go version } -InstallHint "winget install --id GoLang.Go -e --source winget"
Test-ExecutableVersion -Name "Node.js (LTS)" -CheckBlock { node -v } -InstallHint "winget install --id OpenJS.NodeJS.LTS -e --source winget"
Test-ExecutableVersion -Name "npm" -CheckBlock { npm -v } -InstallHint "Installed with Node.js"
Test-ExecutableVersion -Name "Python 3.11+" -CheckBlock { python --version } -InstallHint "winget install --id Python.Python.3.11 -e --source winget"
Test-ExecutableVersion -Name "Flutter" -CheckBlock { flutter --version } -InstallHint "https://docs.flutter.dev/get-started/install/windows"
Test-ExecutableVersion -Name "Android (adb)" -CheckBlock { adb version } -InstallHint "Install Android Studio and platform-tools"

Write-Host "=================================================" -ForegroundColor Cyan
