# DevCLI Windows Automated Installer via PowerShell
# Usage: irm https://raw.githubusercontent.com/phravins/devcli/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

function Show-RocketAnimation {
    param(
        [Parameter(Mandatory=$true)]
        [System.Management.Automation.Job]$Job,
        [Parameter(Mandatory=$true)]
        [string]$Message
    )

    $rocket = [char]::ConvertFromUtf32(128640)
    $rocketFrames = @(
        "$rocket.     ",
        " $rocket.    ",
        "  $rocket.   ",
        "   $rocket.  ",
        "    $rocket. ",
        "     $rocket.",
        "    .$rocket ",
        "   .$rocket  ",
        "  .$rocket   ",
        " .$rocket    "
    )
    $i = 0
    [Console]::CursorVisible = $false
    while ($Job.State -eq "Running") {
        $frame = $rocketFrames[$i % $rocketFrames.Length]
        Write-Host "`r$Message $frame" -NoNewline -ForegroundColor Cyan
        Start-Sleep -Milliseconds 150
        $i++
    }
    [Console]::CursorVisible = $true
    # Clear line
    Write-Host "`r$($Message.PadRight(80))`r" -NoNewline
}

function Install-DevCLI {
    $rocket = [char]::ConvertFromUtf32(128640)
    Write-Host ""
    Write-Host "===========================================" -ForegroundColor Magenta
    Write-Host "      $rocket DevCLI Automated Installer $rocket     " -ForegroundColor Cyan
    Write-Host "===========================================" -ForegroundColor Magenta
    Write-Host ""

    # 1. Check for Go Support
    $goInstalled = $false
    try {
        if (Get-Command "go" -ErrorAction SilentlyContinue) {
            $goInstalled = $true
        }
    } catch {}

    if (-not $goInstalled) {
        Write-Host "[WARN] Go is not installed. Setting up Go 1.23.4 for you..." -ForegroundColor Yellow
        $goMsiUrl = "https://go.dev/dl/go1.23.4.windows-amd64.msi"
        $msiPath = "$env:TEMP\go_installer.msi"
        
        $downloadJob = Start-Job -ScriptBlock {
            param($url, $path)
            Invoke-WebRequest -Uri $url -OutFile $path -UseBasicParsing
        } -ArgumentList $goMsiUrl, $msiPath
        
        Show-RocketAnimation -Job $downloadJob -Message "Downloading Go MSI..."
        Receive-Job -Job $downloadJob | Out-Null
        
        Write-Host "[INFO] Installing Go (A UAC prompt may appear to grant administrative permissions)..." -ForegroundColor Cyan
        
        try {
            $installProc = Start-Process -FilePath "msiexec.exe" -ArgumentList "/i `"$msiPath`" /quiet /qn" -Wait -PassThru -Verb RunAs
            if ($installProc.ExitCode -ne 0) {
                Write-Host "[ERROR] Failed to install Go automatically. Exit code: $($installProc.ExitCode)" -ForegroundColor Red
                Write-Host "Please install Go manually from https://go.dev/dl/ and re-run this script." -ForegroundColor Red
                return
            }
        } catch {
            Write-Host "[ERROR] Failed to install Go. You may need to run PowerShell as Administrator." -ForegroundColor Red
            return
        }
        
        $env:PATH += ";C:\Program Files\Go\bin"
        Write-Host "[SUCCESS] Go installed successfully." -ForegroundColor Green
    } else {
        $goVer = go version
        Write-Host "[SUCCESS] Found Go: $goVer" -ForegroundColor Green
    }

    Write-Host ""
    Write-Host "$rocket Preparing to install DevCLI..." -ForegroundColor Cyan

    $installCmd = {
        $ErrorActionPreference = "Continue"
        if (Test-Path "go.mod") {
            # Attempt local install if we are in the repo root
            $content = Get-Content "go.mod" -Raw
            if ($content -match "module github.com/phravins/devcli") {
                go run main.go install 2>&1 | Out-String
                return
            }
        }
        # Default to remote install
        go run github.com/phravins/devcli@latest install 2>&1 | Out-String
    }

    $devcliJob = Start-Job -ScriptBlock $installCmd
    
    Show-RocketAnimation -Job $devcliJob -Message "Compiling and Installing DevCLI..."
    
    $result = Receive-Job -Job $devcliJob
    
    # If the output indicates success or doesn't have an error
    $failed = $false
    if (($result -match "compile failed") -or ($result -match "build failed")) {
        $failed = $true
    }

    if (-not $failed) {
        Write-Host "[SUCCESS] DevCLI installed successfully! $rocket" -ForegroundColor Green
        
        # Check PATH configuration
        $devcliBinPath = "$env:USERPROFILE\.devcli\bin"
        $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        
        if ($userPath -notmatch [regex]::Escape($devcliBinPath)) {
            $newPath = $devcliBinPath
            if (-not [string]::IsNullOrEmpty($userPath)) {
                $newPath = "$userPath;$devcliBinPath"
            }
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
            Write-Host "[SUCCESS] Added $devcliBinPath to User PATH." -ForegroundColor Green
        } else {
            Write-Host "[SUCCESS] DevCLI is already in your PATH." -ForegroundColor Green
        }
        
        Write-Host "`n[INFO] Installation Complete!" -ForegroundColor Magenta
        Write-Host "[WARN] To apply the new PATH, please restart your terminal or open a new tab." -ForegroundColor Yellow
        Write-Host "       Then verify by running: devcli --help" -ForegroundColor Cyan
    } else {
        Write-Host "[ERROR] DevCLI installation finished with some errors." -ForegroundColor Red
        Write-Host "Logs:" -ForegroundColor Yellow
        Write-Host $result
    }
}

try {
    Install-DevCLI
} catch {
    Write-Host "`n[ERROR] An unexpected error occurred: $_" -ForegroundColor Red
} finally {
    [Console]::CursorVisible = $true
}
