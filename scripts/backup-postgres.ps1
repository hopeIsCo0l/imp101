$ErrorActionPreference = "Stop"

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$backupDir = "backups"
if (-not (Test-Path $backupDir)) {
  New-Item -ItemType Directory -Path $backupDir | Out-Null
}

$dbUser = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
$dbName = if ($env:DB_NAME) { $env:DB_NAME } else { "imp101" }

$outputFile = Join-Path $backupDir "imp101-$timestamp.sql"
docker exec -e PGPASSWORD=$env:DB_PASSWORD imp101-postgres pg_dump -U $dbUser $dbName > $outputFile
Write-Host "Backup created at $outputFile"
