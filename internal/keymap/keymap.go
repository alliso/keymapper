package keymap

import (
	"fmt"
	"sort"
	"strings"

	keybd "github.com/micmonay/keybd_event"
)

var nameToCode = map[string]int{
	"a": keybd.VK_A, "b": keybd.VK_B, "c": keybd.VK_C, "d": keybd.VK_D,
	"e": keybd.VK_E, "f": keybd.VK_F, "g": keybd.VK_G, "h": keybd.VK_H,
	"i": keybd.VK_I, "j": keybd.VK_J, "k": keybd.VK_K, "l": keybd.VK_L,
	"m": keybd.VK_M, "n": keybd.VK_N, "o": keybd.VK_O, "p": keybd.VK_P,
	"q": keybd.VK_Q, "r": keybd.VK_R, "s": keybd.VK_S, "t": keybd.VK_T,
	"u": keybd.VK_U, "v": keybd.VK_V, "w": keybd.VK_W, "x": keybd.VK_X,
	"y": keybd.VK_Y, "z": keybd.VK_Z,

	"0": keybd.VK_0, "1": keybd.VK_1, "2": keybd.VK_2, "3": keybd.VK_3,
	"4": keybd.VK_4, "5": keybd.VK_5, "6": keybd.VK_6, "7": keybd.VK_7,
	"8": keybd.VK_8, "9": keybd.VK_9,

	"space":  keybd.VK_SPACE,
	"enter":  keybd.VK_ENTER,
	"return": keybd.VK_ENTER,
	"tab":    keybd.VK_TAB,
	"escape": keybd.VK_ESC,
	"esc":    keybd.VK_ESC,

	"up":    keybd.VK_UP,
	"down":  keybd.VK_DOWN,
	"left":  keybd.VK_LEFT,
	"right": keybd.VK_RIGHT,

	"f1":  keybd.VK_F1,
	"f2":  keybd.VK_F2,
	"f3":  keybd.VK_F3,
	"f4":  keybd.VK_F4,
	"f5":  keybd.VK_F5,
	"f6":  keybd.VK_F6,
	"f7":  keybd.VK_F7,
	"f8":  keybd.VK_F8,
	"f9":  keybd.VK_F9,
	"f10": keybd.VK_F10,
	"f11": keybd.VK_F11,
	"f12": keybd.VK_F12,
}

// Resolve returns the keycode for a key name (case-insensitive).
func Resolve(name string) (int, error) {
	code, ok := nameToCode[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return 0, fmt.Errorf("tecla desconocida: %q", name)
	}
	return code, nil
}

// Names returns the sorted list of supported key names.
func Names() []string {
	names := make([]string, 0, len(nameToCode))
	for k := range nameToCode {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
