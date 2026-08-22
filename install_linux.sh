#!/usr/bin/env bash

# ==============================================================================
# DevCLI Universal Linux Desktop Installer
# Works on all Linux distributions (Ubuntu, Debian, Fedora, Arch, Alpine, etc.)
# and all Linux Desktop Environments (GNOME, KDE, XFCE, Cinnamon, MATE, LXQt, etc.)
# ==============================================================================

set -e

# Color output helpers
BOLD='\033[1m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BOLD}${BLUE}=====================================================${NC}"
echo -e "${BOLD}${BLUE}        DevCLI Universal Linux Desktop Installer      ${NC}"
echo -e "${BOLD}${BLUE}=====================================================${NC}\n"

# 1. Architecture Detection
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)   GOARCH="amd64" ;;
    aarch64)  GOARCH="arm64" ;;
    arm64)    GOARCH="arm64" ;;
    armv7l)   GOARCH="armv6l" ;;
    *)
        echo -e "${RED}[ERROR] Unsupported architecture: ${ARCH}${NC}"
        exit 1
        ;;
esac

OS=$(uname -s)
if [ "$OS" != "Linux" ]; then
    echo -e "${RED}[ERROR] This installer is for Linux systems. Detected OS: ${OS}${NC}"
    exit 1
fi

echo -e "${BLUE}[INFO] System Architecture: ${BOLD}${ARCH} (${GOARCH})${NC}"

# 2. Check or Install Go Environment
GO_BIN=""
if command -v go &> /dev/null; then
    GO_BIN=$(command -v go)
    echo -e "${GREEN}[OK] Found Go compiler: $(${GO_BIN} version)${NC}"
elif [ -f "$HOME/.local/go/bin/go" ]; then
    GO_BIN="$HOME/.local/go/bin/go"
    export PATH="$HOME/.local/go/bin:$PATH"
    echo -e "${GREEN}[OK] Found user Go installation: $(${GO_BIN} version)${NC}"
elif [ -f "/usr/local/go/bin/go" ]; then
    GO_BIN="/usr/local/go/bin/go"
    export PATH="/usr/local/go/bin:$PATH"
    echo -e "${GREEN}[OK] Found system Go installation: $(${GO_BIN} version)${NC}"
else
    echo -e "${YELLOW}[WARN] Go compiler not found. Installing Go 1.24.0 in user space (~/.local/go)...${NC}"
    GO_VERSION="1.24.0"
    GO_TAR="go${GO_VERSION}.linux-${GOARCH}.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TAR}"
    TMP_DIR=$(mktemp -d)

    echo -e "${BLUE}[INFO] Downloading ${GO_URL}...${NC}"
    if command -v curl &> /dev/null; then
        curl -fsSL -o "${TMP_DIR}/${GO_TAR}" "${GO_URL}"
    elif command -v wget &> /dev/null; then
        wget -q -O "${TMP_DIR}/${GO_TAR}" "${GO_URL}"
    else
        echo -e "${RED}[ERROR] Neither curl nor wget was found. Please install curl or wget.${NC}"
        exit 1
    fi

    mkdir -p "$HOME/.local"
    rm -rf "$HOME/.local/go"
    tar -C "$HOME/.local" -xzf "${TMP_DIR}/${GO_TAR}"
    rm -rf "${TMP_DIR}"

    GO_BIN="$HOME/.local/go/bin/go"
    export PATH="$HOME/.local/go/bin:$PATH"

    if ! [ -x "$GO_BIN" ]; then
        echo -e "${RED}[ERROR] Go installation failed.${NC}"
        exit 1
    fi
    echo -e "${GREEN}[OK] Successfully installed Go in user space!${NC}"
fi

# 3. Build & Deploy DevCLI
INSTALL_DIR="$HOME/.devcli/bin"
mkdir -p "$INSTALL_DIR"

WORK_DIR=$(pwd)
BUILD_TMP=""

if [ -f "main.go" ] && [ -f "go.mod" ]; then
    echo -e "${BLUE}[INFO] Building DevCLI from current directory...${NC}"
    "$GO_BIN" build -o "${INSTALL_DIR}/devcli" main.go
else
    echo -e "${BLUE}[INFO] Cloning latest DevCLI from GitHub repository...${NC}"
    BUILD_TMP=$(mktemp -d)
    git clone --depth 1 https://github.com/phravins/devcli.git "${BUILD_TMP}/devcli"
    cd "${BUILD_TMP}/devcli"
    "$GO_BIN" build -o "${INSTALL_DIR}/devcli" main.go
    cd "$WORK_DIR"
fi

chmod +x "${INSTALL_DIR}/devcli"
echo -e "${GREEN}[OK] DevCLI binary compiled to ${INSTALL_DIR}/devcli${NC}"

# 4. Run CLI Installation & Linux Desktop Integration
"${INSTALL_DIR}/devcli" install

# Cleanup temporary build directory if created
if [ -n "$BUILD_TMP" ] && [ -d "$BUILD_TMP" ]; then
    rm -rf "$BUILD_TMP"
fi

# 5. Shell Configuration (PATH export)
ADD_PATH_LINE="export PATH=\"\$PATH:${INSTALL_DIR}\""

configure_shell_rc() {
    local rc_file="$1"
    if [ -f "$rc_file" ] || [ -f "$(dirname "$rc_file")" ]; then
        if ! grep -q "${INSTALL_DIR}" "$rc_file" 2>/dev/null; then
            mkdir -p "$(dirname "$rc_file")"
            echo "" >> "$rc_file"
            echo "# DevCLI Workspace Path" >> "$rc_file"
            if [[ "$rc_file" == *"fish"* ]]; then
                echo "set -gx PATH \$PATH ${INSTALL_DIR}" >> "$rc_file"
            else
                echo "$ADD_PATH_LINE" >> "$rc_file"
            fi
            echo -e "${GREEN}[OK] Configured PATH in ${rc_file}${NC}"
        else
            echo -e "${BLUE}[INFO] PATH already present in ${rc_file}${NC}"
        fi
    fi
}

configure_shell_rc "$HOME/.bashrc"
configure_shell_rc "$HOME/.zshrc"
configure_shell_rc "$HOME/.profile"
configure_shell_rc "$HOME/.bash_profile"
configure_shell_rc "$HOME/.config/fish/config.fish"

# Also include ~/.local/go/bin if installed in user space
if [ -d "$HOME/.local/go/bin" ]; then
    GO_PATH_LINE="export PATH=\"\$PATH:$HOME/.local/go/bin\""
    for rc in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
        if [ -f "$rc" ] && ! grep -q "\.local/go/bin" "$rc"; then
            echo "$GO_PATH_LINE" >> "$rc"
        fi
    done
fi

echo -e "\n${BOLD}${GREEN}=====================================================${NC}"
echo -e "${BOLD}${GREEN}        DevCLI Linux Installation Complete!          ${NC}"
echo -e "${BOLD}${GREEN}=====================================================${NC}"
echo -e "  • Launch from Terminal : ${BOLD}devcli${NC}"
echo -e "  • Launch from App Menu : ${BOLD}DevCLI${NC} (Application Launcher)"
echo -e "  • Installed Binary     : ${INSTALL_DIR}/devcli"
echo -e "  • Desktop Entry        : $HOME/.local/share/applications/devcli.desktop"
echo -e "\n${YELLOW}Note: If 'devcli' command is not recognized immediately in your active shell, run:${NC}"
echo -e "  ${BOLD}source ~/.bashrc${NC}  (or restart your terminal window)\n"
