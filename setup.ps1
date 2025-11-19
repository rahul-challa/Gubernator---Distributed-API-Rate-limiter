# Gubernator Setup Script for Windows
# This script helps you set up and run the Gubernator rate limiter

Write-Host "Gubernator Setup Script" -ForegroundColor Cyan
Write-Host "=========================" -ForegroundColor Cyan
Write-Host ""

# Check for Go
Write-Host "Checking for Go..." -ForegroundColor Yellow
$goInstalled = Get-Command go -ErrorAction SilentlyContinue
if (-not $goInstalled) {
    Write-Host "[ERROR] Go is not installed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install Go first:" -ForegroundColor Yellow
    Write-Host "1. Visit: https://go.dev/dl/" -ForegroundColor White
    Write-Host "2. Download the Windows installer" -ForegroundColor White
    Write-Host "3. Run the installer" -ForegroundColor White
    Write-Host "4. Restart your terminal and run this script again" -ForegroundColor White
    Write-Host ""
    Write-Host "Press any key to open the download page..."
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    Start-Process "https://go.dev/dl/"
    exit 1
} else {
    $goVersion = go version
    Write-Host "[OK] $goVersion" -ForegroundColor Green
}

Write-Host ""

# Check for Docker
Write-Host "Checking for Docker..." -ForegroundColor Yellow
$dockerInstalled = Get-Command docker -ErrorAction SilentlyContinue
if (-not $dockerInstalled) {
    Write-Host "[WARNING] Docker is not installed (optional)" -ForegroundColor Yellow
    Write-Host "   You can use cloud Redis or install Docker later" -ForegroundColor Gray
} else {
    $dockerVersion = docker --version
    Write-Host "[OK] $dockerVersion" -ForegroundColor Green
}

Write-Host ""

# Check for Python
Write-Host "Checking for Python..." -ForegroundColor Yellow
$pythonInstalled = Get-Command python -ErrorAction SilentlyContinue
if (-not $pythonInstalled) {
    Write-Host "[WARNING] Python is not installed (optional, for visualization)" -ForegroundColor Yellow
} else {
    $pythonVersion = python --version
    Write-Host "[OK] $pythonVersion" -ForegroundColor Green
}

Write-Host ""
Write-Host "Installing Go dependencies..." -ForegroundColor Yellow
go mod download
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] Dependencies installed" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Failed to install dependencies" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Cyan
Write-Host "1. Start Redis (choose one):" -ForegroundColor White
if ($dockerInstalled) {
    Write-Host "   docker run -d -p 6379:6379 --name gubernator-redis redis:7-alpine" -ForegroundColor Gray
}
Write-Host "   OR use a cloud Redis service (Upstash, Redis Cloud)" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Run the server:" -ForegroundColor White
Write-Host "   go run cmd/server/main.go" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Test it (in another terminal):" -ForegroundColor White
Write-Host "   curl http://localhost:8080/health" -ForegroundColor Gray
Write-Host ""
Write-Host "4. Run visualization (optional):" -ForegroundColor White
if ($pythonInstalled) {
    Write-Host "   pip install requests" -ForegroundColor Gray
    Write-Host "   python scripts/flood_test.py" -ForegroundColor Gray
} else {
    Write-Host "   (Install Python first)" -ForegroundColor Gray
}
Write-Host ""

$startRedis = Read-Host "Do you want to start Redis with Docker now? (y/n)"
if ($startRedis -eq "y" -and $dockerInstalled) {
    Write-Host "Starting Redis..." -ForegroundColor Yellow
    docker run -d -p 6379:6379 --name gubernator-redis redis:7-alpine
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Redis started!" -ForegroundColor Green
        Write-Host ""
        $startServer = Read-Host "Do you want to start the server now? (y/n)"
        if ($startServer -eq "y") {
            Write-Host "Starting server..." -ForegroundColor Yellow
            Write-Host "Server will be available at http://localhost:8080" -ForegroundColor Cyan
            Write-Host "Press Ctrl+C to stop" -ForegroundColor Yellow
            Write-Host ""
            go run cmd/server/main.go
        }
    } else {
        Write-Host "[ERROR] Failed to start Redis" -ForegroundColor Red
    }
} elseif ($startRedis -eq "y" -and -not $dockerInstalled) {
    Write-Host "Docker is not installed. Please start Redis manually." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Setup complete! Check QUICKSTART.md for more details." -ForegroundColor Green

