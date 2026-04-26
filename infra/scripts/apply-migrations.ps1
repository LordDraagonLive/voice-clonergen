param(
  [string]$DatabaseUrl = $env:DATABASE_URL
)

if (-not $DatabaseUrl) {
  Write-Error "DATABASE_URL is required"
  exit 1
}

psql $DatabaseUrl -f apps/api/internal/db/migrations/001_init.sql
