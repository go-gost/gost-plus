package config

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"sync"

	"gioui.org/app"
	"github.com/go-gost/core/logger"
	"github.com/go-gost/x/config"
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

func Init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)

	dir, err := app.DataDir()
	if err != nil {
		log.Println(err)
	}
	if dir == "" {
		dir, _ = os.Getwd()
	}
	configDir = filepath.Join(dir, "gost.plus")
	os.MkdirAll(configDir, 0755)

	log.Println("config dir:", configDir)

	if err := global.Load(); err != nil {
		log.Println(err)
		if _, ok := err.(*os.PathError); ok {
			global.Write()
		}
	}
	if global.Log == nil {
		logDir := filepath.Join(configDir, "logs")
		os.MkdirAll(logDir, 0755)
		log.Println("log dir:", logDir)

		global.Log = &xconfig.LogConfig{
			Output: filepath.Join(logDir, "gost-plus.log"),
			Level:  string(logger.InfoLevel),
			Format: string(logger.JSONFormat),
			Rotation: &xconfig.LogRotationConfig{
				MaxSize:    10,
				MaxAge:     7,
				MaxBackups: 10,
				LocalTime:  true,
				Compress:   true,
			},
		}
	}

	logger.SetDefault(logger_parser.ParseLogger(&xconfig.LoggerConfig{Log: global.Log}))
}

var (
	global    = &Config{}
	globalMux sync.RWMutex
)

func Global() *Config {
	globalMux.RLock()
	defer globalMux.RUnlock()

	cfg := &Config{}
	*cfg = *global
	return cfg
}

func Set(c *Config) {
	globalMux.Lock()
	defer globalMux.Unlock()

	global = c
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

	Favorite bool
	Closed   bool
}

type Config struct {
	Settings    *Settings
	Tunnels     []*Tunnel
	EntryPoints []*Tunnel
	Log         *config.LogConfig
}

func (c *Config) Load() error {
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
