package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.Duration.String()), nil
}

type Settings struct {
	DefaultDuration    Duration `yaml:"default_duration"`
	MaxUnblockDuration Duration `yaml:"max_unblock_duration,omitempty"`
	CheckInterval      Duration `yaml:"check_interval"`
	FlushDNS           bool     `yaml:"flush_dns"`
	BlockSubdomains    bool     `yaml:"block_subdomains"`
	UnblockWarnings    []string `yaml:"unblock_warnings,omitempty"`
}

type Config struct {
	Domains  []string `yaml:"domains"`
	Settings Settings `yaml:"settings"`
}

func Default() *Config {
	return &Config{
		Domains: []string{},
		Settings: Settings{
			DefaultDuration: Duration{15 * time.Minute},
			CheckInterval:   Duration{5 * time.Second},
			FlushDNS:        true,
			BlockSubdomains: true,
			UnblockWarnings: []string{
				"You're about to unblock distracting sites.",
				"Consider whether this is truly necessary right now.",
			},
		},
	}
}

func ConfigDir() string  { return "/usr/local/etc/sc" }
func ConfigPath() string { return filepath.Join(ConfigDir(), "config.yaml") }
func DataDir() string    { return "/usr/local/var/sc" }
func SocketPath() string { return filepath.Join(DataDir(), "sc.sock") }
func StatePath() string  { return filepath.Join(DataDir(), "state.yaml") }
func LogsPath() string   { return filepath.Join(DataDir(), "logs.jsonl") }
func DaemonLog() string  { return filepath.Join(DataDir(), "daemon.log") }

func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := Save(cfg, path); err != nil {
				return nil, fmt.Errorf("creating default config: %w", err)
			}
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (c *Config) HasDomain(domain string) bool {
	domain = normalizeDomain(domain)
	for _, d := range c.Domains {
		if d == domain {
			return true
		}
	}
	return false
}

func (c *Config) AddDomain(domain string) bool {
	domain = normalizeDomain(domain)
	if c.HasDomain(domain) {
		return false
	}
	c.Domains = append(c.Domains, domain)
	return true
}

func (c *Config) RemoveDomain(domain string) bool {
	domain = normalizeDomain(domain)
	for i, d := range c.Domains {
		if d == domain {
			c.Domains = append(c.Domains[:i], c.Domains[i+1:]...)
			return true
		}
	}
	return false
}

func normalizeDomain(d string) string {
	return strings.ToLower(strings.TrimSpace(d))
}
