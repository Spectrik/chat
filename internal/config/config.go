package config

import (
	"crypto/tls"
	"errors"
	"os"
	"time"

	"github.com/ondrej/chat/room"
	"github.com/ondrej/chat/room/policies"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Transport TransportConfig `yaml:"transport"`
	DB  DBConfig  `yaml:"db"`
	Chat ChatConfig `yaml:"chat"`
}

type TransportConfig struct {
	Type       string `yaml:"type"`        // "tcp" / "tls"
	Address	   string `yaml:"address"`     // "localhost:9000"
	CertFile   string `yaml:"cert_file"`   // pro TLS
	KeyFile    string `yaml:"key_file"`    // pro TLS
}

type DBConfig struct {
	DSN            string        `yaml:"dsn"`
	MaxOpenConns   int           `yaml:"max_open_conns"`
	MaxIdleConns   int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type ChatConfig struct {
	MaxRooms   int    `yaml:"max_rooms"`
	Rooms	  []RoomConfig `yaml:"rooms"`
}

type RoomConfig struct {
	Name        string `yaml:"name"`
	Policies 	PoliciesConfig `yaml:"policies"`
}

type PoliciesConfig struct {
    Capacity  *int    `yaml:"capacity"`
    Password  *string `yaml:"password"`
    UserLevel *uint    `yaml:"userlevel"`
}

func LoadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// defaulty
	cfg := Config{
		Transport: TransportConfig{
			Type: "tcp",
			Address: "localhost:9000",
		},
		Chat: ChatConfig{
			MaxRooms:   100,
			Rooms:    []RoomConfig{},
		},
	}

	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

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
	if c.Transport.Address == "" {
		return errors.New("chat.listen_addr is required")
	}
	return nil
}

func (c *Config) Certificate() (tls.Certificate, error) {
    cert, err := tls.LoadX509KeyPair(c.Transport.CertFile, c.Transport.KeyFile)
	if err != nil {
		return tls.Certificate{}, err
	}

    return cert, nil
}

func BuildPolicies(cfg PoliciesConfig) (*room.PolicySet, error) {
	var pols room.PolicySet

    if cfg.Capacity != nil {
        pols.Join = append(pols.Join, policies.CapacityPolicy{
            Max: *cfg.Capacity,
        })
    }

    if cfg.Password != nil {
        pols.Join = append(pols.Join, policies.PasswordPolicy{
            Password: *cfg.Password,
        })
    }

    if cfg.UserLevel != nil {
        pols.Join = append(pols.Join, policies.MinLevelPolicy{
            Minlevel: *cfg.UserLevel,
        })
    }

    return &pols, nil
}