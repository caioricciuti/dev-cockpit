package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	// General settings
	Theme           string `mapstructure:"theme"`
	UpdateInterval  int    `mapstructure:"update_interval"`
	EnableTelemetry bool   `mapstructure:"enable_telemetry"`
	LogLevel        string `mapstructure:"log_level"`

	// UI settings
	UI UIConfig `mapstructure:"ui"`

	// Module settings
	Modules ModulesConfig `mapstructure:"modules"`

	// System settings
	System SystemConfig `mapstructure:"system"`

	// Storage settings
	Storage StorageConfig `mapstructure:"storage"`
}

// UIConfig holds UI-related configuration
type UIConfig struct {
	ColorScheme    string `mapstructure:"color_scheme"`
	AnimationSpeed int    `mapstructure:"animation_speed"`
	ShowFPS        bool   `mapstructure:"show_fps"`
	MouseEnabled   bool   `mapstructure:"mouse_enabled"`
}

// ModulesConfig holds module-specific configuration
type ModulesConfig struct {
	Dashboard DashboardConfig `mapstructure:"dashboard"`
	Docker    DockerConfig    `mapstructure:"docker"`
	Network   NetworkConfig   `mapstructure:"network"`
	Security  SecurityConfig  `mapstructure:"security"`
}

// DashboardConfig holds dashboard module configuration
type DashboardConfig struct {
	RefreshRate     int  `mapstructure:"refresh_rate"`
	ShowCPUDetails  bool `mapstructure:"show_cpu_details"`
	ShowMemDetails  bool `mapstructure:"show_mem_details"`
	ShowDiskDetails bool `mapstructure:"show_disk_details"`
	GraphHeight     int  `mapstructure:"graph_height"`
}

// DockerConfig holds Docker module configuration
type DockerConfig struct {
	SocketPath        string `mapstructure:"socket_path"`
	ShowAllContainers bool   `mapstructure:"show_all_containers"`
	AutoRefresh       bool   `mapstructure:"auto_refresh"`
}

// NetworkConfig holds network module configuration
type NetworkConfig struct {
	DefaultInterface string `mapstructure:"default_interface"`
	PacketCapture    bool   `mapstructure:"packet_capture"`
	PortScanTimeout  int    `mapstructure:"port_scan_timeout"`
}

// SecurityConfig holds security module configuration
type SecurityConfig struct {
	ScanInterval   int  `mapstructure:"scan_interval"`
	CheckFirewall  bool `mapstructure:"check_firewall"`
	CheckFileVault bool `mapstructure:"check_filevault"`
	CheckSIP       bool `mapstructure:"check_sip"`
}

// SystemConfig holds system-related configuration
type SystemConfig struct {
	CommandTimeout int    `mapstructure:"command_timeout"`
	MaxRetries     int    `mapstructure:"max_retries"`
	SudoCommand    string `mapstructure:"sudo_command"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	DataDir         string `mapstructure:"data_dir"`
	MaxHistoryDays  int    `mapstructure:"max_history_days"`
	CompressOldData bool   `mapstructure:"compress_old_data"`
}

// Load loads configuration from file and environment
func Load() (*Config, error) {
	homeDir, _ := os.UserHomeDir()
	// Prefer ~/.devcockpit, but gracefully fall back if not writable (sandboxed)
	primaryDir := ""
	if homeDir != "" {
		primaryDir = filepath.Join(homeDir, ".devcockpit")
	}
	fallbackDir := ".devcockpit" // current working directory

	configDir := primaryDir
	if configDir == "" || os.MkdirAll(configDir, 0755) != nil {
		// Fallback path in workspace or CWD
		_ = os.MkdirAll(fallbackDir, 0755)
		configDir = fallbackDir
	}
	configFile := filepath.Join(configDir, "config.yaml")

	// Set up Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")
	// Ensure viper writes back to the selected file
	viper.SetConfigFile(configFile)

	// Set defaults
	setDefaults()

	// Enable environment variables
	viper.SetEnvPrefix("DEVCOCKPIT")
	viper.AutomaticEnv()

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		// If config file doesn't exist, create it with defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			_ = createDefaultConfig(configFile) // Best-effort; continue on failure
			// Try reading again
			_ = viper.ReadInConfig()
		} else {
			// Any other error: proceed with defaults rather than failing hard
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// General defaults
	viper.SetDefault("theme", "dark")
	viper.SetDefault("update_interval", 1000) // 1 second in milliseconds
	viper.SetDefault("enable_telemetry", false)
	viper.SetDefault("log_level", "info")

	// UI defaults
	viper.SetDefault("ui.color_scheme", "cyberpunk")
	viper.SetDefault("ui.animation_speed", 60) // FPS
	viper.SetDefault("ui.show_fps", false)
	viper.SetDefault("ui.mouse_enabled", true)

	// Dashboard defaults
	viper.SetDefault("modules.dashboard.refresh_rate", 1)
	viper.SetDefault("modules.dashboard.show_cpu_details", true)
	viper.SetDefault("modules.dashboard.show_mem_details", true)
	viper.SetDefault("modules.dashboard.show_disk_details", true)
	viper.SetDefault("modules.dashboard.graph_height", 10)

	// Docker defaults
	viper.SetDefault("modules.docker.socket_path", "/var/run/docker.sock")
	viper.SetDefault("modules.docker.show_all_containers", false)
	viper.SetDefault("modules.docker.auto_refresh", true)

	// Network defaults
	viper.SetDefault("modules.network.default_interface", "en0")
	viper.SetDefault("modules.network.packet_capture", false)
	viper.SetDefault("modules.network.port_scan_timeout", 2)

	// Security defaults
	viper.SetDefault("modules.security.scan_interval", 300) // 5 minutes
	viper.SetDefault("modules.security.check_firewall", true)
	viper.SetDefault("modules.security.check_filevault", true)
	viper.SetDefault("modules.security.check_sip", true)

	// System defaults
	viper.SetDefault("system.command_timeout", 30)
	viper.SetDefault("system.max_retries", 3)
	viper.SetDefault("system.sudo_command", "sudo")

	// Storage defaults
	homeDir, _ := os.UserHomeDir()
	viper.SetDefault("storage.data_dir", filepath.Join(homeDir, ".devcockpit", "data"))
	viper.SetDefault("storage.max_history_days", 30)
	viper.SetDefault("storage.compress_old_data", true)
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(configFile string) error {
	defaultConfig := `# Dev Cockpit Configuration
# https://devcockpit.dev/docs/configuration

# General Settings
theme: dark
update_interval: 1000
enable_telemetry: false
log_level: info

# UI Settings
ui:
  color_scheme: cyberpunk
  animation_speed: 60
  show_fps: false
  mouse_enabled: true

# Module Settings
modules:
  dashboard:
    refresh_rate: 1
    show_cpu_details: true
    show_mem_details: true
    show_disk_details: true
    graph_height: 10

  docker:
    socket_path: /var/run/docker.sock
    show_all_containers: false
    auto_refresh: true

  network:
    default_interface: en0
    packet_capture: false
    port_scan_timeout: 2

  security:
    scan_interval: 300
    check_firewall: true
    check_filevault: true
    check_sip: true

# System Settings
system:
  command_timeout: 30
  max_retries: 3
  sudo_command: sudo

# Storage Settings
storage:
  max_history_days: 30
  compress_old_data: true
`

	return os.WriteFile(configFile, []byte(defaultConfig), 0644)
}

// Save saves the current configuration to file
func (c *Config) Save() error {
	return viper.WriteConfig()
}
