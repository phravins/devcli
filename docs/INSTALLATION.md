# Installation & Requirements

## System Requirements

**Operating System:**
*   **Windows:** Windows 10, Windows 11 (or newer)
*   **Linux:** Ubuntu, Debian, Fedora, Arch Linux (or any modern distribution)
*   **macOS:** macOS 11 (Big Sur) or newer

**Hardware:**
*   **Processor:** Dual-core processor (Intel/AMD x64 or ARM64)
*   **RAM:** 4GB minimum (8GB+ recommended for development workloads)
*   **Storage:** 200MB free disk space for application components

**Software:**
*   Go 1.21 or higher (for building from source)
*   Python 3.8 or newer (required for virtual environment features)
*   Node.js (optional, needed for JavaScript project templates)
*   Terminal with Unicode support (for proper rendering)

## Installation Methods

### METHOD 1: Automated Installation (Windows)

For a complete "no-hassle" installation that sets up Go and DevCLI automatically, open PowerShell and run:

```powershell
irm https://raw.githubusercontent.com/phravins/devcli/main/install.ps1 | iex
```

Alternatively, you can manually use the setup script:
1.  Download the [setup_devcli.bat](../setup_devcli.bat) script.
2.  Right-click the file and select **"Run as administrator"**.
3.  The script will check for Go, install the latest DevCLI, and configure your PATH.

### METHOD 2: Automated Universal Linux Installation (All Linux Desktops)

For all Linux distributions (Ubuntu, Debian, Fedora, Arch, Alpine, openSUSE, RHEL, Pop!_OS, etc.) and desktop environments (GNOME, KDE Plasma, XFCE, Cinnamon, MATE, LXQt, etc.), run:

```bash
chmod +x install_linux.sh
./install_linux.sh
```

Or run via single-command curl:

```bash
curl -fsSL https://raw.githubusercontent.com/phravins/devcli/main/install_linux.sh | bash
```

This installer:
* Automatically detects your architecture (`amd64`, `arm64`, etc.).
* Checks for Go; if missing, installs Go automatically in user space (`~/.local/go`) without needing root/sudo permissions.
* Compiles and installs the `devcli` binary to `~/.devcli/bin/devcli`.
* Creates a Linux Application Launcher (`.desktop` file) in `~/.local/share/applications/` so DevCLI appears in your Desktop App Launcher / Main Menu with icon support.
* Configures PATH automatically across Bash, Zsh, Fish, and POSIX profile environments.

### METHOD 3: Single Command Installation (If Go is already installed)

Install DevCLI directly using the `go run` command:

```bash
go run github.com/phravins/devcli@latest install
```

This will download the latest version, run the interactive installer, install the DevCLI binary to your `~/.devcli/bin` directory, and automatically configure your system PATH.

### METHOD 4: Building from Source

Clone the repository and build manually:

```bash
git clone https://github.com/phravins/devcli.git
cd devcli
go build -o devcli.exe .
```
