# DevCLI v1.1.0 Installation & System Requirements

## System Requirements

**Operating Systems:**
* **Linux:** Ubuntu, Debian, Fedora, Arch Linux, openSUSE, Alpine, RHEL, Pop!_OS, Linux Mint, Void, etc. (All distributions and desktop environments supported)
* **Windows:** Windows 10, Windows 11 (or newer)
* **macOS:** macOS 11 (Big Sur) or newer

**Hardware Requirements:**
* **Processor:** Dual-core processor (Intel/AMD x64 or ARM64)
* **RAM:** 4GB minimum (8GB+ recommended for development workloads)
* **Storage:** 200MB free disk space for application components

**Software Prerequisites:**
* Go 1.21 or higher (auto-installed in user space by Linux installer if missing)
* Python 3.8+ (for Python venv & script runner)
* Node.js, Rust, or GCC/G++ (optional, for multi-language Web IDE execution)

---

## Installation Methods

### METHOD 1: Universal Single-Command Installation (Linux)

For all Linux distributions and desktop environments (GNOME, KDE, XFCE, Cinnamon, MATE, LXQt, etc.):

```bash
curl -fsSL https://raw.githubusercontent.com/phravins/devcli/main/install_linux.sh | bash
```

Or from local directory:

```bash
chmod +x install_linux.sh
./install_linux.sh
```

**What this installer does:**
1. Detects system architecture (`amd64`, `arm64`, etc.).
2. Checks for Go; if missing, **automatically installs Go 1.24.0 in user space (`~/.local/go`) without requiring sudo/root**.
3. Compiles the `devcli` binary to `~/.devcli/bin/devcli`.
4. Registers a **Linux Desktop Application Entry** (`~/.local/share/applications/devcli.desktop`) and installs icon assets to `~/.local/share/icons/` so DevCLI appears in your system Application Launcher.
5. Configures PATH automatically in `.bashrc`, `.zshrc`, `.profile`, and `config.fish`.

---

### METHOD 2: Automated Installation (Windows)

Open PowerShell as Administrator and run:

```powershell
irm https://raw.githubusercontent.com/phravins/devcli/main/install.ps1 | iex
```

Or use the setup batch file:
1. Download `setup_devcli.bat`.
2. Right-click and select **"Run as administrator"**.

---

### METHOD 3: Go Direct Install

If Go is installed on your system:

```bash
go run github.com/phravins/devcli@latest install
```

---

### METHOD 4: Building from Source

Clone the repository and compile manually:

```bash
git clone https://github.com/phravins/devcli.git
cd devcli
go build -o devcli main.go
./devcli install
```
