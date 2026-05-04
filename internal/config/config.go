package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/alliso/keymapper/internal/gamepad"
	"github.com/alliso/keymapper/internal/keymap"
)

// Keys is a sequence of key names to emit in order. It accepts two YAML forms:
// a scalar (single key, e.g. `on_press: space`) or a sequence
// (e.g. `on_press: [g, g]`). The scalar form is preserved on marshal so the
// `learn` wizard keeps producing the historical single-key format.
type Keys []string

func (k *Keys) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		if node.Value == "" {
			*k = nil
			return nil
		}
		*k = Keys{node.Value}
		return nil
	case yaml.SequenceNode:
		var items []string
		if err := node.Decode(&items); err != nil {
			return err
		}
		*k = Keys(items)
		return nil
	default:
		return fmt.Errorf("on_press/on_release debe ser una tecla o una lista de teclas")
	}
}

func (k Keys) MarshalYAML() (any, error) {
	if len(k) == 1 {
		return k[0], nil
	}
	return []string(k), nil
}

// Binding holds the keys to emit on each edge of a gamepad button.
// Both fields are optional: an empty field means "no key on this edge".
// Each field may be a single key or a sequence emitted in order.
type Binding struct {
	OnPress   Keys `yaml:"on_press,omitempty"`
	OnRelease Keys `yaml:"on_release,omitempty"`
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
		if len(b.OnPress) == 0 && len(b.OnRelease) == 0 {
			return fmt.Errorf("mappings[%s]: define on_press, on_release, o ambos", btnName)
		}
		for i, name := range b.OnPress {
			if _, err := keymap.Resolve(name); err != nil {
				return fmt.Errorf("mappings[%s].on_press[%d]: %w", btnName, i, err)
			}
		}
		for i, name := range b.OnRelease {
			if _, err := keymap.Resolve(name); err != nil {
				return fmt.Errorf("mappings[%s].on_release[%d]: %w", btnName, i, err)
			}
		}
	}
	return nil
}
