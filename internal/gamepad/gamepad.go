package gamepad

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/veandco/go-sdl2/sdl"
)

// Info describes a detected joystick/gamepad.
type Info struct {
	Index           int
	Name            string
	GUID            string
	Buttons         int
	Axes            int
	Hats            int
	IsGameController bool
}

// Init initializes the SDL subsystems needed to read joysticks.
// Must be called from the main OS thread.
func Init() error {
	if err := sdl.Init(sdl.INIT_JOYSTICK | sdl.INIT_EVENTS); err != nil {
		return fmt.Errorf("SDL init: %w", err)
	}
	return nil
}

// Quit releases SDL resources.
func Quit() {
	sdl.Quit()
}

// List returns all connected joystick-class devices.
func List() []Info {
	n := sdl.NumJoysticks()
	out := make([]Info, 0, n)
	for i := 0; i < n; i++ {
		info := Info{
			Index:            i,
			Name:             sdl.JoystickNameForIndex(i),
			GUID:             sdl.JoystickGetGUIDString(sdl.JoystickGetDeviceGUID(i)),
			IsGameController: sdl.IsGameController(i),
		}
		if js := sdl.JoystickOpen(i); js != nil {
			info.Buttons = js.NumButtons()
			info.Axes = js.NumAxes()
			info.Hats = js.NumHats()
			js.Close()
		}
		out = append(out, info)
	}
	return out
}

// Open opens the joystick at the given index. Caller must Close it.
func Open(index int) (*sdl.Joystick, error) {
	n := sdl.NumJoysticks()
	if index < 0 || index >= n {
		return nil, fmt.Errorf("gamepad_index %d fuera de rango (detectados %d dispositivos)", index, n)
	}
	js := sdl.JoystickOpen(index)
	if js == nil {
		return nil, fmt.Errorf("no se pudo abrir el joystick %d: %s", index, sdl.GetError())
	}
	return js, nil
}

var buttonNameRE = regexp.MustCompile(`^b(\d+)$`)

// ParseButtonName turns an "bN" label into a zero-based button index.
func ParseButtonName(name string) (uint8, error) {
	m := buttonNameRE.FindStringSubmatch(name)
	if m == nil {
		return 0, fmt.Errorf("nombre de botón inválido %q (usa formato bN, p.ej. b0, b1, b2)", name)
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n < 0 || n > 255 {
		return 0, fmt.Errorf("índice de botón fuera de rango en %q (máx 255)", name)
	}
	return uint8(n), nil
}

// ButtonLabel is the inverse of ParseButtonName.
func ButtonLabel(idx uint8) string {
	return fmt.Sprintf("b%d", idx)
}
