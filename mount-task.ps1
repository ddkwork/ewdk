$ErrorActionPreference = "Stop"

$TASK_NAME = "EWDK_Mount"
$MOUNT_LETTER = "E:"
$GITHUB_WORKSPACE = $env:GITHUB_WORKSPACE

if ($GITHUB_WORKSPACE) {
    $EWDK_ISO_PATH = "$env:TEMP\ewdk.iso"
    $IS_CI = $true
    Write-Host "=== CI Environment Detected ===" -ForegroundColor Cyan
    Write-Host "ISO path: $EWDK_ISO_PATH" -ForegroundColor Yellow
} else {
    $EWDK_ISO_PATH = "d:\ewdk\EWDK_br_release_28000_251103-1709.iso"
    $IS_CI = $false
    Write-Host "=== Local Environment ===" -ForegroundColor Cyan
    Write-Host "ISO path: $EWDK_ISO_PATH" -ForegroundColor Yellow
}

Write-Host ""

if (-not (Test-Path $EWDK_ISO_PATH)) {
    Write-Host "Error: ISO not found: $EWDK_ISO_PATH" -ForegroundColor Red
    exit 1
}

$existingTask = Get-ScheduledTask -TaskName $TASK_NAME -ErrorAction SilentlyContinue
if ($existingTask) {
    Write-Host "Removing existing task..." -ForegroundColor Yellow
    Unregister-ScheduledTask -TaskName $TASK_NAME -Confirm:$false
}

$trigger = New-ScheduledTaskTrigger -AtLogOn
$trigger.Delay = "PT10S"

$action = New-ScheduledTaskAction -Execute "powershell.exe" -Argument @"
-ExecutionPolicy Bypass -Command "
`$isoPath = '$EWDK_ISO_PATH'
`$mountLetter = '$MOUNT_LETTER'
`$alreadyMounted = `$false
try {
    `$vol = Get-Volume -DriveLetter `$mountLetter.Replace(':', '') -ErrorAction SilentlyContinue
    if (`$vol -and `$vol.DriveType -eq 'CD-ROM') { `$alreadyMounted = `$true }
} catch {}
if (-not `$alreadyMounted) {
    Mount-DiskImage -ImagePath `$isoPath -PassThru | Out-Null
    Write-Host 'EWDK mounted'
}
"@

$principal = New-ScheduledTaskPrincipal -UserId $env:USERNAME -LogonType Interactive -RunLevel Highest

Register-ScheduledTask -TaskName $TASK_NAME -Trigger $trigger -Action $action -Principal $principal -Description "Auto-mount EWDK ISO at logon" | Out-Null

Write-Host "Task registered: $TASK_NAME" -ForegroundColor Green
Write-Host "Triggers: At logon (delay 10s)" -ForegroundColor Yellow
Write-Host ""
Write-Host "To run now without waiting:" -ForegroundColor Cyan
Write-Host "  Start-ScheduledTask -TaskName '$TASK_NAME'" -ForegroundColor White
