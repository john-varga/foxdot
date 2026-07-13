package engine

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"

	"foxdot/internal/config"
	"foxdot/internal/input"
)

// App owns the raylib window lifecycle and drives a Game every frame: poll
// input, Update, clear, Draw, present. It's intentionally small — nearly
// everything interesting lives in the Game implementation and the packages
// it composes (sim, camera, world, input).
type App struct {
	cfg      config.Config
	inputMgr *input.Manager
	game     Game
}

// NewApp wires together a config, an input manager (already configured with
// whichever Sources you want — keyboard, gamepad, or both) and a Game.
func NewApp(cfg config.Config, inputMgr *input.Manager, game Game) *App {
	return &App{cfg: cfg, inputMgr: inputMgr, game: game}
}

// Run opens the window, initializes the game, and blocks running the main
// loop until the window is closed. It always tears the window and game down
// cleanly on the way out, even on error.
func (a *App) Run() (err error) {
	rl.SetConfigFlags(a.windowFlags())
	rl.InitWindow(a.cfg.Window.Width, a.cfg.Window.Height, a.cfg.Window.Title)
	defer rl.CloseWindow()

	if a.cfg.Window.Fullscreen {
		rl.ToggleFullscreen()
	}
	rl.SetTargetFPS(a.cfg.Window.TargetFPS)
	// Disable raylib's built-in "Escape closes the window" behavior; we
	// surface Escape to the game as the Pause action instead and quit
	// explicitly (see the loop below) so a real pause menu can intercept it
	// later without the window yanking itself shut first.
	rl.SetExitKey(rl.KeyNull)

	if initErr := a.game.Init(); initErr != nil {
		return fmt.Errorf("engine: init game: %w", initErr)
	}
	defer a.game.Shutdown()

	bg := parseHexColor(a.cfg.Graphics.BackgroundColorHex)

	for !rl.WindowShouldClose() {
		dt := rl.GetFrameTime()
		frame := a.inputMgr.Poll()

		// TODO: once a real pause menu exists, this should toggle it
		// instead of quitting outright.
		if frame.Pause.Pressed {
			break
		}

		if updateErr := a.game.Update(dt, frame); updateErr != nil {
			return fmt.Errorf("engine: update: %w", updateErr)
		}

		rl.BeginDrawing()
		rl.ClearBackground(bg)
		a.game.Draw()
		rl.EndDrawing()
	}
	return nil
}

func (a *App) windowFlags() uint32 {
	var flags uint32
	if a.cfg.Window.MSAA4x {
		flags |= rl.FlagMsaa4xHint
	}
	if a.cfg.Window.Resizable {
		flags |= rl.FlagWindowResizable
	}
	if a.cfg.Window.VSync {
		flags |= rl.FlagVsyncHint
	}
	return flags
}
