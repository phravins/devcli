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

For a complete "no-hassle" installation that sets up Go and DevCLI automatically:

1.  Download the [setup_devcli.bat](../setup_devcli.bat) script.
2.  Right-click the file and select **"Run as administrator"**.
3.  The script will:
    -   Check if Go is installed (and automatically download/install it if missing).
    -   Install the latest version of DevCLI.
    -   Configure your system PATH.

### METHOD 2: Automated Installation (Linux/macOS)

For Linux and macOS users, use the provided install script:

1.  Download the [install.sh](../install.sh) script.
2.  Run the script in your terminal:

```bash
chmod +x install.sh
./install.sh
```

This will automatically detect your OS/Arch, set up the binary, and configure your shell (Bash, Zsh, or Fish).

### METHOD 3: Single Command Installation (If Go is already installed)

Install DevCLI directly using the `go install` command:

```bash
go install github.com/phravins/devcli@latest
```

This will download, build, and install the DevCLI binary to your `$GOPATH/bin` directory. Ensure that `$GOPATH/bin` is in your system PATH.

### METHOD 4: Building from Source

Clone the repository and build manually:

```bash
git clone https://github.com/phravins/devcli.git
cd devcli
go build -o devcli.exe .
```
