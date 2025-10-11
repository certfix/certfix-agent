# certfix-agent

## Development

### Docker Environment

```
make docker-up
```

```
make docker-build
```

```
make docker-run
```

```
make docker-down
```

### Dentro do Container

```
make docker-shell
```

## Local Development

```bash
# Build for current platform
make build

# Build for all supported architectures
make build-all

# Build specific architectures
make build-amd64
make build-arm64
make build-armv7

# Clean build directory
make clean
```

## Release

Create a new release by pushing a tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

This will automatically build and release binaries for:

- Linux x86_64 (amd64)
- Linux ARM64 (aarch64)
- Linux ARMv7

## Installation

### Automatic Installation (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/certfix/certfix-agent/main/scripts/install.sh | sudo bash
```

### Manual Installation

```bash
# Download the install script
curl -fsSL https://raw.githubusercontent.com/certfix/certfix-agent/main/scripts/install.sh -o install.sh
chmod +x install.sh
sudo ./install.sh
```

### Verify Installation

```bash
# Check service status
sudo systemctl status certfix-agent

# View logs
sudo journalctl -u certfix-agent -f
```

## Uninstallation

### Automatic Removal

```bash
curl -fsSL https://raw.githubusercontent.com/certfix/certfix-agent/main/scripts/uninstall.sh | sudo bash
```

### Manual Removal

```bash
# Download the uninstall script
curl -fsSL https://raw.githubusercontent.com/certfix/certfix-agent/main/scripts/uninstall.sh -o uninstall.sh
chmod +x uninstall.sh
sudo ./uninstall.sh
```

## Configuration

The agent configuration is stored in `/etc/certfix-agent/config.json`:

```json
{
  "token": "your-api-token",
  "endpoint": "https://api.example.com",
  "auto_update": true,
  "current_version": "0.0.1",
  "architecture": "amd64"
}
```

## Supported Architectures

- **Linux x86_64** (Intel/AMD 64-bit)
- **Linux ARM64** (aarch64)
- **Linux ARMv7** (32-bit ARM)

## Service Management

```bash
# Start service
sudo systemctl start certfix-agent

# Stop service
sudo systemctl stop certfix-agent

# Restart service
sudo systemctl restart certfix-agent

# Enable auto-start
sudo systemctl enable certfix-agent

# Disable auto-start
sudo systemctl disable certfix-agent

# View service status
sudo systemctl status certfix-agent

# View logs
sudo journalctl -u certfix-agent -f
```

## Manual remove

```
# Stop and disable service
sudo systemctl stop certfix-agent
sudo systemctl disable certfix-agent

# Remove service file
sudo rm -f /etc/systemd/system/certfix-agent.service
sudo systemctl daemon-reload

# Remove binary
sudo rm -f /usr/local/bin/certfix-agent

# Remove configuration (optional)
sudo rm -rf /etc/certfix-agent

# Reset any failed service states
sudo systemctl reset-failed certfix-agent
```

## Updates

### Automatic Update

```bash
curl -fsSL https://raw.githubusercontent.com/certfix/certfix-agent/main/scripts/update.sh | sudo bash
```

### Automatic Update (No Confirmation)

```bash
curl -fsSL https://raw.githubusercontent.com/certfix/certfix-agent/main/scripts/update.sh | sudo bash -s -- --yes
```

### Manual Update

```bash
# Download the update script
curl -fsSL https://raw.githubusercontent.com/certfix/certfix-agent/main/scripts/update.sh -o update.sh
chmod +x update.sh
sudo ./update.sh
```

The update script will:

- Check for newer versions
- Download the appropriate binary for your architecture
- Create a backup of the current version
- Update the service safely
- Rollback automatically if the update fails
