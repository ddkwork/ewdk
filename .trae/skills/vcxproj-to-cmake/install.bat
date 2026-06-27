@echo off
for %%I in ("%~dp0.") do set "DRIVER_NAME=%%~nI"
echo ========================================
echo  %DRIVER_NAME% - Install Driver
echo ========================================
echo.

set DRIVER_PATH=

for /r "%~dp0Debug" %%F in (*.sys) do (
    if defined DRIVER_PATH (
        echo [WARN] Multiple .sys files found:
        echo       %DRIVER_PATH%
        echo       %%F
        echo.
        echo Please specify which driver to install by editing this script.
        pause
        exit /b 1
    )
    set "DRIVER_PATH=%%F"
)

if not defined DRIVER_PATH (
    for /r "%~dp0Release" %%F in (*.sys) do (
        if defined DRIVER_PATH (
            echo [WARN] Multiple .sys files found:
            echo       %DRIVER_PATH%
            echo       %%F
            echo.
            echo Please specify which driver to install by editing this script.
            pause
            exit /b 1
        )
        set "DRIVER_PATH=%%F"
    )
)

if not defined DRIVER_PATH (
    echo [ERROR] No .sys file found in Release or Debug directories.
    echo Please build the project first.
    pause
    exit /b 1
)

sc query %DRIVER_NAME% >nul 2>&1
if errorlevel 1 (
    echo [*] Creating driver service...
    sc create %DRIVER_NAME% type= kernel binPath= "%DRIVER_PATH%"
    if errorlevel 1 (
        echo [ERROR] Failed to create service.
        pause
        exit /b 1
    )
    echo [OK] Service created.
) else (
    echo [SKIP] Service already exists.
)

sc query %DRIVER_NAME% | find "RUNNING" >nul
if not errorlevel 1 (
    echo [SKIP] Driver is already running.
    pause
    exit /b 0
)

echo [*] Starting driver...
sc start %DRIVER_NAME%
if errorlevel 1 (
    echo [ERROR] Failed to start driver.
) else (
    echo [OK] Driver started.
)
pause