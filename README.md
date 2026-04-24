# keymapper

Aplicación de consola en Go que mapea botones de un mando a pulsaciones del teclado. Estilo *sv-mapper*, con soporte macOS y Windows.

Cada botón del mando define dos acciones independientes:

- `on_press`: tecla que se emite al **pulsar** el botón (opcional).
- `on_release`: tecla que se emite al **soltar** el botón (opcional).

Esto permite comportamiento "palanca" (misma tecla en ambos flancos), teclas distintas por flanco, o un único flanco activo.

## Requisitos

- Go 1.25+.
- **SDL2** instalada en el sistema.
  - macOS: `brew install sdl2`
  - Windows: copiar `SDL2.dll` junto al binario (descarga oficial: <https://libsdl.org/>).
- Permisos:
  - **macOS**: simular pulsaciones de teclado requiere conceder *Accesibilidad* a la app que ejecuta `keymapper` (Terminal, iTerm, etc.) en *Ajustes del Sistema → Privacidad y seguridad → Accesibilidad*.
  - **Windows**: normalmente no hace falta elevar, salvo que el juego destino capture el teclado a bajo nivel.

## Instalación

```bash
git clone https://github.com/alliso/keymapper.git
cd keymapper
go build ./...
```

## Uso

```bash
# Listar los mandos detectados
./keymapper list

# Generar una config interactivamente (raw mode: pulsa la tecla física)
./keymapper learn --output config.yaml
#   ESC     = saltar el flanco actual
#   Ctrl+C  = terminar el wizard
# Para mapear a la tecla ESC, edita el YAML a mano tras generar la base.

# Arrancar el mapeo con una config existente
./keymapper run --config config.yaml
```

Parar con `Ctrl+C` o desconectando el mando.

### Formato de la config

Ver [`config.example.yaml`](config.example.yaml). Resumen:

```yaml
gamepad_index: 0
mappings:
  b0:
    on_press: space
    on_release: space
  b1:
    on_press: enter
  b2:
    on_press: e
    on_release: r
```

Los botones se referencian por **índice físico** `bN` (tal como los publica el firmware del mando / HID). Esto funciona con cualquier joystick HID, incluidos Arduinos custom y mandos sin mapeo estándar en SDL. Usa el wizard `learn` para descubrir qué índice corresponde a cada botón físico.

Nombres de tecla soportados: `a`-`z`, `0`-`9`, `space`, `enter`/`return`, `tab`, `escape`/`esc`, `up`, `down`, `left`, `right`, `f1`..`f12`.

## Flags

- `run --config PATH` (default `config.yaml`)
- `run --tap-ms N` duración en ms entre `press` y `release` al simular una tecla (default 15)
- `learn --output PATH` (default `config.yaml`)
- `learn --gamepad N` índice del mando (solo informativo en el wizard)

## Licencia

MIT. Ver [`LICENSE`](LICENSE).
