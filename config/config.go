package config

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Transport TransportConfig `yaml:"transport"`
	App AppConfig `yaml:"app"`
	DB  DBConfig  `yaml:"db"`
	Chat ChatConfig `yaml:"chat"`
}

type TransportConfig struct {
	Type       string `yaml:"type"`        // "tcp" / "tls"
	Port	   string `yaml:"port"`        // 9000
	CertFile   string `yaml:"cert_file"`   // pro TLS
	KeyFile    string `yaml:"key_file"`    // pro TLS
}

type AppConfig struct {
	Name    string        `yaml:"name"`
	Env     string        `yaml:"env"`      // dev/prod
	LogLevel string       `yaml:"log_level"` // info/debug
	Timeout time.Duration `yaml:"timeout"`   // "3s"
}

type DBConfig struct {
	DSN            string        `yaml:"dsn"`
	MaxOpenConns   int           `yaml:"max_open_conns"`
	MaxIdleConns   int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type ChatConfig struct {
	ListenAddr string `yaml:"listen_addr"` // ":9000"
	MaxRooms   int    `yaml:"max_rooms"`
	Rooms	  []RoomConfig `yaml:"rooms"`
}

type RoomConfig struct {
	Name        string `yaml:"name"`
	Capacity    int    `yaml:"capacity"`
	Password    string `yaml:"password"`
}

func LoadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// defaulty
	cfg := Config{
		App: AppConfig{
			Name:     "my-app",
			Env:      "dev",
			LogLevel: "info",
			Timeout:  3 * time.Second,
		},
		Chat: ChatConfig{
			ListenAddr: ":9000",
			MaxRooms:   100,
			Rooms:    []RoomConfig{},
		},
	}

	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	// ENV override příklad (třeba tajnosti)
	if v := os.Getenv("DB_DSN"); v != "" {
		cfg.DB.DSN = v
	}

	// if err := cfg.Validate(); err != nil {
	// 	return nil, err
	// }
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.DB.DSN == "" {
		return errors.New("db.dsn is required")
	}
	if c.Chat.ListenAddr == "" {
		return errors.New("chat.listen_addr is required")
	}
	return nil
}

func (c *Config) Certificate() (tls.Certificate, error) {
    cert, err := tls.LoadX509KeyPair(c.Transport.CertFile, c.Transport.KeyFile)
	if err != nil {
		fmt.Println(err)
	}

    return cert, nil
}