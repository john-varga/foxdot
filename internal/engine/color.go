package engine

import (
	"fmt"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// parseHexColor parses a "#RRGGBB" or "#RRGGBBAA" string into an rl.Color,
// falling back to opaque black if the string is malformed, so a typo in a
// config file never crashes the game.
func parseHexColor(hex string) rl.Color {
	c, err := tryParseHexColor(hex)
	if err != nil {
		return rl.Black
	}
	return c
}

func tryParseHexColor(hex string) (rl.Color, error) {
	h := strings.TrimPrefix(hex, "#")
	if len(h) != 6 && len(h) != 8 {
		return rl.Color{}, fmt.Errorf("engine: invalid color %q", hex)
	}
	v, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return rl.Color{}, fmt.Errorf("engine: invalid color %q: %w", hex, err)
	}

	if len(h) == 6 {
		return rl.Color{
			R: uint8(v >> 16),
			G: uint8(v >> 8),
			B: uint8(v),
			A: 255,
		}, nil
	}
	return rl.Color{
		R: uint8(v >> 24),
		G: uint8(v >> 16),
		B: uint8(v >> 8),
		A: uint8(v),
	}, nil
}
