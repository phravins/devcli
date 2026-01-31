#!/bin/bash

# DevCLI Automated Installer (macOS/Linux)

set -e

echo "==========================================="
echo "      DevCLI Automated Installer"
echo "==========================================="

# Check for Go
echo "[INFO] Checking for Go installation..."
if ! command -v go &> /dev/null; then
    echo "[WARN] Go not found. Starting installation..."
    
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
        echo "Please install Go manually from https://go.dev/dl/"
        exit 1
    fi

    GO_VER="1.23.4"
    GO_TAR="go${GO_VER}.${GOOS}-${GOARCH}.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TAR}"
    
    echo "[INFO] Downloading Go ${GO_VER} for ${GOOS}/${GOARCH}..."
    curl -L -o "/tmp/${GO_TAR}" "${GO_URL}"
    
    echo "[INFO] Installing Go to /usr/local/go..."
    if [ "$EUID" -ne 0 ]; then
        echo "Root privileges required to install Go to /usr/local/go. Please enter password if prompted."
        sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf "/tmp/${GO_TAR}"
    else
        rm -rf /usr/local/go && tar -C /usr/local -xzf "/tmp/${GO_TAR}"
    fi
    
    export PATH=$PATH:/usr/local/go/bin
    echo "[INFO] Go installed successfully."
else
    echo "[INFO] Go is already installed."
fi

# Install DevCLI
echo "[INFO] Installing DevCLI..."

if [ -f "go.mod" ]; then
    echo "[INFO] 'go.mod' found. Installing from local source..."
    go install .
else
    echo "[INFO] Installing latest version from GitHub..."
    go install github.com/phravins/devcli@latest
fi

# Ensure GOPATH/bin is in PATH
GOPATH=$(go env GOPATH)
GOBIN="$GOPATH/bin"

if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
    echo "[WARN] $GOBIN is not in your PATH."
    SHELL_RC=""
    if [ -f "$HOME/.zshrc" ]; then
        SHELL_RC="$HOME/.zshrc"
    elif [ -f "$HOME/.bashrc" ]; then
        SHELL_RC="$HOME/.bashrc"
    fi
    
    if [ -n "$SHELL_RC" ]; then
        echo "Would you like to add it to $SHELL_RC? [y/N]"
        read -r response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])+$ ]]; then
            echo "export PATH=\$PATH:$GOBIN" >> "$SHELL_RC"
            echo "[SUCCESS] Added to $SHELL_RC. Run 'source $SHELL_RC' to apply."
        else
            echo "Please manually add '$GOBIN' to your PATH."
        fi
    else
         echo "Please manually add '$GOBIN' to your PATH."
    fi
fi

# Verify
echo "[INFO] Verifying DevCLI installation..."
if command -v devcli &> /dev/null; then
    echo ""
    echo "[SUCCESS] DevCLI has been installed successfully!"
    echo ""
    devcli --version
else
    echo "[WARN] DevCLI installed but not found in PATH."
    echo "Make sure '$GOBIN' is in your PATH."
fi

echo ""
echo "==========================================="
echo "       Installation Complete"
echo "==========================================="
