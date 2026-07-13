package storage

import (
	"errors"
	"testing"
	"time"
)

func TestSaveLoadJSON(t *testing.T) {
	s := New(t.TempDir())

	type payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	in := payload{Name: "acorn", Count: 3}

	if err := s.SaveJSON("nested/dir/data.json", in); err != nil {
		t.Fatalf("SaveJSON: %v", err)
	}
	if !s.Exists("nested/dir/data.json") {
		t.Fatalf("expected file to exist after save")
	}

	var out payload
	if err := s.LoadJSON("nested/dir/data.json", &out); err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}
	if out != in {
		t.Fatalf("roundtrip mismatch: got %+v, want %+v", out, in)
	}
}

func TestLoadJSONMissing(t *testing.T) {
	s := New(t.TempDir())
	var out struct{}
	err := s.LoadJSON("does/not/exist.json", &out)
	if !errors.Is(err, ErrNotExist) {
		t.Fatalf("expected ErrNotExist, got %v", err)
	}
}

func TestPathEscapeRejected(t *testing.T) {
	s := New(t.TempDir())
	if _, err := s.Path("../escape.json"); err == nil {
		t.Fatalf("expected error for path escaping base dir")
	}
	if _, err := s.Path("/absolute.json"); err == nil {
		t.Fatalf("expected error for absolute path")
	}
}

func TestDeleteMissingIsNotError(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Delete("missing.json"); err != nil {
		t.Fatalf("Delete on missing file should be a no-op, got %v", err)
	}
}

func TestSaveGameSlotRoundTrip(t *testing.T) {
	s := New(t.TempDir())
	want := SaveGame{
		Version:      CurrentSaveVersion,
		SavedAt:      time.Now().UTC().Truncate(time.Second),
		FoxPosition:  Vec3{X: 1, Y: 2, Z: 3},
		FoxYaw:       0.5,
		CameraYaw:    1.2,
		CameraPitch:  -0.3,
		PlaytimeSecs: 42.5,
	}

	if err := s.SaveGameSlot("slot1", want); err != nil {
		t.Fatalf("SaveGameSlot: %v", err)
	}

	got, err := s.LoadGameSlot("slot1")
	if err != nil {
		t.Fatalf("LoadGameSlot: %v", err)
	}
	if !got.SavedAt.Equal(want.SavedAt) || got.FoxPosition != want.FoxPosition {
		t.Fatalf("roundtrip mismatch: got %+v, want %+v", got, want)
	}

	slots, err := s.ListGameSlots()
	if err != nil {
		t.Fatalf("ListGameSlots: %v", err)
	}
	if len(slots) != 1 || slots[0] != "slot1" {
		t.Fatalf("expected [slot1], got %v", slots)
	}

	if err := s.DeleteGameSlot("slot1"); err != nil {
		t.Fatalf("DeleteGameSlot: %v", err)
	}
	slots, err = s.ListGameSlots()
	if err != nil {
		t.Fatalf("ListGameSlots after delete: %v", err)
	}
	if len(slots) != 0 {
		t.Fatalf("expected no slots after delete, got %v", slots)
	}
}

func TestSaveGameSlotRejectsBadName(t *testing.T) {
	s := New(t.TempDir())
	if err := s.SaveGameSlot("../evil", SaveGame{}); err == nil {
		t.Fatalf("expected error for unsafe save name")
	}
}
