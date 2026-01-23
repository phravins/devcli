@echo off
setlocal EnableDelayedExpansion

echo ===========================================
echo       DevCLI Automated Installer
echo ===========================================

:: Check for Admin privileges
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] This script requires Administrator privileges.
    echo Please right-click and select "Run as administrator".
    pause
    exit /b 1
)

:: Check if Go is installed
echo [INFO] Checking for Go installation...
go version >nul 2>&1
if %errorLevel% equ 0 (
    echo [INFO] Go is already installed.
) else (
    echo [WARN] Go not found. Starting installation...
    
    :: Define Go version and URL
    set "GO_VER=1.23.4"
    set "GO_MSI=go%GO_VER%.windows-amd64.msi"
    set "GO_URL=https://go.dev/dl/%GO_MSI%"
    
    echo [INFO] Downloading Go %GO_VER%...
    curl -o "%TEMP%\%GO_MSI%" "%GO_URL%"
    if %errorLevel% neq 0 (
        echo [ERROR] Failed to download Go. Please check your internet connection.
        pause
        exit /b 1
    )
    
    echo [INFO] Installing Go... (This may take a minute)
    msiexec /i "%TEMP%\%GO_MSI%" /quiet
    if %errorLevel% neq 0 (
        echo [ERROR] Failed to install Go.
        pause
        exit /b 1
    )
    
    echo [INFO] Go installed successfully.
    
    :: Manually add Go to PATH for the current session
    set "PATH=%PATH%;C:\Program Files\Go\bin"
)

:: Verify Go again
go version >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] Go installation seems to have failed or PATH is not updated.
    echo Please restart the terminal and try again.
    pause
    exit /b 1
)

echo [INFO] Go is ready.

:: Install DevCLI
echo [INFO] Installing DevCLI...

:: Check if running from within the source code
if exist "go.mod" (
    echo [INFO] 'go.mod' found. Installing from local source...
    echo [EXEC] go install .
    go install .
) else (
    echo [INFO] Installing latest version from GitHub...
    echo [EXEC] go install github.com/phravins/devcli@latest
    go install github.com/phravins/devcli@latest
)

if %errorLevel% neq 0 (
    echo [ERROR] Failed to install DevCLI.
    pause
    exit /b 1
)

:: Ensure GOPATH/bin is in PATH for current session
if exist "%USERPROFILE%\go\bin" (
    set "PATH=%PATH%;%USERPROFILE%\go\bin"
)

:: Verify DevCLI
echo [INFO] Verifying DevCLI installation...
devcli --version >nul 2>&1
if %errorLevel% equ 0 (
    echo.
    echo [SUCCESS] DevCLI has been installed successfully!
    echo.
    devcli --version
) else (
    echo [WARN] DevCLI installed but not found in PATH.
    echo You may need to add %USERPROFILE%\go\bin to your PATH or restart your terminal.
)

echo.
echo ===========================================
echo       Installation Complete
echo ===========================================
pause
