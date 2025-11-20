# 📋 certfix-agent Testing Guide

Based on the code analysis, here's how to test the agent from configuration through to running:

## **Architecture Overview**

The agent is a simple heartbeat service that:

- Runs as a systemd service on Linux
- Logs a heartbeat every 10 seconds
- Supports multi-architecture builds (amd64, arm64, armv7)

---

## 🧪 Testing Methods

### **1. Local Development Testing (Fastest)**

#### Build and run locally:

```bash
cd /Users/guilhermecardoso/Development/UFSCar/certfix/certfix-agent

# Build for your native platform
make build-dev

# Run the agent
./build/certfix-agent-dev
```

**Expected output:**

```
[certfix-agent] Starting agent...
[certfix-agent] Heartbeat: agent is alive
[certfix-agent] Heartbeat: agent is alive
...
```

---

### **2. Docker-based Testing (Isolated Environment)**

#### Start the development environment:

```bash
cd /Users/guilhermecardoso/Development/UFSCar/certfix/certfix-agent

# Start Docker container
make docker-up

# Verify container is running
docker ps | grep certfix-agent-dev

# Enter the container
make docker-shell
```

#### Inside the container, test builds:

```bash
# Build for development
make build-dev

# Run the agent
./build/certfix-agent-dev

# Build for all architectures
make build-all

# Check the builds
ls -lh build/
```

#### From your host machine:

```bash
# Build and run in one step
make docker-run

# View logs
make docker-logs

# Run tests
make docker-test

# Stop the environment
make docker-down
```

---

### **3. Production-like Testing (systemd service)**

This simulates the actual installation:

#### Create a test configuration:

```bash
sudo mkdir -p /etc/certfix-agent

sudo tee /etc/certfix-agent/config.json > /dev/null <<EOF
{
  "token": "test-token-123",
  "endpoint": "https://api.example.com",
  "current_version": "0.0.1",
  "architecture": "amd64"
}
EOF

sudo chmod 600 /etc/certfix-agent/config.json
```

#### Install the binary:

```bash
# Build for Linux
cd /Users/guilhermecardoso/Development/UFSCar/certfix/certfix-agent
make build

# Copy to system location
sudo cp build/certfix-agent /usr/local/bin/certfix-agent
sudo chmod +x /usr/local/bin/certfix-agent
```

#### Create systemd service:

```bash
sudo tee /etc/systemd/system/certfix-agent.service > /dev/null <<'EOF'
[Unit]
Description=CertFix Agent Service
After=network.target

[Service]
ExecStart=/usr/local/bin/certfix-agent
Restart=always
RestartSec=5
User=root
WorkingDirectory=/etc/certfix-agent

[Install]
WantedBy=multi-user.target
EOF
```

#### Test the service:

```bash
# Reload systemd
sudo systemctl daemon-reload

# Start the service
sudo systemctl start certfix-agent

# Check status
sudo systemctl status certfix-agent

# View logs in real-time
sudo journalctl -u certfix-agent -f

# Test restart
sudo systemctl restart certfix-agent

# Stop the service
sudo systemctl stop certfix-agent
```

---

### **4. Installation Script Testing**

⚠️ **Note:** The install script expects GitHub releases. For local testing, you'll need to:

1. Create a local release or modify the script to use local files
2. Or test with an actual GitHub release

#### Manual installation test:

```bash
# Make the install script executable
chmod +x /Users/guilhermecardoso/Development/UFSCar/certfix/certfix-agent/scripts/install.sh

# Review the script before running
cat scripts/install.sh

# Run installation (requires GitHub release)
sudo ./scripts/install.sh
```

---

### **5. Update Script Testing**

#### Test update flow:

```bash
# Ensure you have an existing installation first
sudo systemctl status certfix-agent

# Test update interactively
sudo ./scripts/update.sh

# Test automatic update (no prompts)
sudo ./scripts/update.sh --yes

# Check version after update
cat /etc/certfix-agent/config.json | grep current_version
```

---

### **6. Multi-Architecture Build Testing**

```bash
cd /Users/guilhermecardoso/Development/UFSCar/certfix/certfix-agent

# Build for all architectures
make build-all

# Verify all binaries were created
ls -lh build/

# Expected output:
# certfix-agent-linux-amd64
# certfix-agent-linux-arm64
# certfix-agent-linux-armv7

# Check binary sizes (should be small, ~2-5 MB each)
file build/certfix-agent-linux-*
```

---

## 🔍 Testing Checklist

### **Configuration Testing:**

- ✅ Config file created at `/etc/certfix-agent/config.json`
- ✅ Correct permissions (600)
- ✅ Valid JSON format
- ✅ All required fields present

### **Build Testing:**

- ✅ Development build works (`make build-dev`)
- ✅ Production build works (`make build`)
- ✅ All architecture builds succeed (`make build-all`)
- ✅ Binaries are executable and stripped

### **Runtime Testing:**

- ✅ Agent starts successfully
- ✅ Heartbeat logs appear every 10 seconds
- ✅ Agent runs continuously without crashes
- ✅ Graceful shutdown on interrupt (Ctrl+C)

### **Service Testing:**

- ✅ systemd service starts
- ✅ Service auto-restarts on failure
- ✅ Logs appear in journalctl
- ✅ Service enabled at boot
- ✅ Service stops cleanly

### **Update Testing:**

- ✅ Current version detection works
- ✅ Latest version fetched from GitHub
- ✅ Binary download succeeds
- ✅ Backup created before update
- ✅ Service restarts after update
- ✅ Rollback works on failure

---

## 🐛 Common Issues & Solutions

| Issue                           | Solution                                                  |
| ------------------------------- | --------------------------------------------------------- |
| **"Failed to download binary"** | Ensure you have a GitHub release with the binary attached |
| **Service won't start**         | Check `journalctl -u certfix-agent` for errors            |
| **Permission denied**           | Run with `sudo` or as root                                |
| **Architecture mismatch**       | Verify with `uname -m` and check binary name              |
| **Config not found**            | Ensure `/etc/certfix-agent/config.json` exists            |

---

## 📝 Quick Test Script

Here's a complete test you can run:

```bash
#!/bin/bash
cd /Users/guilhermecardoso/Development/UFSCar/certfix/certfix-agent

echo "=== Testing certfix-agent ==="
echo ""

echo "1. Building development version..."
make build-dev
echo ""

echo "2. Running agent for 30 seconds..."
timeout 30s ./build/certfix-agent-dev &
AGENT_PID=$!
sleep 31
echo ""

echo "3. Testing multi-arch builds..."
make build-all
ls -lh build/
echo ""

echo "4. Testing Docker environment..."
make docker-up
sleep 3
make docker-build-dev
make docker-down
echo ""

echo "=== All tests completed! ==="
```

---

## 📚 Additional Resources

### Code Structure

- `cmd/agent.go` - Main agent code (simple heartbeat loop)
- `scripts/install.sh` - Installation script with architecture detection
- `scripts/update.sh` - Update script with version comparison and rollback
- `scripts/uninstall.sh` - Clean uninstallation
- `Makefile` - Build targets for all platforms
- `docker/` - Docker development environment

### Key Features

- **Multi-architecture support**: Automatic detection and build for amd64, arm64, armv7
- **Systemd integration**: Runs as a system service with auto-restart
- **Rollback capability**: Automatic rollback on failed updates
- **Configuration management**: JSON-based configuration in `/etc/certfix-agent/`

The agent is currently very simple (just a heartbeat logger), so testing is straightforward. The main complexity is in the installation, update, and service management scripts.
