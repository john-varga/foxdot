package engine

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestParseHexColorRGB(t *testing.T) {
	got := parseHexColor("#8FD3F4")
	want := rl.Color{R: 0x8F, G: 0xD3, B: 0xF4, A: 255}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParseHexColorRGBA(t *testing.T) {
	got := parseHexColor("#11223344")
	want := rl.Color{R: 0x11, G: 0x22, B: 0x33, A: 0x44}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParseHexColorFallsBackOnInvalid(t *testing.T) {
	got := parseHexColor("not-a-color")
	if got != rl.Black {
		t.Fatalf("expected fallback to black, got %+v", got)
	}
}
