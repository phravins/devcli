#!/bin/bash

# DevCLI Automated Installer (macOS/Linux)

set -e

echo "==========================================="
echo "      DevCLI Automated Installer"
echo "==========================================="
echo ""

# Check for Go
echo "[INFO] Checking for Go installation..."
if ! command -v go &> /dev/null; then
    echo "[WARN] Go not found. Starting installation..."
    echo ""
    
    OS="$(uname -s)"
    ARCH="$(uname -m)"
    
    # Determine OS/Arch for download
    case "${OS}" in
        Linux*)     GOOS=linux;;
        Darwin*)    GOOS=darwin;;
        *)          GOOS="UNKNOWN:${OS}"
    esac
    
    case "${ARCH}" in
        x86_64)    GOARCH=amd64;;
        aarch64)   GOARCH=arm64;;
        arm64)     GOARCH=arm64;;
        *)         GOARCH="UNKNOWN:${ARCH}"
    esac

    if [[ "$GOOS" == *"UNKNOWN"* || "$GOARCH" == *"UNKNOWN"* ]]; then
        echo "[ERROR] Could not detect OS/Arch automatically ($GOOS / $GOARCH)."
        echo "[ERROR] Please install Go manually from https://go.dev/dl/"
        exit 1
    fi

    GO_VER="1.23.4"
    GO_TAR="go${GO_VER}.${GOOS}-${GOARCH}.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TAR}"
    
    echo "[INFO] Downloading Go ${GO_VER} for ${GOOS}/${GOARCH}..."
    if ! curl -L -o "/tmp/${GO_TAR}" "${GO_URL}"; then
        echo "[ERROR] Failed to download Go. Please check your internet connection."
        echo "[ERROR] You can manually download from: ${GO_URL}"
        exit 1
    fi
    
    echo "[INFO] Installing Go to /usr/local/go..."
    if [ "$EUID" -ne 0 ]; then
        echo "[INFO] Root privileges required to install Go to /usr/local/go."
        echo "[INFO] Please enter password if prompted."
        if ! sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf "/tmp/${GO_TAR}"; then
            echo "[ERROR] Failed to install Go. Please try installing manually."
            exit 1
        fi
    else
        if ! rm -rf /usr/local/go && tar -C /usr/local -xzf "/tmp/${GO_TAR}"; then
            echo "[ERROR] Failed to install Go. Please try installing manually."
            exit 1
        fi
    fi
    
    export PATH=$PATH:/usr/local/go/bin
    echo "[INFO] Go installed successfully."
else
    echo "[INFO] Go is already installed."
    echo "[INFO] $(go version)"
fi

echo ""
echo "[INFO] Verifying Go installation..."
if ! command -v go &> /dev/null; then
    echo "[ERROR] Go installation verification failed."
    echo "[ERROR] Please restart your terminal and try again, or install Go manually."
    exit 1
fi

echo "[INFO] Go is ready."
echo ""

# Install DevCLI
echo "[INFO] Installing DevCLI..."

if [ -f "go.mod" ]; then
    echo "[INFO] Found 'go.mod'. Installing from local source..."
    go install .
else
    echo "[INFO] Installing latest version from GitHub..."
    go install github.com/phravins/devcli@latest
fi

if [ $? -ne 0 ]; then
    echo "[ERROR] Failed to install DevCLI."
    echo "[ERROR] Please check your internet connection and Go installation."
    exit 1
fi

# Ensure GOPATH/bin is in PATH
GOPATH=$(go env GOPATH)
GOBIN="$GOPATH/bin"

echo ""
echo "[INFO] Verifying DevCLI installation..."

# Check if PATH already contains GOBIN
PATH_UPDATED=0
if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
    echo "[WARN] $GOBIN is not in your PATH."
    echo "[INFO] Adding to PATH configuration..."
    
    # Detect shell and appropriate RC file
    SHELL_RC=""
    SHELL_NAME=$(basename "$SHELL")
    
    case "$SHELL_NAME" in
        bash)
            if [ -f "$HOME/.bashrc" ]; then
                SHELL_RC="$HOME/.bashrc"
            elif [ -f "$HOME/.bash_profile" ]; then
                SHELL_RC="$HOME/.bash_profile"
            elif [ -f "$HOME/.profile" ]; then
                SHELL_RC="$HOME/.profile"
            fi
            ;;
        zsh)
            SHELL_RC="$HOME/.zshrc"
            ;;
        fish)
            SHELL_RC="$HOME/.config/fish/config.fish"
            ;;
        *)
            if [ -f "$HOME/.profile" ]; then
                SHELL_RC="$HOME/.profile"
            fi
            ;;
    esac
    
    if [ -n "$SHELL_RC" ]; then
        # Check if already in RC file
        if grep -q "GOPATH/bin" "$SHELL_RC" 2>/dev/null || grep -q "$GOBIN" "$SHELL_RC" 2>/dev/null; then
            echo "[INFO] PATH export already exists in $SHELL_RC"
        else
            # Add PATH export to RC file
            if [ "$SHELL_NAME" = "fish" ]; then
                echo "set -gx PATH \$PATH $GOBIN" >> "$SHELL_RC"
            else
                echo "" >> "$SHELL_RC"
                echo "# Add Go binaries to PATH (added by DevCLI installer)" >> "$SHELL_RC"
                echo "export PATH=\$PATH:$GOBIN" >> "$SHELL_RC"
            fi
            echo "[SUCCESS] Added PATH export to $SHELL_RC"
            PATH_UPDATED=1
        fi
    else
        echo "[WARN] Could not detect shell configuration file."
        echo "[WARN] Please manually add '$GOBIN' to your PATH."
    fi
    
    # Add to current session
    export PATH=$PATH:$GOBIN
else
    echo "[INFO] $GOBIN is already in your PATH."
fi

# Verify DevCLI
echo ""
echo "[INFO] Final verification..."
if command -v devcli &> /dev/null; then
    echo ""
    echo "[SUCCESS] DevCLI has been installed successfully!"
    echo ""
    devcli --version
    echo ""
    echo "[INFO] You can now use 'devcli' command in your terminal."
    
    if [ $PATH_UPDATED -eq 1 ]; then
        echo ""
        echo "[INFO] PATH has been updated in $SHELL_RC"
        echo "[INFO] To use DevCLI in this session, run:"
        echo "[INFO]   source $SHELL_RC"
        echo "[INFO] Or close and reopen your terminal."
    fi
else
    echo "[WARN] DevCLI installed but not immediately available."
    echo ""
    echo "[INFO] Please follow these steps:"
    echo "[INFO]   1. Close this terminal window"
    echo "[INFO]   2. Open a new terminal window"
    echo "[INFO]   3. Run 'devcli --version' to verify installation"
    echo ""
    if [ -n "$SHELL_RC" ]; then
        echo "[INFO] Or run: source $SHELL_RC"
    fi
    echo ""
    echo "[INFO] If you still have issues, ensure '$GOBIN' is in your PATH."
fi

echo ""
echo "==========================================="
echo "       Installation Complete"
echo "==========================================="
echo ""
echo "For help, run: devcli --help"
echo ""

