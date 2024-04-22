package config

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"gioui.org/app"
	"github.com/go-gost/core/logger"
	xconfig "github.com/go-gost/x/config"
	logger_parser "github.com/go-gost/x/config/parsing/logger"
	"gopkg.in/yaml.v3"
)

const (
	configFile = "config.yml"
)

var (
	configDir string
)

func init() {
	config.Store(&Config{})
}

func Init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))

	dir, err := app.DataDir()
	if err != nil {
		slog.Error(fmt.Sprintf("appDir: %v", err))
	}
	if dir == "" {
		dir, _ = os.Getwd()
	}
	configDir = filepath.Join(dir, "gost.plus")
	os.MkdirAll(configDir, 0755)

	slog.Info(fmt.Sprintf("appDir: %s", configDir))

	cfg := Get()
	if err := cfg.load(); err != nil {
		slog.Error(fmt.Sprintf("load config: %v", err))
		if _, ok := err.(*os.PathError); ok {
			cfg.Write()
		}
	}
	Set(cfg)

	initLog()
}

func initLog() {
	cfg := Get().Log
	if cfg == nil {
		cfg = &xconfig.LogConfig{
			Output: "stdout",
			Level:  string(logger.InfoLevel),
			Format: string(logger.JSONFormat),
		}
	}

	logger.SetDefault(logger_parser.ParseLogger(&xconfig.LoggerConfig{Log: cfg}))
}

var (
	config atomic.Value
)

func Get() *Config {
	c := config.Load().(*Config)
	cfg := &Config{}
	*cfg = *c
	return cfg
}

func Set(c *Config) {
	if c == nil {
		c = &Config{}
	}
	config.Store(c)
}

type Settings struct {
	Lang  string
	Theme string
}

type Tunnel struct {
	ID        string
	Name      string
	Type      string
	Endpoint  string
	Hostname  string `yaml:",omitempty"`
	Username  string `yaml:",omitempty"`
	Password  string `yaml:",omitempty"`
	EnableTLS bool   `yaml:"enableTLS,omitempty"`
	Keepalive bool   `yaml:",omitempty"`
	TTL       int    `yaml:"ttl,omitempty"`

	Stats     ServiceStats
	Favorite  bool
	Closed    bool
	CreatedAt time.Time
}

type Config struct {
	Settings    *Settings
	Tunnels     []*Tunnel
	EntryPoints []*Tunnel
	Log         *xconfig.LogConfig
}

func (c *Config) load() error {
	f, err := os.Open(filepath.Join(configDir, configFile))
	if err != nil {
		return err
	}
	defer f.Close()

	return yaml.NewDecoder(f).Decode(c)
}

func (c *Config) Write() error {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	defer enc.Close()

	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(configDir, configFile), buf.Bytes(), 0644)
}

type ServiceStats struct {
	Time            time.Time
	TotalConns      uint64
	RequestRate     float64
	CurrentConns    uint64
	TotalErrs       uint64
	InputBytes      uint64
	InputRateBytes  uint64
	OutputBytes     uint64
	OutputRateBytes uint64
}
