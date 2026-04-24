package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/alliso/keymapper/internal/config"
	"github.com/alliso/keymapper/internal/gamepad"
	"github.com/alliso/keymapper/internal/mapper"
)

func init() {
	// SDL event polling requires the main OS thread, especially on macOS.
	runtime.LockOSThread()
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case "run":
		os.Exit(cmdRun(args))
	case "list":
		os.Exit(cmdList(args))
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "comando desconocido: %q\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `keymapper — mapea botones de un mando a teclas del teclado.

Uso:
  keymapper run  [--config config.yaml] [--tap-ms 15]
  keymapper list

Comandos:
  run    Carga el YAML y arranca el loop de mapeo (Ctrl+C para salir).
  list   Lista los joysticks detectados por SDL (índice, nombre, GUID, botones).
`)
}

func cmdRun(args []string) int {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	cfgPath := fs.String("config", "config.yaml", "Ruta al YAML de mapeos.")
	tapMs := fs.Int("tap-ms", 15, "Duración en ms entre press y release al simular una tecla.")
	_ = fs.Parse(args)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if err := mapper.Run(cfg, time.Duration(*tapMs)*time.Millisecond); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}

func cmdList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	_ = fs.Parse(args)

	if err := gamepad.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	defer gamepad.Quit()

	infos := gamepad.List()
	if len(infos) == 0 {
		fmt.Println("No se detectó ningún joystick ni mando.")
		return 0
	}
	for _, info := range infos {
		gcTag := ""
		if info.IsGameController {
			gcTag = " [GameController]"
		}
		fmt.Printf("[%d] %s%s\n", info.Index, info.Name, gcTag)
		fmt.Printf("    GUID=%s  botones=%d  ejes=%d  hats=%d\n",
			info.GUID, info.Buttons, info.Axes, info.Hats)
	}
	return 0
}
