@echo off
for %%I in ("%~dp0.") do set "DRIVER_NAME=%%~nI"
echo ========================================
echo  %DRIVER_NAME% - Uninstall Driver
echo ========================================
echo.

echo [*] Stopping driver...
sc stop %DRIVER_NAME%
if errorlevel 1 (
    echo [WARN] Could not stop driver (may already be stopped).
)

echo [*] Deleting driver service...
sc delete %DRIVER_NAME%
if errorlevel 1 (
    echo [ERROR] Failed to delete service.
    pause
    exit /b 1
)

echo [OK] Driver uninstalled successfully.
pause
