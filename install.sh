#!/bin/bash
set -e

# GitHub repository and release info
REPO_OWNER="certfix"
REPO_NAME="certfix-agent"
SERVICE_NAME="certfix-agent"

# Installation paths
BIN_PATH="/usr/local/bin/certfix-agent"
CONFIG_DIR="/etc/certfix-agent"
CONFIG_FILE="$CONFIG_DIR/config.json"

# Detect architecture
detect_arch() {
    local arch
    arch=$(uname -m)
    case $arch in
        x86_64)
            echo "amd64"
            ;;
        aarch64)
            echo "arm64"
            ;;
        armv7l)
            echo "armv7"
            ;;
        *)
            echo "[ERROR] Unsupported architecture: $arch"
            echo "Supported architectures: x86_64, aarch64, armv7l"
            exit 1
            ;;
    esac
}

ARCH=$(detect_arch)
BINARY_NAME="certfix-agent-linux-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/${BINARY_NAME}"

echo "[INFO] Installing Certfix Agent for $(uname -m) (${ARCH})..."

# Check if running as root or with sudo
if [[ $EUID -ne 0 ]]; then
   echo "[ERROR] This script must be run as root or with sudo"
   exit 1
fi

# Create config directory
mkdir -p "$CONFIG_DIR"

# Configuration setup (only on first install)
if [ ! -f "$CONFIG_FILE" ]; then
  echo "[INFO] First time setup - please provide configuration:"
  read -p "API Token: " token
  read -p "Server endpoint (e.g., https://api.example.com): " endpoint
  read -p "Enable auto-update? (y/n): " autoupdate

  if [[ "$autoupdate" =~ ^[Yy]$ ]]; then
    autoupdate=true
  else
    autoupdate=false
  fi

  # Create config file
  cat > "$CONFIG_FILE" <<EOF
{
  "token": "$token",
  "endpoint": "$endpoint",
  "auto_update": $autoupdate,
  "current_version": "0.0.1",
  "architecture": "$ARCH"
}
EOF
  chmod 600 "$CONFIG_FILE"
  echo "[INFO] Configuration saved to $CONFIG_FILE"
fi

# Download the latest binary
echo "[INFO] Downloading latest release for ${ARCH}..."
if ! curl -fsSL "$DOWNLOAD_URL" -o "$BIN_PATH"; then
  echo "[ERROR] Failed to download binary from $DOWNLOAD_URL"
  echo "[INFO] Make sure you have created a release with the binary attached"
  exit 1
fi

chmod +x "$BIN_PATH"
echo "[INFO] Binary installed to $BIN_PATH"

# Create systemd service
echo "[INFO] Creating systemd service..."
cat > "/etc/systemd/system/$SERVICE_NAME.service" <<EOF
[Unit]
Description=CertFix Agent Service
After=network.target

[Service]
ExecStart=$BIN_PATH
Restart=always
RestartSec=5
User=root
WorkingDirectory=/etc/certfix-agent

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
systemctl start "$SERVICE_NAME"

# Check service status
if systemctl is-active --quiet "$SERVICE_NAME"; then
  echo "[SUCCESS] Certfix Agent installed and running!"
  echo "Architecture: $(uname -m) (${ARCH})"
  echo "Service status: $(systemctl is-active $SERVICE_NAME)"
  echo "To check logs: journalctl -u $SERVICE_NAME -f"
else
  echo "[WARNING] Service installed but not running. Check logs with:"
  echo "journalctl -u $SERVICE_NAME"
fi