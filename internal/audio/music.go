package audio

import "foxdot/internal/worldtime"

// Situation is the non-time-of-day context that can override which music
// plays — set by gameplay code (e.g. a "danger" state once predators exist,
// or a menu screen once one exists). Defaults to Explore, which just
// follows time of day.
type Situation int

const (
	SituationExplore Situation = iota
	SituationMainMenu
	SituationDiscovery
	SituationDanger
	SituationCredits
)

// ChooseMusicTrack picks a catalog track name for the given situation and
// time-of-day phase. Pure and testable: the *decision* of what should be
// playing is independent of how it's actually crossfaded in (Mixer) or
// played back (Engine).
func ChooseMusicTrack(situation Situation, phase worldtime.Phase) string {
	switch situation {
	case SituationMainMenu:
		return "main_menu"
	case SituationDiscovery:
		return "discovery"
	case SituationDanger:
		return "quiet_danger"
	case SituationCredits:
		return "credits"
	default: // SituationExplore
		if phase == worldtime.Night {
			return "forest_night"
		}
		return "forest_day"
	}
}

// MusicDirector picks the desired music track from Situation + time of day
// and crossfades to it via a Mixer whenever the desired track changes.
type MusicDirector struct {
	mixer     *Mixer
	Situation Situation

	current string
}

// NewMusicDirector creates a MusicDirector that crossfades over fadeSeconds
// using engine e.
func NewMusicDirector(e *Engine, fadeSeconds float32) *MusicDirector {
	return &MusicDirector{mixer: NewMixer(e, fadeSeconds)}
}

// Update re-evaluates the desired track for the current phase/situation and
// advances the crossfade.
func (d *MusicDirector) Update(dt float32, phase worldtime.Phase) {
	desired := ChooseMusicTrack(d.Situation, phase)
	if desired != d.current {
		d.current = desired
		d.mixer.SetWeights(map[string]float32{desired: 1})
	}
	d.mixer.Update(dt)
}

// Current returns the name of the currently desired (fading-in-or-playing)
// track.
func (d *MusicDirector) Current() string {
	return d.current
}
