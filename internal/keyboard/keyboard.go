package keyboard

import (
	"fmt"
	"time"

	keybd "github.com/micmonay/keybd_event"
)

// Keyboard simulates keystrokes on the host.
type Keyboard struct {
	kb      keybd.KeyBonding
	holdFor time.Duration
}

// New creates a Keyboard with the given tap hold duration (time between press
// and release). A typical value is 15ms.
func New(holdFor time.Duration) (*Keyboard, error) {
	kb, err := keybd.NewKeyBonding()
	if err != nil {
		return nil, fmt.Errorf("inicializando teclado: %w", err)
	}
	return &Keyboard{kb: kb, holdFor: holdFor}, nil
}

// Tap emits a single keycode as a press + release.
func (k *Keyboard) Tap(keycode int) error {
	k.kb.Clear()
	k.kb.SetKeys(keycode)
	if err := k.kb.Press(); err != nil {
		return fmt.Errorf("press: %w", err)
	}
	time.Sleep(k.holdFor)
	if err := k.kb.Release(); err != nil {
		return fmt.Errorf("release: %w", err)
	}
	return nil
}
