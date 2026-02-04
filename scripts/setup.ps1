# Quick Setup Script for Notifier Service
# Ensures database tables are created before starting the service

Write-Host "=== Notifier Service Setup ===" -ForegroundColor Cyan

# Check if database is running
Write-Host "`nChecking database connection..." -ForegroundColor Yellow
$dbTest = try {
    $null = Test-NetConnection -ComputerName localhost -Port 5434 -WarningAction SilentlyContinue -ErrorAction Stop
    $true
} catch {
    $false
}

if (-not $dbTest) {
    Write-Host "❌ Database not running on port 5434" -ForegroundColor Red
    Write-Host "Start database with: docker-compose up -d" -ForegroundColor Yellow
    exit 1
}

Write-Host "✅ Database is running" -ForegroundColor Green

# Run sms_templates migration
Write-Host "`nCreating SMS templates table..." -ForegroundColor Yellow
cd E:\Projects\Minisource\notifier
go run scripts/create_sms_templates.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Database setup complete!" -ForegroundColor Green
    Write-Host "`nYou can now start the notifier service:" -ForegroundColor Cyan
    Write-Host "  go run cmd/server/main.go" -ForegroundColor Gray
} else {
    Write-Host "`n❌ Setup failed" -ForegroundColor Red
    exit 1
}
