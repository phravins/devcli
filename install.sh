#!/bin/bash

# DevCLI Installer for Linux and macOS

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Starting DevCLI installation...${NC}"

# Detect OS
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux)
        OS_TYPE="linux"
        ;;
    Darwin)
        OS_TYPE="darwin"
        ;;
    *)
        echo -e "${RED}Unsupported OS: $OS${NC}"
        exit 1
        ;;
esac

# Map architecture
case "$ARCH" in
    x86_64)
        ARCH_TYPE="amd64"
        ;;
    aarch64|arm64)
        ARCH_TYPE="arm64"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo -e "Detected OS: $OS_TYPE"
echo -e "Detected Arch: $ARCH_TYPE"

# Define installation paths
INSTALL_DIR="$HOME/.devcli/bin"
BINARY_NAME="devcli"
TARGET_PATH="$INSTALL_DIR/$BINARY_NAME"

# Create directory
mkdir -p "$INSTALL_DIR"

# Check if running from source directory (dev mode) or downloading
if [ -f "./devcli" ]; then
    echo -e "${BLUE}Installing locally built binary...${NC}"
    cp "./devcli" "$TARGET_PATH"
elif [ -f "./devcli-$OS_TYPE-$ARCH_TYPE" ]; then
    echo -e "${BLUE}Installing cross-compiled binary...${NC}"
    cp "./devcli-$OS_TYPE-$ARCH_TYPE" "$TARGET_PATH"
else
    # In the future, this could download from GitHub Releases
    # RELEASE_URL="https://github.com/phravins/devcli/releases/latest/download/devcli-$OS_TYPE-$ARCH_TYPE"
    echo -e "${RED}Error: Binary not found in current directory.${NC}"
    echo "Please build the project first or download the binary for your platform."
    exit 1
fi

chmod +x "$TARGET_PATH"
echo -e "${GREEN}Installed binary to $TARGET_PATH${NC}"

# Setup PATH
SHELL_CONFIG=""
case "$SHELL" in
    */zsh)
        SHELL_CONFIG="$HOME/.zshrc"
        ;;
    */bash)
        SHELL_CONFIG="$HOME/.bashrc"
        ;;
    */fish)
        SHELL_CONFIG="$HOME/.config/fish/config.fish"
        ;;
    *)
        echo -e "${BLUE}Could not detect shell configuration file. You may need to add the path manually.${NC}"
        ;;
esac

if [ -n "$SHELL_CONFIG" ]; then
    if [[ "$SHELL" == *"fish"* ]]; then
        # Fish syntax
        if ! grep -q "$INSTALL_DIR" "$SHELL_CONFIG"; then
            echo "set -gx PATH \$PATH $INSTALL_DIR" >> "$SHELL_CONFIG"
            echo -e "${GREEN}Added to PATH in $SHELL_CONFIG${NC}"
        else
            echo "PATH already configured in $SHELL_CONFIG"
        fi
    else
        # Bash/Zsh syntax
        if ! grep -q "$INSTALL_DIR" "$SHELL_CONFIG"; then
            echo "" >> "$SHELL_CONFIG"
            echo "# DevCLI" >> "$SHELL_CONFIG"
            echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$SHELL_CONFIG"
            echo -e "${GREEN}Added to PATH in $SHELL_CONFIG${NC}"
        else
            echo "PATH already configured in $SHELL_CONFIG"
        fi
    fi
fi

echo -e "\n${GREEN}Installation Complete!${NC}"
echo -e "Please restart your terminal or run: ${BLUE}source $SHELL_CONFIG${NC}"
