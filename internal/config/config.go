package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/patnaikankit/Gaze/pkg/logger"
	yamlv2 "gopkg.in/yaml.v2"
)

type Config struct {
	WatchDir      string   `json:"watchDir" yaml:"watchDir"`
	BuildCMD      string   `json:"buildCmd" yaml:"buildCmd"`
	RunCMD        string   `json:"runCmd" yaml:"runCmd"`
	IgnorePattern []string `json:"ignorePattern" yaml:"ignorePattern"`
	Port          string   `json:"port" yaml:"port"`
	IsWindows     bool     `json:"-" yaml:"-"`
	MainFile      string   `json:"main" yaml:"main"`
}

// Default configuration
func defaultConfig() *Config {
	return &Config{
		WatchDir: ".",
		IgnorePattern: []string{
			"temp", "temp/*",
			".git", ".git/*",
			"node_modules", "node_modules/*",
			"vendor", "vendor/*",
			"*.exe", "*.tmp", "*.log",
		},
		BuildCMD:  "",
		RunCMD:    "",
		Port:      "",
		IsWindows: runtime.GOOS == "windows",
	}
}

// checks for JSON/YAML config and merges into base config
func LoadConfigFile(base *Config, path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	fileConfig := filepath.Ext(path)
	switch fileConfig {
	case ".json":
		if err := json.Unmarshal(data, base); err != nil {
			return nil, fmt.Errorf("failed parsing JSON: %w", err)
		}

	case ".yaml", ".yml":
		if err := yamlv2.Unmarshal(data, base); err != nil {
			return nil, fmt.Errorf("failed parsing YAML: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported config format: %s", err)
	}

	return base, nil
}

func ParseConfig() (*Config, error) {
	cfg := defaultConfig()

	watchDir := flag.String("watch", cfg.WatchDir, "Directory to watch for changes")
	mainFile := flag.String("main", "", "Path to the main Go file (e.g., ./cmd/server/main.go)")
	port := flag.String("port", cfg.Port, "Port number if the app runs an HTTP server (e.g., 8080)")
	configFile := flag.String("config", "", "Path to JSON/YAML config file")

	flag.Parse()

	// Merge from file if provided
	if *configFile != "" {
		loadedCfg, err := LoadConfigFile(cfg, *configFile)
		if err != nil {
			logger.Error("Failed to load config file: %v", err)
			return nil, err
		}

		cfg = loadedCfg
	}

	// CLI flags override file values
	if *watchDir != "" {
		cfg.WatchDir = *watchDir
	}
	if *mainFile != "" {
		cfg.MainFile = *mainFile
	}
	if *port != "" {
		cfg.Port = *port
	}

	if cfg.MainFile == "" {
		logger.Error("Main file must be specified using --main flag or in the config file")
		flag.Usage()
		return nil, fmt.Errorf("main file not specified")
	}

	// Platform-specific build/run commands
	if cfg.IsWindows {
		cfg.BuildCMD = fmt.Sprintf("go build -mod=mod -o .\\temp\\gaze.exe %s", cfg.MainFile)
		cfg.RunCMD = ".\\temp\\gaze.exe"
	} else {
		cfg.BuildCMD = fmt.Sprintf("go build -mod=mod -o ./temp/gaze %s", cfg.MainFile)
		cfg.RunCMD = "./temp/gaze"
	}

	if isWebServer(cfg.MainFile) && cfg.Port == "" {
		logger.Info("Warning: The application appears to be a web server but no port is specified. Use --port flag or set PORT env variable.")
	}

	logger.Debug("Configuration: %+v", *cfg)
	return cfg, nil
}

func isWebServer(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Warn("Could not read main file to determine if it's a web server: %v", err)
		return false
	}

	text := strings.ToLower(string(data))

	signatures := []string{
		"net/http", "http.listenandserve", "router", "handlefunc",
		"gin.default", "echo.new", "fiber.new", "mux.newrouter",
	}

	for _, sig := range signatures {
		if strings.Contains(text, sig) {
			return true
		}
	}
	return false
}
