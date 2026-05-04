package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestKeysUnmarshalScalar(t *testing.T) {
	var b Binding
	if err := yaml.Unmarshal([]byte("on_press: space\n"), &b); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(b.OnPress) != 1 || b.OnPress[0] != "space" {
		t.Fatalf("got %#v, want Keys{\"space\"}", b.OnPress)
	}
	if len(b.OnRelease) != 0 {
		t.Fatalf("OnRelease should be empty, got %#v", b.OnRelease)
	}
}

func TestKeysUnmarshalSequence(t *testing.T) {
	var b Binding
	if err := yaml.Unmarshal([]byte("on_press: [a, b, c]\n"), &b); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	want := []string{"a", "b", "c"}
	if len(b.OnPress) != len(want) {
		t.Fatalf("got %#v, want %#v", b.OnPress, want)
	}
	for i, v := range want {
		if b.OnPress[i] != v {
			t.Fatalf("idx %d: got %q want %q", i, b.OnPress[i], v)
		}
	}
}

func TestKeysMarshalRoundtrip(t *testing.T) {
	cases := []struct {
		name string
		in   Binding
		want string
	}{
		{
			name: "single key marshals as scalar",
			in:   Binding{OnPress: Keys{"space"}},
			want: "on_press: space\n",
		},
		{
			name: "sequence marshals as flow list",
			in:   Binding{OnPress: Keys{"a", "b"}},
			want: "on_press:\n    - a\n    - b\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := yaml.Marshal(tc.in)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(out) != tc.want {
				t.Fatalf("got %q want %q", string(out), tc.want)
			}
		})
	}
}

func TestValidateAcceptsSequence(t *testing.T) {
	cfg := &Config{
		GamepadIndex: 0,
		Mappings: map[string]Binding{
			"b0": {OnPress: Keys{"a", "b", "enter"}},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestValidateRejectsUnknownKeyInSequence(t *testing.T) {
	cfg := &Config{
		GamepadIndex: 0,
		Mappings: map[string]Binding{
			"b0": {OnPress: Keys{"a", "foo"}},
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "on_press[1]") {
		t.Fatalf("error should reference on_press[1], got: %v", err)
	}
}
