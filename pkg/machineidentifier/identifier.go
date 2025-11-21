package machineidentifier

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

const (
	MACHINE_ID_FILE = "/etc/certfix-agent/machine-id"
)

// GenerateMachineID creates a unique, stable identifier for this machine
// It uses multiple hardware characteristics to ensure stability across reinstalls
func GenerateMachineID() (string, error) {
	// First, check if we already have a stored machine ID
	if id, err := loadStoredMachineID(); err == nil && id != "" {
		return id, nil
	}

	// Generate new machine ID based on hardware characteristics
	id, err := generateFromHardware()
	if err != nil {
		return "", fmt.Errorf("failed to generate machine ID: %w", err)
	}

	// Store it for future use
	if err := storeMachineID(id); err != nil {
		// Log warning but don't fail - the ID is still valid
		fmt.Fprintf(os.Stderr, "Warning: failed to store machine ID: %v\n", err)
	}

	return id, nil
}

// loadStoredMachineID reads the machine ID from disk if it exists
func loadStoredMachineID() (string, error) {
	data, err := os.ReadFile(MACHINE_ID_FILE)
	if err != nil {
		return "", err
	}

	id := strings.TrimSpace(string(data))
	if len(id) != 64 { // SHA-256 hex = 64 chars
		return "", fmt.Errorf("invalid stored machine ID length")
	}

	return id, nil
}

// storeMachineID saves the machine ID to disk
func storeMachineID(id string) error {
	// Ensure directory exists
	dir := "/etc/certfix-agent"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write machine ID with restrictive permissions
	if err := os.WriteFile(MACHINE_ID_FILE, []byte(id), 0644); err != nil {
		return fmt.Errorf("failed to write machine ID file: %w", err)
	}

	return nil
}

// generateFromHardware creates a machine ID from hardware characteristics
func generateFromHardware() (string, error) {
	var components []string

	// 1. System UUID (most stable identifier)
	if uuid := getSystemUUID(); uuid != "" {
		components = append(components, uuid)
	}

	// 2. MAC addresses (stable unless network hardware changes)
	if macs := getMACAddresses(); len(macs) > 0 {
		components = append(components, strings.Join(macs, ","))
	}

	// 3. Machine ID from OS (if available)
	if osID := getOSMachineID(); osID != "" {
		components = append(components, osID)
	}

	// 4. CPU info (relatively stable)
	if cpuID := getCPUInfo(); cpuID != "" {
		components = append(components, cpuID)
	}

	// 5. Boot ID (changes on reboot, but used as fallback)
	if bootID := getBootID(); bootID != "" {
		components = append(components, bootID)
	}

	if len(components) == 0 {
		return "", fmt.Errorf("could not collect any hardware identifiers")
	}

	// Combine all components and hash
	combined := strings.Join(components, "|")
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:]), nil
}

// getSystemUUID retrieves the system/motherboard UUID
func getSystemUUID() string {
	var uuid string

	switch runtime.GOOS {
	case "linux":
		// Try DMI/SMBIOS UUID
		data, err := os.ReadFile("/sys/class/dmi/id/product_uuid")
		if err == nil {
			uuid = strings.TrimSpace(string(data))
			if uuid != "" && uuid != "00000000-0000-0000-0000-000000000000" {
				return uuid
			}
		}

		// Try board serial
		data, err = os.ReadFile("/sys/class/dmi/id/board_serial")
		if err == nil {
			serial := strings.TrimSpace(string(data))
			if serial != "" && serial != "None" {
				return serial
			}
		}

	case "darwin":
		// macOS: Use IOPlatformUUID
		cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "IOPlatformUUID") {
					parts := strings.Split(line, "\"")
					if len(parts) >= 4 {
						return parts[3]
					}
				}
			}
		}

		// Fallback to hardware UUID
		cmd = exec.Command("system_profiler", "SPHardwareDataType")
		output, err = cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Hardware UUID:") {
					parts := strings.Split(line, ":")
					if len(parts) >= 2 {
						return strings.TrimSpace(parts[1])
					}
				}
			}
		}
	}

	return uuid
}

// getMACAddresses retrieves all non-loopback MAC addresses
func getMACAddresses() []string {
	var macs []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return macs
	}

	for _, iface := range interfaces {
		// Skip loopback, virtual, and empty MACs
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		mac := iface.HardwareAddr.String()
		if mac != "" && !isVirtualMAC(mac) {
			macs = append(macs, mac)
		}
	}

	// Sort for consistency
	sort.Strings(macs)
	return macs
}

// isVirtualMAC checks if a MAC address belongs to a virtual interface
func isVirtualMAC(mac string) bool {
	// Common virtual MAC prefixes
	virtualPrefixes := []string{
		"00:00:00", // Null MAC
		"00:05:69", // VMware
		"00:0c:29", // VMware
		"00:50:56", // VMware
		"00:1c:14", // VMware
		"08:00:27", // VirtualBox
		"00:15:5d", // Hyper-V
		"00:16:3e", // Xen
	}

	macLower := strings.ToLower(mac)
	for _, prefix := range virtualPrefixes {
		if strings.HasPrefix(macLower, strings.ToLower(prefix)) {
			return true
		}
	}

	return false
}

// getOSMachineID retrieves the OS-level machine ID
func getOSMachineID() string {
	var paths []string

	switch runtime.GOOS {
	case "linux":
		paths = []string{
			"/etc/machine-id",
			"/var/lib/dbus/machine-id",
		}
	case "darwin":
		// macOS uses IOPlatformUUID (already in getSystemUUID)
		return ""
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			id := strings.TrimSpace(string(data))
			if id != "" {
				return id
			}
		}
	}

	return ""
}

// getCPUInfo retrieves CPU identification information
func getCPUInfo() string {
	switch runtime.GOOS {
	case "linux":
		// Read CPU serial or processor ID
		data, err := os.ReadFile("/proc/cpuinfo")
		if err != nil {
			return ""
		}

		var serial, processor string
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Serial") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					serial = strings.TrimSpace(parts[1])
				}
			}
			if strings.HasPrefix(line, "processor") && processor == "" {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					processor = strings.TrimSpace(parts[1])
				}
			}
		}

		if serial != "" && serial != "0000000000000000" {
			return serial
		}
		return processor

	case "darwin":
		// macOS: Use CPU brand string
		cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	}

	return ""
}

// getBootID retrieves the boot instance ID (changes on reboot)
func getBootID() string {
	switch runtime.GOOS {
	case "linux":
		// Linux boot_id changes on every boot
		data, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	case "darwin":
		// macOS: Use boot time
		cmd := exec.Command("sysctl", "-n", "kern.boottime")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	}

	return ""
}

// GetMachineFingerprint returns a human-readable fingerprint of the machine
func GetMachineFingerprint() string {
	id, err := GenerateMachineID()
	if err != nil {
		return "unknown"
	}

	// Return first 16 characters for display
	if len(id) >= 16 {
		return id[:16]
	}
	return id
}

// ValidateMachineID checks if a stored machine ID is still valid
func ValidateMachineID(storedID string) bool {
	currentID, err := generateFromHardware()
	if err != nil {
		return false
	}

	return storedID == currentID
}
