#!/bin/bash
set -e

REPO_URL="https://githubusercontent.com/certfix/certfix-agent/main"
BIN_PATH="/usr/local/bin/certfix-agent"
CONFIG_DIR="/etc/certfix-agent"
CONFIG_FILE="$CONFIG_DIR/config.json"
SERVICE_NAME="certfix-agent"

echo "[INFO] Instalando Certfix Agent..."

sudo mkdir -p "$CONFIG_DIR"

if [ ! -f "$CONFIG_FILE" ]; then
  echo "[INFO] Primeira configuração:"
  read -p "Token da API: " token
  read -p "Endpoint do servidor (ex: https://api.exemplo.com): " endpoint
  read -p "Deseja auto-update? (y/n): " autoupdate

  if [[ "$autoupdate" == "y" ]]; then
    autoupdate=true
  else
    autoupdate=false
  fi

  echo "{\"token\": \"$token\", \"endpoint\": \"$endpoint\", \"auto_update\": $autoupdate, \"current_version\": \"0.0.1\"}" | sudo tee "$CONFIG_FILE" >/dev/null
fi

sudo curl -fsSL "$REPO_URL/build/certfix-agent" -o "$BIN_PATH"
sudo chmod +x "$BIN_PATH"

sudo bash -c "cat > /etc/systemd/system/$SERVICE_NAME.service" <<EOF
[Unit]
Description=CertFix Agent Service
After=network.target

[Service]
ExecStart=$BIN_PATH
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable $SERVICE_NAME
sudo systemctl start $SERVICE_NAME

echo "[OK] Instalação concluída!"
