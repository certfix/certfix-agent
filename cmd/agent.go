package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/certfix/certfix-agent/pkg/machineidentifier"
)

const (
	CONFIG_FILE       = "/etc/certfix-agent/config.json"
	DEFAULT_VERSION   = "0.0.0"
	HEARTBEAT_INTERVAL = 5 * time.Minute
	REGISTER_RETRY_DELAY = 30 * time.Second
)

type Config struct {
	Token          string `json:"token"`
	Endpoint       string `json:"endpoint"`
	CurrentVersion string `json:"current_version,omitempty"`
	Architecture   string `json:"architecture,omitempty"`
}

type InstanceData struct {
	MachineID    string                 `json:"machine_id"`
	Hostname     string                 `json:"hostname"`
	OSType       string                 `json:"os_type"`
	OSVersion    string                 `json:"os_version"`
	Architecture string                 `json:"architecture"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	MACAddress   string                 `json:"mac_address,omitempty"`
	AgentVersion string                 `json:"agent_version"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type RegisterResponse struct {
	InstanceID  string `json:"instance_id"`
	KeyID       string `json:"key_id"`
	ServiceHash string `json:"service_hash"`
	ServiceName string `json:"service_name"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

// Load configuration from file
func loadConfig() (*Config, error) {
	data, err := os.ReadFile(CONFIG_FILE)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.Token == "" {
		return nil, fmt.Errorf("token is required in config file")
	}

	if config.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required in config file")
	}

	// Set default version if not specified
	if config.CurrentVersion == "" {
		config.CurrentVersion = DEFAULT_VERSION
	}

	return &config, nil
}

// Save configuration to file
func saveConfig(config *Config) error {
	// Create directory if it doesn't exist
	configDir := filepath.Dir(CONFIG_FILE)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file with 0644 permissions (owner read/write, others read)
	if err := os.WriteFile(CONFIG_FILE, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Ensure correct permissions (in case umask interferes)
	if err := os.Chmod(CONFIG_FILE, 0644); err != nil {
		return fmt.Errorf("failed to set config file permissions: %w", err)
	}

	return nil
}

// Get hostname
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("[WARNING] Failed to get hostname: %v", err)
		return "unknown"
	}
	return hostname
}

// Get OS version
func getOSVersion() string {
	switch runtime.GOOS {
	case "linux":
		// Try to read /etc/os-release
		data, err := os.ReadFile("/etc/os-release")
		if err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				}
			}
		}
		
		// Fallback to uname
		cmd := exec.Command("uname", "-r")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	case "darwin":
		cmd := exec.Command("sw_vers", "-productVersion")
		output, err := cmd.Output()
		if err == nil {
			return "macOS " + strings.TrimSpace(string(output))
		}
	}
	return "unknown"
}

// Get primary IP address
func getIPAddress() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// Get MAC address of primary interface
func getMACAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Return first valid MAC address
		if iface.HardwareAddr.String() != "" {
			return iface.HardwareAddr.String()
		}
	}
	return ""
}

// Collect instance data
func collectInstanceData(version string) (*InstanceData, error) {
	// Generate machine ID
	machineID, err := machineidentifier.GenerateMachineID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate machine ID: %w", err)
	}

	return &InstanceData{
		MachineID:    machineID,
		Hostname:     getHostname(),
		OSType:       runtime.GOOS,
		OSVersion:    getOSVersion(),
		Architecture: runtime.GOARCH,
		IPAddress:    getIPAddress(),
		MACAddress:   getMACAddress(),
		AgentVersion: version,
		Metadata: map[string]interface{}{
			"num_cpu":      runtime.NumCPU(),
			"go_version":   runtime.Version(),
			"fingerprint":  machineidentifier.GetMachineFingerprint(),
		},
	}, nil
}

// Register instance with the API
func registerInstance(config *Config, instanceData *InstanceData) (*RegisterResponse, error) {
	// Prepare request body
	reqBody, err := json.Marshal(instanceData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal instance data: %w", err)
	}

	// Create HTTP request
	url := strings.TrimRight(config.Endpoint, "/") + "/instances/register"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", config.Token)

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var registerResp RegisterResponse
	if err := json.Unmarshal(body, &registerResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &registerResp, nil
}

// Send heartbeat to update last_seen_at
func sendHeartbeat(config *Config, instanceID string) error {
	url := strings.TrimRight(config.Endpoint, "/") + "/instances/" + instanceID + "/heartbeat"
	
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create heartbeat request: %w", err)
	}

	req.Header.Set("X-API-Key", config.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "configure":
		handleConfigure()
	case "config":
		handleShowConfig()
	case "start":
		handleStart()
	case "version":
		handleVersion()
	case "machine-id":
		handleMachineID()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf("CertFix Agent v%s\n\n", getVersionString())
	fmt.Println("Usage:")
	fmt.Println("  certfix-agent configure --token <api-key> --endpoint <url>")
	fmt.Println("  certfix-agent config")
	fmt.Println("  certfix-agent start")
	fmt.Println("  certfix-agent machine-id")
	fmt.Println("  certfix-agent version")
	fmt.Println("  certfix-agent help")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  configure  Configure agent with token and endpoint")
	fmt.Println("  config     Show current configuration")
	fmt.Println("  start      Start the agent service")
	fmt.Println("  machine-id Show unique machine identifier")
	fmt.Println("  version    Show version information")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Configure Options:")
	fmt.Println("  --token     API token for authentication (required)")
	fmt.Println("  --endpoint  API endpoint URL (required)")
}

func getVersionString() string {
	config, err := loadConfig()
	if err != nil {
		return DEFAULT_VERSION
	}
	return config.CurrentVersion
}

func handleVersion() {
	fmt.Printf("CertFix Agent v%s\n", getVersionString())
	fmt.Printf("OS: %s\n", runtime.GOOS)
	fmt.Printf("Architecture: %s\n", runtime.GOARCH)
	fmt.Printf("Go Version: %s\n", runtime.Version())
}

func handleMachineID() {
	machineID, err := machineidentifier.GenerateMachineID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to generate machine ID: %v\n", err)
		os.Exit(1)
	}

	fingerprint := machineidentifier.GetMachineFingerprint()
	
	fmt.Println("Machine Identifier Information")
	fmt.Println("==============================")
	fmt.Printf("Full ID:      %s\n", machineID)
	fmt.Printf("Fingerprint:  %s\n", fingerprint)
	fmt.Printf("Hostname:     %s\n", getHostname())
	fmt.Printf("OS:           %s\n", runtime.GOOS)
	fmt.Printf("Architecture: %s\n", runtime.GOARCH)
	
	// Check if machine ID file exists
	if _, err := os.Stat(machineidentifier.MACHINE_ID_FILE); err == nil {
		fmt.Printf("\nStored at:    %s\n", machineidentifier.MACHINE_ID_FILE)
	} else {
		fmt.Printf("\nNote: Machine ID will be stored at %s on first registration\n", machineidentifier.MACHINE_ID_FILE)
	}
}

func handleShowConfig() {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("[ERROR] Failed to load configuration: %v\n", err)
		fmt.Printf("[INFO] Config file: %s\n", CONFIG_FILE)
		fmt.Println("[INFO] Run 'certfix-agent configure' to set up the agent")
		os.Exit(1)
	}

	fmt.Println("Current Configuration:")
	fmt.Println("─────────────────────────────────────────────────")
	fmt.Printf("Config File:  %s\n", CONFIG_FILE)
	fmt.Printf("Version:      %s\n", config.CurrentVersion)
	fmt.Printf("Endpoint:     %s\n", config.Endpoint)
	fmt.Printf("Token:        %s\n", maskToken(config.Token))
	fmt.Printf("Architecture: %s\n", config.Architecture)
	fmt.Println("─────────────────────────────────────────────────")
}

func handleConfigure() {
	configureCmd := flag.NewFlagSet("configure", flag.ExitOnError)
	token := configureCmd.String("token", "", "API token for authentication")
	endpoint := configureCmd.String("endpoint", "", "API endpoint URL")

	configureCmd.Parse(os.Args[2:])

	if *token == "" {
		fmt.Println("Error: --token is required")
		fmt.Println()
		configureCmd.Usage()
		os.Exit(1)
	}

	if *endpoint == "" {
		fmt.Println("Error: --endpoint is required")
		fmt.Println()
		configureCmd.Usage()
		os.Exit(1)
	}

	// Load existing config if available to preserve version
	existingConfig, _ := loadConfig()
	version := DEFAULT_VERSION
	if existingConfig != nil && existingConfig.CurrentVersion != "" {
		version = existingConfig.CurrentVersion
	}

	// Create config
	config := &Config{
		Token:          *token,
		Endpoint:       *endpoint,
		CurrentVersion: version,
		Architecture:   runtime.GOARCH,
	}

	// Save config
	if err := saveConfig(config); err != nil {
		fmt.Printf("\n[ERROR] Failed to save configuration: %v\n", err)
		fmt.Printf("[INFO] Config file location: %s\n", CONFIG_FILE)
		fmt.Println()
		if os.Geteuid() != 0 {
			fmt.Println("⚠️  Permission denied. Try running with sudo:")
			fmt.Printf("   sudo certfix-agent configure --token \"%s\" --endpoint \"%s\"\n", *token, *endpoint)
		} else {
			fmt.Println("⚠️  Ensure the parent directory exists and is writable:")
			fmt.Printf("   sudo mkdir -p %s\n", filepath.Dir(CONFIG_FILE))
			fmt.Printf("   sudo chmod 755 %s\n", filepath.Dir(CONFIG_FILE))
		}
		os.Exit(1)
	}

	fmt.Printf("[SUCCESS] Configuration saved to %s\n", CONFIG_FILE)
	fmt.Printf("[INFO] Token: %s\n", maskToken(*token))
	fmt.Printf("[INFO] Endpoint: %s\n", *endpoint)
	fmt.Println()
	fmt.Println("You can now start the agent with: certfix-agent start")
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func handleStart() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("[FATAL] Failed to load configuration: %v", err)
	}

	log.Println("[certfix-agent] Starting agent version", config.CurrentVersion)
	log.Printf("[INFO] Configuration loaded from %s", CONFIG_FILE)
	log.Printf("[INFO] Endpoint: %s", config.Endpoint)

	// Collect instance data
	instanceData, err := collectInstanceData(config.CurrentVersion)
	if err != nil {
		log.Fatalf("[FATAL] Failed to collect instance data: %v", err)
	}

	log.Printf("[INFO] Instance Info: %s (%s %s) on %s", 
		instanceData.Hostname, 
		instanceData.OSType, 
		instanceData.Architecture,
		instanceData.OSVersion,
	)
	log.Printf("[INFO] Machine ID: %s", instanceData.Metadata["fingerprint"])

	// Register with retry logic
	var registerResp *RegisterResponse
	for {
		log.Println("[INFO] Registering instance with API...")
		registerResp, err = registerInstance(config, instanceData)
		if err != nil {
			log.Printf("[ERROR] Failed to register instance: %v", err)
			log.Printf("[INFO] Retrying in %v...", REGISTER_RETRY_DELAY)
			time.Sleep(REGISTER_RETRY_DELAY)
			continue
		}
		break
	}

	log.Printf("[SUCCESS] Instance registered successfully!")
	log.Printf("[INFO] Instance ID: %s", registerResp.InstanceID)
	log.Printf("[INFO] Service: %s (%s)", registerResp.ServiceName, registerResp.ServiceHash)
	log.Printf("[INFO] Key ID: %s", registerResp.KeyID)

	// Start heartbeat ticker
	heartbeatTicker := time.NewTicker(HEARTBEAT_INTERVAL)
	defer heartbeatTicker.Stop()

	// Main loop
	for {
		select {
		case <-heartbeatTicker.C:
			log.Println("[INFO] Sending heartbeat...")
			if err := sendHeartbeat(config, registerResp.InstanceID); err != nil {
				log.Printf("[ERROR] Heartbeat failed: %v", err)
			} else {
				log.Println("[INFO] Heartbeat sent successfully")
			}
		}
	}
}
