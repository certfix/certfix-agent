#!/bin/bash
set -e

# Service and installation paths
SERVICE_NAME="certfix-agent"
BIN_PATH="/usr/local/bin/certfix-agent"
CONFIG_DIR="/etc/certfix-agent"
CONFIG_FILE="$CONFIG_DIR/config.json"
SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"

echo "[INFO] Uninstalling Certfix Agent..."

# Check if running as root or with sudo
if [[ $EUID -ne 0 ]]; then
   echo "[ERROR] This script must be run as root or with sudo"
   exit 1
fi

# Stop and disable service
if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
    echo "[INFO] Stopping service..."
    systemctl stop "$SERVICE_NAME"
fi

if systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
    echo "[INFO] Disabling service..."
    systemctl disable "$SERVICE_NAME"
fi

# Remove service file
if [ -f "$SERVICE_FILE" ]; then
    echo "[INFO] Removing service file..."
    rm -f "$SERVICE_FILE"
    systemctl daemon-reload
fi

# Remove binary
if [ -f "$BIN_PATH" ]; then
    echo "[INFO] Removing binary..."
    rm -f "$BIN_PATH"
fi

# Ask about config removal
if [ -d "$CONFIG_DIR" ]; then
    read -p "[QUESTION] Remove configuration directory $CONFIG_DIR? (y/n): " remove_config
    if [[ "$remove_config" =~ ^[Yy]$ ]]; then
        echo "[INFO] Removing configuration directory..."
        rm -rf "$CONFIG_DIR"
    else
        echo "[INFO] Keeping configuration directory for future reinstalls"
    fi
fi

# Clean up any remaining systemd artifacts
systemctl reset-failed "$SERVICE_NAME" 2>/dev/null || true

echo "[SUCCESS] Certfix Agent has been completely removed!"
echo ""
echo "Removed items:"
echo "  - Service: $SERVICE_NAME"
echo "  - Binary: $BIN_PATH"
echo "  - Service file: $SERVICE_FILE"
if [[ "$remove_config" =~ ^[Yy]$ ]]; then
    echo "  - Configuration: $CONFIG_DIR"
fi
echo ""
echo "You can now safely reinstall if needed."