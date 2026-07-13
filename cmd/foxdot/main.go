// Command foxdot launches the FoxDot demo: a stylized, low-poly forest you
// explore in third person as a fox. This is the concrete wiring point —
// everything interesting lives in the internal packages, all of which can
// be unit tested independently of this file and of raylib's window/GPU
// context.
package main

import (
	"fmt"
	"os"

	"foxdot/internal/config"
	"foxdot/internal/engine"
	"foxdot/internal/game"
	"foxdot/internal/input"
	"foxdot/internal/storage"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "foxdot:", err)
		os.Exit(1)
	}
}

func run() error {
	store, err := storage.Default()
	if err != nil {
		return fmt.Errorf("resolve storage location: %w", err)
	}

	cfg, err := config.Load(store)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	inputMgr := input.NewManager(cfg.Input,
		input.NewKeyboardMouseSource(),
		input.NewGamepadSource(),
	)

	scene := game.New(cfg, store)
	app := engine.NewApp(cfg, inputMgr, scene)
	return app.Run()
}
