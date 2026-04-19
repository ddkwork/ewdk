$ErrorActionPreference = "Stop"

reg delete "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v INCLUDE /f
reg delete "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v LIB /f

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

function Get-MountedDriveLetter {
    param([string]$ImagePath)
    $diskImg = Get-DiskImage -ImagePath $ImagePath -ErrorAction SilentlyContinue
    if ($diskImg -and $diskImg.Attached) {
        $vol = Get-Volume -DiskImage $diskImg -ErrorAction SilentlyContinue
        if ($vol) { return $vol.DriveLetter }
    }
    return $null
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
`$diskImg = Get-DiskImage -ImagePath `$isoPath -ErrorAction SilentlyContinue
if (-not (`$diskImg -and `$diskImg.Attached)) {
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


$mountedDrive = Get-MountedDriveLetter -ImagePath $EWDK_ISO_PATH
if ($mountedDrive) {
    Write-Host "EWDK is mounted at: ${mountedDrive}:" -ForegroundColor Green
    $wdkRoot = "${mountedDrive}:\Program Files\Windows Kits\10"
    $setupEnvCmd = "${mountedDrive}:\BuildEnv\SetupBuildEnv.cmd"
    setx /M WDKContentRoot $wdkRoot
    setx /M WDK_ROOT $wdkRoot
    setx /M EWDKSetupEnvCmd $setupEnvCmd
    Write-Host "setx /M WDKContentRoot `"$wdkRoot`"" -ForegroundColor Cyan
    Write-Host "setx /M WDK_ROOT `"$wdkRoot`"" -ForegroundColor Cyan
    Write-Host "setx /M EWDKSetupEnvCmd `"$setupEnvCmd`"" -ForegroundColor Cyan
} else {
    Write-Host "Warning: Could not detect mounted drive letter. setx commands skipped." -ForegroundColor Yellow
    Write-Host "Manually run after mounting: setx /M WDKContentRoot `"E:\Program Files\Windows Kits\10`"" -ForegroundColor Yellow
}
