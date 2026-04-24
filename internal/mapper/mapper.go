package mapper

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/alliso/keymapper/internal/config"
	"github.com/alliso/keymapper/internal/gamepad"
	"github.com/alliso/keymapper/internal/keyboard"
	"github.com/alliso/keymapper/internal/keymap"
)

// Run is the main mapping loop. Must be called from the main OS thread because
// SDL event polling requires it (especially on macOS).
func Run(cfg *config.Config, tapHold time.Duration) error {
	if err := gamepad.Init(); err != nil {
		return err
	}
	defer gamepad.Quit()

	js, err := gamepad.Open(cfg.GamepadIndex)
	if err != nil {
		return err
	}
	defer js.Close()
	instanceID := js.InstanceID()

	fmt.Printf("Mando: %s (índice %d, %d botones)\n", js.Name(), cfg.GamepadIndex, js.NumButtons())

	onPress, onRelease, err := resolveTables(cfg, js.NumButtons())
	if err != nil {
		return err
	}

	kb, err := keyboard.New(tapHold)
	if err != nil {
		return err
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigc)

	fmt.Println("Mapeo activo. Ctrl+C para salir.")

	for {
		select {
		case <-sigc:
			fmt.Println("\nSaliendo…")
			return nil
		default:
		}

		ev := sdl.WaitEventTimeout(100)
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *sdl.JoyButtonEvent:
			if e.Which != instanceID {
				continue
			}
			var code int
			var ok bool
			switch e.Type {
			case sdl.JOYBUTTONDOWN:
				code, ok = onPress[e.Button]
			case sdl.JOYBUTTONUP:
				code, ok = onRelease[e.Button]
			}
			if !ok {
				continue
			}
			if err := kb.Tap(code); err != nil {
				fmt.Fprintf(os.Stderr, "tap %s falló: %v\n", gamepad.ButtonLabel(e.Button), err)
			}

		case *sdl.JoyDeviceRemovedEvent:
			if e.Which == instanceID {
				fmt.Println("\nMando desconectado. Saliendo.")
				return nil
			}

		case *sdl.QuitEvent:
			return nil
		}
	}
}

func resolveTables(cfg *config.Config, numButtons int) (onPress, onRelease map[uint8]int, err error) {
	onPress = make(map[uint8]int)
	onRelease = make(map[uint8]int)
	for btnName, b := range cfg.Mappings {
		idx, err := gamepad.ParseButtonName(btnName)
		if err != nil {
			return nil, nil, err
		}
		if int(idx) >= numButtons {
			return nil, nil, fmt.Errorf("mappings[%s]: el mando solo tiene %d botones (0..%d)", btnName, numButtons, numButtons-1)
		}
		if b.OnPress != "" {
			code, err := keymap.Resolve(b.OnPress)
			if err != nil {
				return nil, nil, fmt.Errorf("mappings[%s].on_press: %w", btnName, err)
			}
			onPress[idx] = code
		}
		if b.OnRelease != "" {
			code, err := keymap.Resolve(b.OnRelease)
			if err != nil {
				return nil, nil, fmt.Errorf("mappings[%s].on_release: %w", btnName, err)
			}
			onRelease[idx] = code
		}
	}
	return onPress, onRelease, nil
}
