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
BACKUP_DIR="/tmp/certfix-agent-backup"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Function to get current version
get_current_version() {
    if [ -f "$CONFIG_FILE" ]; then
        # Try to extract version from config file
        local version=$(grep -o '"current_version"[[:space:]]*:[[:space:]]*"[^"]*"' "$CONFIG_FILE" 2>/dev/null | cut -d'"' -f4)
        if [ -n "$version" ] && [ "$version" != "0.0.1" ]; then
            echo "$version"
            return
        fi
    fi
    
    # Try to get version from binary (if it supports --version)
    if [ -f "$BIN_PATH" ]; then
        local version=$("$BIN_PATH" --version 2>/dev/null | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+' | head -1)
        if [ -n "$version" ]; then
            echo "$version"
            return
        fi
    fi
    
    echo "unknown"
}

# Function to get latest version from GitHub
get_latest_version() {
    local latest=$(curl -fsSL "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
    if [ -z "$latest" ]; then
        print_error "Failed to fetch latest version from GitHub"
        exit 1
    fi
    echo "$latest"
}

# Function to detect architecture
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
            print_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
}

# Function to compare versions
version_greater_than() {
    local version1="$1"
    local version2="$2"
    
    # Remove 'v' prefix if present
    version1=${version1#v}
    version2=${version2#v}
    
    # Compare versions using sort
    [ "$(printf '%s\n' "$version1" "$version2" | sort -V | head -n1)" != "$version1" ]
}

# Function to backup current binary
backup_binary() {
    if [ -f "$BIN_PATH" ]; then
        print_info "Creating backup of current binary..."
        mkdir -p "$BACKUP_DIR"
        cp "$BIN_PATH" "$BACKUP_DIR/certfix-agent-backup-$(date +%Y%m%d-%H%M%S)"
        print_success "Backup created in $BACKUP_DIR"
    fi
}

# Function to update configuration with new version
update_config_version() {
    local new_version="$1"
    if [ -f "$CONFIG_FILE" ]; then
        # Create a temporary file with updated version
        local temp_file=$(mktemp)
        if command -v jq >/dev/null 2>&1; then
            # Use jq if available
            jq --arg version "$new_version" '.current_version = $version' "$CONFIG_FILE" > "$temp_file"
        else
            # Fallback to sed
            sed "s/\"current_version\"[[:space:]]*:[[:space:]]*\"[^\"]*\"/\"current_version\": \"$new_version\"/" "$CONFIG_FILE" > "$temp_file"
        fi
        mv "$temp_file" "$CONFIG_FILE"
        chmod 600 "$CONFIG_FILE"
    fi
}

# Function to rollback on failure
rollback() {
    print_warning "Update failed. Attempting rollback..."
    local backup_file=$(ls -t "$BACKUP_DIR"/certfix-agent-backup-* 2>/dev/null | head -1)
    if [ -n "$backup_file" ] && [ -f "$backup_file" ]; then
        cp "$backup_file" "$BIN_PATH"
        chmod +x "$BIN_PATH"
        systemctl restart "$SERVICE_NAME" 2>/dev/null || true
        print_success "Rollback completed"
    else
        print_error "No backup found for rollback"
    fi
}

# Main update function
main() {
    print_info "Certfix Agent Update Script"
    echo "================================"
    
    # Check if running as root
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root or with sudo"
        exit 1
    fi
    
    # Check if agent is installed
    if [ ! -f "$BIN_PATH" ]; then
        print_error "Certfix Agent is not installed. Please install it first."
        exit 1
    fi
    
    # Get current and latest versions
    print_info "Checking current version..."
    CURRENT_VERSION=$(get_current_version)
    print_info "Current version: $CURRENT_VERSION"
    
    print_info "Fetching latest version from GitHub..."
    LATEST_VERSION=$(get_latest_version)
    print_info "Latest version: $LATEST_VERSION"
    
    # Compare versions
    if [ "$CURRENT_VERSION" = "$LATEST_VERSION" ]; then
        print_success "Already running the latest version ($CURRENT_VERSION)"
        exit 0
    fi
    
    if [ "$CURRENT_VERSION" != "unknown" ] && ! version_greater_than "$LATEST_VERSION" "$CURRENT_VERSION"; then
        print_success "Current version ($CURRENT_VERSION) is newer than or equal to latest release ($LATEST_VERSION)"
        exit 0
    fi
    
    print_info "Update available: $CURRENT_VERSION → $LATEST_VERSION"
    
    # Confirm update
    if [ "${1:-}" != "--yes" ] && [ "${1:-}" != "-y" ]; then
        read -p "Do you want to proceed with the update? (y/N): " confirm
        if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
            print_info "Update cancelled by user"
            exit 0
        fi
    fi
    
    # Detect architecture and prepare download
    ARCH=$(detect_arch)
    BINARY_NAME="certfix-agent-linux-${ARCH}"
    DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${LATEST_VERSION}/${BINARY_NAME}"
    TEMP_BINARY="/tmp/certfix-agent-new"
    
    print_info "Downloading new version for architecture: $ARCH"
    
    # Download new binary
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_BINARY"; then
        print_error "Failed to download new version from $DOWNLOAD_URL"
        exit 1
    fi
    
    # Verify download
    if [ ! -f "$TEMP_BINARY" ] || [ ! -s "$TEMP_BINARY" ]; then
        print_error "Downloaded file is empty or corrupted"
        rm -f "$TEMP_BINARY"
        exit 1
    fi
    
    chmod +x "$TEMP_BINARY"
    
    # Test the new binary (basic check)
    if ! "$TEMP_BINARY" --help >/dev/null 2>&1 && ! "$TEMP_BINARY" --version >/dev/null 2>&1; then
        print_warning "New binary might not be compatible (failed basic test)"
        read -p "Continue anyway? (y/N): " force_continue
        if [[ ! "$force_continue" =~ ^[Yy]$ ]]; then
            rm -f "$TEMP_BINARY"
            exit 1
        fi
    fi
    
    # Create backup
    backup_binary
    
    # Stop service
    print_info "Stopping service..."
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        systemctl stop "$SERVICE_NAME"
        SERVICE_WAS_RUNNING=true
    else
        SERVICE_WAS_RUNNING=false
    fi
    
    # Replace binary
    print_info "Installing new binary..."
    if ! mv "$TEMP_BINARY" "$BIN_PATH"; then
        print_error "Failed to replace binary"
        rollback
        exit 1
    fi
    
    # Update configuration
    update_config_version "$LATEST_VERSION"
    
    # Start service if it was running
    if [ "$SERVICE_WAS_RUNNING" = true ]; then
        print_info "Starting service..."
        if ! systemctl start "$SERVICE_NAME"; then
            print_error "Failed to start service after update"
            rollback
            exit 1
        fi
        
        # Wait a moment and check if service is running
        sleep 2
        if ! systemctl is-active --quiet "$SERVICE_NAME"; then
            print_error "Service failed to start properly after update"
            rollback
            exit 1
        fi
    fi
    
    # Cleanup
    rm -f "$TEMP_BINARY"
    
    print_success "Update completed successfully!"
    print_info "Updated from $CURRENT_VERSION to $LATEST_VERSION"
    print_info "Architecture: $(uname -m) ($ARCH)"
    
    if [ "$SERVICE_WAS_RUNNING" = true ]; then
        print_info "Service status: $(systemctl is-active $SERVICE_NAME)"
        print_info "To check logs: journalctl -u $SERVICE_NAME -f"
    fi
    
    # Clean old backups (keep last 5)
    if [ -d "$BACKUP_DIR" ]; then
        ls -t "$BACKUP_DIR"/certfix-agent-backup-* 2>/dev/null | tail -n +6 | xargs rm -f 2>/dev/null || true
    fi
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  -y, --yes    Skip confirmation prompt"
        echo "  -h, --help   Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0           # Interactive update"
        echo "  $0 --yes     # Automatic update without confirmation"
        exit 0
        ;;
esac

# Run main function
main "$@"