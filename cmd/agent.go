package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	CONFIG_FILE       = "/etc/certfix-agent/config.json"
	AGENT_VERSION     = "0.1.0"
	HEARTBEAT_INTERVAL = 5 * time.Minute
	REGISTER_RETRY_DELAY = 30 * time.Second
)

type Config struct {
	Token          string `json:"token"`
	Endpoint       string `json:"endpoint"`
	CurrentVersion string `json:"current_version"`
	Architecture   string `json:"architecture"`
}

type InstanceData struct {
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

	return &config, nil
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
func collectInstanceData() *InstanceData {
	return &InstanceData{
		Hostname:     getHostname(),
		OSType:       runtime.GOOS,
		OSVersion:    getOSVersion(),
		Architecture: runtime.GOARCH,
		IPAddress:    getIPAddress(),
		MACAddress:   getMACAddress(),
		AgentVersion: AGENT_VERSION,
		Metadata: map[string]interface{}{
			"num_cpu":   runtime.NumCPU(),
			"go_version": runtime.Version(),
		},
	}
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
	log.Println("[certfix-agent] Starting agent version", AGENT_VERSION)

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("[FATAL] Failed to load configuration: %v", err)
	}

	log.Printf("[INFO] Configuration loaded from %s", CONFIG_FILE)
	log.Printf("[INFO] Endpoint: %s", config.Endpoint)

	// Collect instance data
	instanceData := collectInstanceData()
	log.Printf("[INFO] Instance Info: %s (%s %s) on %s", 
		instanceData.Hostname, 
		instanceData.OSType, 
		instanceData.Architecture,
		instanceData.OSVersion,
	)

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
