package audio

import "foxdot/internal/worldtime"

// AmbienceWeights decides the ambience loop mix for the current time of day
// plus how close the listener is to a campfire (0 = right on top of it, 1+
// = far enough that it shouldn't be heard at all). This is pure/testable —
// deciding *what* to play is separate from *how* it's mixed (Mixer) or
// played back (Engine), matching how internal/sim decides fox behaviour
// without touching raylib.
//
// The mix is intentionally simple: birds and gentle wind by day, crickets
// and frogs by night, wind and rustling leaves year-round as a bed, and
// campfire crackle fading in with proximity regardless of time of day.
func AmbienceWeights(phase worldtime.Phase, campfireProximity float32) map[string]float32 {
	w := map[string]float32{
		"wind":            0.35,
		"rustling_leaves": 0.3,
	}

	switch phase {
	case worldtime.Day:
		w["birds"] = 0.6
	case worldtime.Dawn:
		w["birds"] = 0.35
		w["crickets"] = 0.15
	case worldtime.Dusk:
		w["crickets"] = 0.25
		w["frogs"] = 0.15
	case worldtime.Night:
		w["crickets"] = 0.45
		w["frogs"] = 0.3
	}

	if fire := campfireLoopWeight(campfireProximity); fire > 0 {
		w["campfire"] = fire
	}

	return w
}

// campfireLoopWeight fades the campfire loop in linearly as proximity
// approaches 0 (right next to it) and out to silence by proximity 1.
func campfireLoopWeight(proximity float32) float32 {
	if proximity >= 1 {
		return 0
	}
	if proximity <= 0 {
		return 0.8
	}
	return 0.8 * (1 - proximity)
}

// AmbiencePlayer combines AmbienceWeights with a Mixer so game code just
// reports the current phase and campfire proximity each frame.
type AmbiencePlayer struct {
	mixer *Mixer
}

// NewAmbiencePlayer creates an AmbiencePlayer that crossfades loops over
// fadeSeconds using engine e.
func NewAmbiencePlayer(e *Engine, fadeSeconds float32) *AmbiencePlayer {
	return &AmbiencePlayer{mixer: NewMixer(e, fadeSeconds)}
}

// Update recomputes the target mix for the given phase/proximity and
// advances the crossfade.
func (p *AmbiencePlayer) Update(dt float32, phase worldtime.Phase, campfireProximity float32) {
	p.mixer.SetWeights(AmbienceWeights(phase, campfireProximity))
	p.mixer.Update(dt)
}
