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
			var codes []int
			switch e.Type {
			case sdl.JOYBUTTONDOWN:
				codes = onPress[e.Button]
			case sdl.JOYBUTTONUP:
				codes = onRelease[e.Button]
			}
			if len(codes) == 0 {
				continue
			}
			if err := kb.TapSequence(codes); err != nil {
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

func resolveTables(cfg *config.Config, numButtons int) (onPress, onRelease map[uint8][]int, err error) {
	onPress = make(map[uint8][]int)
	onRelease = make(map[uint8][]int)
	resolveSeq := func(btnName, edge string, names []string) ([]int, error) {
		codes := make([]int, len(names))
		for i, name := range names {
			code, err := keymap.Resolve(name)
			if err != nil {
				return nil, fmt.Errorf("mappings[%s].%s[%d]: %w", btnName, edge, i, err)
			}
			codes[i] = code
		}
		return codes, nil
	}
	for btnName, b := range cfg.Mappings {
		idx, err := gamepad.ParseButtonName(btnName)
		if err != nil {
			return nil, nil, err
		}
		if int(idx) >= numButtons {
			return nil, nil, fmt.Errorf("mappings[%s]: el mando solo tiene %d botones (0..%d)", btnName, numButtons, numButtons-1)
		}
		if len(b.OnPress) > 0 {
			codes, err := resolveSeq(btnName, "on_press", b.OnPress)
			if err != nil {
				return nil, nil, err
			}
			onPress[idx] = codes
		}
		if len(b.OnRelease) > 0 {
			codes, err := resolveSeq(btnName, "on_release", b.OnRelease)
			if err != nil {
				return nil, nil, err
			}
			onRelease[idx] = codes
		}
	}
	return onPress, onRelease, nil
}
