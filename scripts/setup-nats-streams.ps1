# NATS Jetstream Stream Setup Script (Windows PowerShell)
# Creates streams for all services in GIIA Core Engine

$ErrorActionPreference = "Stop"

$NatsUrl = if ($env:NATS_URL) { $env:NATS_URL } else { "nats://localhost:4222" }

Write-Host "Setting up NATS Jetstream streams..." -ForegroundColor Green
Write-Host "NATS URL: $NatsUrl" -ForegroundColor Cyan
Write-Host ""

function New-NatsStream {
    param (
        [string]$StreamName,
        [string]$Subjects
    )

    Write-Host "Creating stream: $StreamName" -ForegroundColor Yellow

    try {
        nats stream add $StreamName `
            --subjects="$Subjects" `
            --storage=file `
            --retention=limits `
            --max-age=7d `
            --max-bytes=1G `
            --replicas=1 `
            --discard=old `
            --max-msg-size=8MB `
            --dupe-window=2m `
            --server="$NatsUrl" `
            --defaults
    } catch {
        Write-Host "Stream $StreamName may already exist" -ForegroundColor DarkYellow
    }

    Write-Host ""
}

New-NatsStream -StreamName "AUTH_EVENTS" -Subjects "auth.>"
New-NatsStream -StreamName "CATALOG_EVENTS" -Subjects "catalog.>"
New-NatsStream -StreamName "DDMRP_EVENTS" -Subjects "ddmrp.>"
New-NatsStream -StreamName "EXECUTION_EVENTS" -Subjects "execution.>"
New-NatsStream -StreamName "ANALYTICS_EVENTS" -Subjects "analytics.>"
New-NatsStream -StreamName "AI_AGENT_EVENTS" -Subjects "ai_agent.>"
New-NatsStream -StreamName "DLQ_EVENTS" -Subjects "dlq.>"

Write-Host "✅ All streams created successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Verify streams with: nats stream list --server=$NatsUrl" -ForegroundColor Cyan
