package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Transmission TransmissionConfig `mapstructure:"transmission"`
	Cleaner      CleanerConfig      `mapstructure:"cleaner"`
	Server       ServerConfig       `mapstructure:"server"`
	Daemon       DaemonConfig       `mapstructure:"daemon"`
	DryRun       bool               `mapstructure:"dry_run"`
	LogLevel     string             `mapstructure:"log_level"`
}

// TransmissionConfig holds Transmission connection settings
type TransmissionConfig struct {
	URL      string `mapstructure:"url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// CleanerConfig holds cleaner behavior settings
type CleanerConfig struct {
	MinFreeSpaceRaw string `mapstructure:"min_free_space"` // Can be bytes or with unit (e.g., "100GB")
	MinFreeSpace    int64  `mapstructure:"-"` // Parsed value in bytes
	MinTorrentsPerTracker int `mapstructure:"min_torrents_per_tracker"`
}

// ServerConfig holds web UI server settings
type ServerConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	WebRoot string `mapstructure:"webroot"`
}

// DaemonConfig holds daemon mode settings
type DaemonConfig struct {
	Enabled       bool          `mapstructure:"enabled"`
	CheckInterval time.Duration `mapstructure:"check_interval"`
}

// parseSize parses a size string that can be either a plain number (bytes)
// or a number with a unit suffix (KB, MB, GB, TB)
// Examples: "100", "100GB", "500MB", "1.5TB"
func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	
	// Try to parse as plain number (bytes)
	if val, err := strconv.ParseInt(s, 10, 64); err == nil {
		return val, nil
	}
	
	// Parse with unit suffix
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*(KB|MB|GB|TB|K|M|G|T)?$`)
	matches := re.FindStringSubmatch(strings.ToUpper(s))
	
	if matches == nil {
		return 0, fmt.Errorf("invalid size format: %s", s)
	}
	
	numStr := matches[1]
	unit := matches[2]
	
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", numStr)
	}
	
	multiplier := int64(1)
	switch unit {
	case "KB", "K":
		multiplier = 1024
	case "MB", "M":
		multiplier = 1024 * 1024
	case "GB", "G":
		multiplier = 1024 * 1024 * 1024
	case "TB", "T":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "":
		// No unit, already in bytes
		multiplier = 1
	}
	
	return int64(num * float64(multiplier)), nil
}

// Load loads configuration from file, environment variables, and CLI flags
func Load(version string) (*Config, error) {
	// Set defaults
	viper.SetDefault("transmission.url", "http://localhost:9091/transmission/rpc")
	viper.SetDefault("cleaner.min_free_space", 100*1024*1024*1024) // 100 GB
	viper.SetDefault("cleaner.min_torrents_per_tracker", 2)
	viper.SetDefault("server.enabled", false)
	viper.SetDefault("server.port", 8888)
	viper.SetDefault("server.webroot", "/")
	viper.SetDefault("daemon.enabled", false)
	viper.SetDefault("daemon.check_interval", "1m")
	viper.SetDefault("dry_run", false)
	viper.SetDefault("log_level", "info")

	// Setup CLI flags
	pflag.StringP("transmission-url", "u", "", "Transmission RPC URL")
	pflag.StringP("transmission-user", "U", "", "Transmission username")
	pflag.StringP("transmission-pass", "P", "", "Transmission password")
	pflag.Int64P("min-free-space", "s", 0, "Minimum free space in GB (default: 100)")
	pflag.IntP("min-torrents", "m", 0, "Minimum torrents per tracker (default: 2)")
	pflag.BoolP("daemon", "d", false, "Run in daemon mode")
	pflag.DurationP("check-interval", "i", 0, "Check interval in daemon mode (default: 1m)")
	pflag.BoolP("dry-run", "n", false, "Dry run mode (simulate only)")
	pflag.BoolP("web-ui", "w", false, "Enable web UI")
	pflag.IntP("web-port", "p", 0, "Web UI port (default: 8888)")
	pflag.StringP("web-root", "r", "", "Web UI root path for reverse proxy (default: /)")
	pflag.StringP("config", "c", "", "Config file path")
	pflag.StringP("log-level", "l", "", "Log level (debug, info, warn, error)")
	pflag.BoolP("version", "v", false, "Show version and exit")
	pflag.Parse()
	
	// Handle version flag
	if pflag.Lookup("version").Changed {
		fmt.Printf("btcleaner version %s\n", version)
		os.Exit(0)
	}

	// Bind flags to viper
	viper.BindPFlag("transmission.url", pflag.Lookup("transmission-url"))
	viper.BindPFlag("transmission.username", pflag.Lookup("transmission-user"))
	viper.BindPFlag("transmission.password", pflag.Lookup("transmission-pass"))
	viper.BindPFlag("daemon.enabled", pflag.Lookup("daemon"))
	viper.BindPFlag("daemon.check_interval", pflag.Lookup("check-interval"))
	viper.BindPFlag("dry_run", pflag.Lookup("dry-run"))
	viper.BindPFlag("server.enabled", pflag.Lookup("web-ui"))
	viper.BindPFlag("server.port", pflag.Lookup("web-port"))
	viper.BindPFlag("server.webroot", pflag.Lookup("web-root"))
	viper.BindPFlag("log_level", pflag.Lookup("log-level"))

	// Handle min-free-space (convert GB to bytes)
	if pflag.Lookup("min-free-space").Changed {
		gb := pflag.Lookup("min-free-space").Value.String()
		if gb != "" && gb != "0" {
			var gbInt int64
			fmt.Sscanf(gb, "%d", &gbInt)
			viper.Set("cleaner.min_free_space", gbInt*1024*1024*1024)
		}
	}
	
	// Handle min-torrents
	if pflag.Lookup("min-torrents").Changed {
		mt := pflag.Lookup("min-torrents").Value.String()
		if mt != "" && mt != "0" {
			var mtInt int
			fmt.Sscanf(mt, "%d", &mtInt)
			viper.Set("cleaner.min_torrents_per_tracker", mtInt)
		}
	}

	// Environment variables
	viper.SetEnvPrefix("BTCLEANER")
	viper.AutomaticEnv()

	// Config file
	configFile := pflag.Lookup("config").Value.String()
	if configFile != "" {
		// User specified a config file
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("error reading config file %s: %w", configFile, err)
		}
	} else {
		// Try multiple config file locations in order
		configPaths := []string{
			"./btcleaner.yaml",
			"./config.yaml",
			os.ExpandEnv("$HOME/.config/btcleaner.yaml"),
			"/etc/btcleaner.yaml",
		}
		
		configLoaded := false
		for _, path := range configPaths {
			if _, err := os.Stat(path); err == nil {
				viper.SetConfigFile(path)
				if err := viper.ReadInConfig(); err != nil {
					return nil, fmt.Errorf("error reading config file %s: %w", path, err)
				}
				configLoaded = true
				break
			}
		}
		
		// Config file is optional, so don't error if not found
		if !configLoaded {
			// No config file found, use defaults
		}
	}

	// Unmarshal config
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Parse min_free_space from config file (can be with units like "100GB")
	if cfg.Cleaner.MinFreeSpaceRaw != "" {
		parsed, err := parseSize(cfg.Cleaner.MinFreeSpaceRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid min_free_space value: %w", err)
		}
		cfg.Cleaner.MinFreeSpace = parsed
	} else {
		// Use default if not specified
		cfg.Cleaner.MinFreeSpace = 100 * 1024 * 1024 * 1024 // 100 GB
	}

	// Validate required fields
	if cfg.Transmission.URL == "" {
		return nil, fmt.Errorf("transmission URL is required")
	}

	return &cfg, nil
}

// GenerateExampleConfig generates an example configuration file
func GenerateExampleConfig(path string) error {
	example := `# BTCleaner Configuration File

# Transmission settings
transmission:
  url: "http://localhost:9091/transmission/rpc"
  username: ""
  password: ""

# Cleaner settings
cleaner:
  # Minimum free space (can use units: GB, MB, KB, or raw bytes)
  # Examples: "100GB", "500MB", "1024", 107374182400
  min_free_space: "100GB"
  # Minimum torrents to keep per tracker
  min_torrents_per_tracker: 2

# Web UI settings
server:
  enabled: false
  port: 8888
  # Webroot for reverse proxy (default: "/", example for reverse proxy: "/btcleaner")
  webroot: "/"

# Daemon mode settings
daemon:
  enabled: false
  # Check interval (e.g., "1m", "5m", "1h")
  check_interval: "1m"

# Dry run mode (simulate only, don't delete)
dry_run: false

# Log level (debug, info, warn, error)
log_level: "info"
`
	return os.WriteFile(path, []byte(example), 0644)
}
