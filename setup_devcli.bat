@echo off
setlocal EnableDelayedExpansion

echo ===========================================
echo       DevCLI Automated Installer
echo ===========================================
echo.

:: Check for Admin privileges (only needed for Go installation)
set "NEED_ADMIN=0"
go version >nul 2>&1
if %errorLevel% neq 0 (
    set "NEED_ADMIN=1"
)

if %NEED_ADMIN%==1 (
    net session >nul 2>&1
    if %errorLevel% neq 0 (
        echo [WARN] Go is not installed and Administrator privileges are required to install it.
        echo.
        echo Please choose one of the following options:
        echo   1. Right-click this script and select "Run as administrator"
        echo   2. Install Go manually from https://go.dev/dl/ and run this script again
        echo.
        pause
        exit /b 1
    )
)

:: Check if Go is installed
echo [INFO] Checking for Go installation...
go version >nul 2>&1
if %errorLevel% equ 0 (
    echo [INFO] Go is already installed.
    for /f "tokens=*" %%i in ('go version') do set "GO_VERSION=%%i"
    echo [INFO] !GO_VERSION!
) else (
    echo [WARN] Go not found. Starting installation...
    echo.
    
    :: Define Go version and URL
    set "GO_VER=1.24.0"
    set "GO_MSI=go%GO_VER%.windows-amd64.msi"
    set "GO_URL=https://go.dev/dl/%GO_MSI%"
    
    echo [INFO] Downloading Go %GO_VER%...
    curl -o "%TEMP%\%GO_MSI%" "%GO_URL%"
    if %errorLevel% neq 0 (
        echo [ERROR] Failed to download Go. Please check your internet connection.
        echo [ERROR] You can manually download from: %GO_URL%
        pause
        exit /b 1
    )
    
    echo [INFO] Installing Go... (This may take a minute)
    msiexec /i "%TEMP%\%GO_MSI%" /quiet /qn
    if %errorLevel% neq 0 (
        echo [ERROR] Failed to install Go.
        echo [ERROR] Please try installing manually from https://go.dev/dl/
        pause
        exit /b 1
    )
    
    echo [INFO] Go installed successfully.
    
    :: Add Go to PATH for the current session
    set "PATH=%PATH%;C:\Program Files\Go\bin"
)

:: Verify Go again
echo.
echo [INFO] Verifying Go installation...
go version >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] Go installation verification failed.
    echo [ERROR] Please restart your terminal and try again, or install Go manually.
    pause
    exit /b 1
)

echo [INFO] Go is ready.
echo.

:: Install DevCLI
echo [INFO] Installing DevCLI...

:: Check if running from within the source code
if exist "go.mod" (
    echo [INFO] Found 'go.mod'. Installing from local source...
    echo [EXEC] go run main.go install
    go run main.go install
) else (
    echo [INFO] Installing latest version from GitHub...
    echo [EXEC] go run github.com/phravins/devcli@latest install
    go run github.com/phravins/devcli@latest install
)

if %errorLevel% neq 0 (
    echo [ERROR] Failed to install DevCLI.
    echo [ERROR] Please check your internet connection and Go installation.
    pause
    exit /b 1
)

:: Add DevCLI bin to PATH for current session
if exist "%USERPROFILE%\.devcli\bin" (
    set "PATH=%PATH%;%USERPROFILE%\.devcli\bin"
)

echo.
echo [INFO] Verifying DevCLI installation...
devcli --version >nul 2>&1
if %errorLevel% neq 0 (
    echo [WARN] DevCLI installed but not found in current PATH.
    echo [INFO] Installing to PATH...
)

:: Permanently add DevCLI bin to user PATH using setx
echo [INFO] Adding DevCLI to your PATH permanently...
for /f "tokens=2*" %%a in ('reg query "HKCU\Environment" /v PATH 2^>nul') do set "USER_PATH=%%b"

:: Check if DevCLI bin is already in PATH
echo %USER_PATH% | findstr /C:"%USERPROFILE%\.devcli\bin" >nul
if %errorLevel% neq 0 (
    :: Not in PATH, add it
    if defined USER_PATH (
        setx PATH "%USER_PATH%;%USERPROFILE%\.devcli\bin" >nul
    ) else (
        setx PATH "%USERPROFILE%\.devcli\bin" >nul
    )
    if %errorLevel% equ 0 (
        echo [SUCCESS] Added %USERPROFILE%\.devcli\bin to your PATH.
    ) else (
        echo [WARN] Could not automatically add to PATH.
        echo [WARN] Please manually add %USERPROFILE%\.devcli\bin to your PATH.
    )
) else (
    echo [INFO] %USERPROFILE%\.devcli\bin is already in your PATH.
)

:: Update PATH for current session
set "PATH=%PATH%;%USERPROFILE%\.devcli\bin"

:: Verify DevCLI
echo.
echo [INFO] Final verification...
devcli --version >nul 2>&1
if %errorLevel% equ 0 (
    echo.
    echo [SUCCESS] DevCLI has been installed successfully!
    echo.
    for /f "tokens=*" %%i in ('devcli --version') do echo %%i
    echo.
    echo [INFO] You can now use 'devcli' command in your terminal.
    echo [INFO] If the command is not recognized, please:
    echo [INFO]   1. Close this terminal
    echo [INFO]   2. Open a new terminal window
    echo [INFO]   3. Run 'devcli --version' to verify
) else (
    echo [WARN] DevCLI installed but not immediately available.
    echo.
    echo [INFO] Please follow these steps:
    echo [INFO]   1. Close this terminal window
    echo [INFO]   2. Open a new terminal window
    echo [INFO]   3. Run 'devcli --version' to verify installation
    echo.
    echo [INFO] If you still have issues, ensure %USERPROFILE%\.devcli\bin is in your PATH.
)

echo.
echo ===========================================
echo       Installation Complete
echo ===========================================
echo.
echo For help, run: devcli --help
echo.
pause
