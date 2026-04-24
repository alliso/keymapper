package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/alliso/keymapper/internal/gamepad"
	"github.com/alliso/keymapper/internal/keymap"
)

// Binding holds the keys to emit on each edge of a gamepad button.
// Both fields are optional: an empty field means "no key on this edge".
type Binding struct {
	OnPress   string `yaml:"on_press,omitempty"`
	OnRelease string `yaml:"on_release,omitempty"`
}

// Config is the full YAML schema.
type Config struct {
	GamepadIndex int                `yaml:"gamepad_index"`
	Mappings     map[string]Binding `yaml:"mappings"`
}

// Load reads and validates a YAML config.
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("leyendo %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parseando YAML %s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config as YAML.
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("serializando YAML: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("escribiendo %s: %w", path, err)
	}
	return nil
}

// Validate checks that every referenced button and key is recognized.
func (c *Config) Validate() error {
	if c.GamepadIndex < 0 {
		return fmt.Errorf("gamepad_index debe ser >= 0 (leído %d)", c.GamepadIndex)
	}
	if len(c.Mappings) == 0 {
		return fmt.Errorf("config sin 'mappings': define al menos un botón")
	}
	for btnName, b := range c.Mappings {
		if _, err := gamepad.ParseButtonName(btnName); err != nil {
			return fmt.Errorf("mappings[%s]: %w", btnName, err)
		}
		if b.OnPress == "" && b.OnRelease == "" {
			return fmt.Errorf("mappings[%s]: define on_press, on_release, o ambos", btnName)
		}
		if b.OnPress != "" {
			if _, err := keymap.Resolve(b.OnPress); err != nil {
				return fmt.Errorf("mappings[%s].on_press: %w", btnName, err)
			}
		}
		if b.OnRelease != "" {
			if _, err := keymap.Resolve(b.OnRelease); err != nil {
				return fmt.Errorf("mappings[%s].on_release: %w", btnName, err)
			}
		}
	}
	return nil
}
