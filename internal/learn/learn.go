package learn

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/alliso/keymapper/internal/config"
	"github.com/alliso/keymapper/internal/gamepad"
)

// Run drives the interactive wizard and writes a config YAML to outputPath.
//
// Flow:
//  1. Open the joystick at gamepadIndex.
//  2. User presses a physical button on the gamepad; the wizard records its index.
//  3. User presses a keyboard key for on_press (ESC to skip).
//  4. User presses a keyboard key for on_release (ESC to skip).
//  5. Loop until Ctrl+C.
//
// SDL events and keyboard input are multiplexed in a single loop (main OS thread).
func Run(outputPath string, gamepadIndex int) error {
	if err := gamepad.Init(); err != nil {
		return err
	}
	defer gamepad.Quit()

	js, err := gamepad.Open(gamepadIndex)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el mando: %w", err)
	}
	defer js.Close()
	instanceID := js.InstanceID()

	fmt.Printf("Mando: %s (índice %d, %d botones detectados)\n\n",
		js.Name(), gamepadIndex, js.NumButtons())

	fmt.Println("Wizard de mapeo (modo raw).")
	fmt.Println("1) Pulsa un botón del mando — se registra su índice.")
	fmt.Println("2) Pulsa la tecla del teclado que quieres para on_press.")
	fmt.Println("3) Pulsa la tecla del teclado que quieres para on_release.")
	fmt.Println("Repite con otros botones. ESC=saltar flanco, Ctrl+C=terminar.")
	fmt.Println("Teclas soportadas: a-z, 0-9, space, enter, tab, up/down/left/right, f1..f12.")
	fmt.Println()

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return fmt.Errorf("el wizard necesita un terminal interactivo; no redirijas stdin ni ejecutes con pipe")
	}
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("habilitando raw mode: %w", err)
	}
	defer term.Restore(fd, oldState)

	rr := newRawReader(os.Stdin)
	cfg := &config.Config{
		GamepadIndex: gamepadIndex,
		Mappings:     map[string]config.Binding{},
	}

	type stateKind int
	const (
		stateWaitButton stateKind = iota
		stateWaitPressKey
		stateWaitReleaseKey
	)

	state := stateWaitButton
	var currentIdx uint8
	var currentBinding config.Binding

	promptWait := func() {
		fmt.Printf("\r\nPulsa un botón del mando (Ctrl+C para terminar):\r\n")
	}
	promptEdge := func(edge string) {
		fmt.Printf("  %s (pulsa tecla, ESC=saltar): ", edge)
	}
	finalize := func() {
		label := gamepad.ButtonLabel(currentIdx)
		if len(currentBinding.OnPress) == 0 && len(currentBinding.OnRelease) == 0 {
			fmt.Printf("  → %s saltado (sin teclas).\r\n", label)
		} else {
			cfg.Mappings[label] = currentBinding
			fmt.Printf("  → %s guardado: press=%q release=%q\r\n",
				label, strings.Join(currentBinding.OnPress, " "), strings.Join(currentBinding.OnRelease, " "))
		}
		currentBinding = config.Binding{}
		state = stateWaitButton
		promptWait()
	}

	promptWait()
	done := false

	for !done {
		// 1. Pump SDL event with short timeout.
		ev := sdl.WaitEventTimeout(50)
		if ev != nil {
			switch e := ev.(type) {
			case *sdl.JoyButtonEvent:
				if e.Which == instanceID && e.Type == sdl.JOYBUTTONDOWN && state == stateWaitButton {
					currentIdx = e.Button
					currentBinding = config.Binding{}
					fmt.Printf("  → botón detectado: %s\r\n", gamepad.ButtonLabel(currentIdx))
					promptEdge("on_press")
					state = stateWaitPressKey
				}
				// Ignore button events in other states.
			case *sdl.JoyDeviceRemovedEvent:
				if e.Which == instanceID {
					fmt.Printf("\r\n[mando desconectado]\r\n")
					done = true
				}
			}
		}

		// 2. Non-blocking stdin read.
		select {
		case b, ok := <-rr.ch:
			if !ok {
				done = true
				continue
			}
			// Ctrl+C ends the wizard at any state.
			if b == 0x03 {
				done = true
				continue
			}
			switch state {
			case stateWaitButton:
				// Ignore keyboard until a button is pressed.
			case stateWaitPressKey, stateWaitReleaseKey:
				key, act := decodeKey(b, rr)
				switch act {
				case actionQuit:
					done = true
				case actionAssign:
					fmt.Printf("%s\r\n", key)
					if state == stateWaitPressKey {
						currentBinding.OnPress = config.Keys{key}
						promptEdge("on_release")
						state = stateWaitReleaseKey
					} else {
						currentBinding.OnRelease = config.Keys{key}
						finalize()
					}
				case actionSkipEdge:
					fmt.Printf("[saltado]\r\n")
					if state == stateWaitPressKey {
						promptEdge("on_release")
						state = stateWaitReleaseKey
					} else {
						finalize()
					}
				case actionRetry:
					fmt.Printf("[tecla no soportada, reintenta]\r\n")
					if state == stateWaitPressKey {
						promptEdge("on_press")
					} else {
						promptEdge("on_release")
					}
				}
			}
		default:
		}
	}

	term.Restore(fd, oldState)
	fmt.Println()
	if len(cfg.Mappings) == 0 {
		return fmt.Errorf("wizard no generó ningún mapeo, no se escribe archivo")
	}
	if err := config.Save(outputPath, cfg); err != nil {
		return err
	}
	fmt.Printf("Guardado: %s (%d botones mapeados)\n", outputPath, len(cfg.Mappings))
	return nil
}

type action int

const (
	actionAssign action = iota
	actionSkipEdge
	actionRetry
	actionQuit
)

// rawReader pumps bytes from stdin into a channel so we can read with a
// timeout (needed to distinguish standalone ESC from escape sequences).
type rawReader struct {
	ch chan byte
}

func newRawReader(r io.Reader) *rawReader {
	rr := &rawReader{ch: make(chan byte, 64)}
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := r.Read(buf)
			if err != nil || n == 0 {
				close(rr.ch)
				return
			}
			rr.ch <- buf[0]
		}
	}()
	return rr
}

func (r *rawReader) readTimeout(d time.Duration) (byte, bool) {
	select {
	case b, ok := <-r.ch:
		return b, ok
	case <-time.After(d):
		return 0, false
	}
}

// decodeKey maps a byte (and possibly a follow-up escape sequence) into a
// canonical key name or a control action.
func decodeKey(b byte, rr *rawReader) (string, action) {
	switch {
	case b == 0x03: // Ctrl+C
		return "", actionQuit
	case b == 0x1B: // ESC, possibly start of an escape sequence
		next, ok := rr.readTimeout(30 * time.Millisecond)
		if !ok {
			return "", actionSkipEdge
		}
		return decodeEscSeq(next, rr)
	case b >= 'a' && b <= 'z':
		return string(b), actionAssign
	case b >= 'A' && b <= 'Z':
		return string(b + 32), actionAssign
	case b >= '0' && b <= '9':
		return string(b), actionAssign
	case b == ' ':
		return "space", actionAssign
	case b == '\t':
		return "tab", actionAssign
	case b == '\r', b == '\n':
		return "enter", actionAssign
	}
	return "", actionRetry
}

func decodeEscSeq(b byte, rr *rawReader) (string, action) {
	switch b {
	case '[':
		c, ok := rr.readTimeout(30 * time.Millisecond)
		if !ok {
			return "", actionRetry
		}
		switch c {
		case 'A':
			return "up", actionAssign
		case 'B':
			return "down", actionAssign
		case 'C':
			return "right", actionAssign
		case 'D':
			return "left", actionAssign
		}
		// Tilde-terminated sequence: CSI N(N)~ — F5..F12 on most terminals.
		buf := []byte{c}
		for {
			d, ok := rr.readTimeout(30 * time.Millisecond)
			if !ok {
				return "", actionRetry
			}
			if d == '~' {
				break
			}
			buf = append(buf, d)
			if len(buf) > 4 {
				return "", actionRetry
			}
		}
		return tildeToFKey(string(buf))
	case 'O':
		// SS3 sequence used by macOS Terminal.app for F1-F4.
		c, ok := rr.readTimeout(30 * time.Millisecond)
		if !ok {
			return "", actionRetry
		}
		switch c {
		case 'P':
			return "f1", actionAssign
		case 'Q':
			return "f2", actionAssign
		case 'R':
			return "f3", actionAssign
		case 'S':
			return "f4", actionAssign
		}
		return "", actionRetry
	}
	return "", actionRetry
}

func tildeToFKey(s string) (string, action) {
	switch s {
	case "11":
		return "f1", actionAssign
	case "12":
		return "f2", actionAssign
	case "13":
		return "f3", actionAssign
	case "14":
		return "f4", actionAssign
	case "15":
		return "f5", actionAssign
	case "17":
		return "f6", actionAssign
	case "18":
		return "f7", actionAssign
	case "19":
		return "f8", actionAssign
	case "20":
		return "f9", actionAssign
	case "21":
		return "f10", actionAssign
	case "23":
		return "f11", actionAssign
	case "24":
		return "f12", actionAssign
	}
	return "", actionRetry
}
