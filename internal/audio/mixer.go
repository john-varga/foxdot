package audio

// Mixer crossfades a set of streamed (music/ambience) tracks toward target
// volume weights in [0, 1]. It's the one piece of crossfade logic shared by
// both the ambience mixer (many loops blended at once, e.g. birds + wind)
// and the music director (exactly one track active, with the old one fading
// out while the new one fades in) — "play one track at full weight" is just
// a Mixer.SetWeights call with a single key.
type Mixer struct {
	engine *Engine

	target  map[string]float32
	current map[string]float32

	// FadeSeconds is how long a full 0→1 (or 1→0) fade takes.
	FadeSeconds float32
}

// NewMixer creates a Mixer with a sensible default fade time.
func NewMixer(e *Engine, fadeSeconds float32) *Mixer {
	if fadeSeconds <= 0 {
		fadeSeconds = 1
	}
	return &Mixer{
		engine:      e,
		target:      make(map[string]float32),
		current:     make(map[string]float32),
		FadeSeconds: fadeSeconds,
	}
}

// SetWeights replaces the desired mix. Tracks present in weights will be
// started (if not already playing) and faded toward their weight; tracks
// no longer present fade out and are stopped once inaudible.
func (m *Mixer) SetWeights(weights map[string]float32) {
	m.target = weights
}

// Update advances every track's current volume toward its target by dt
// seconds worth of fade, starting/stopping the underlying loops as needed.
func (m *Mixer) Update(dt float32) {
	step := dt / m.FadeSeconds
	if step > 1 {
		step = 1
	}

	for name, target := range m.target {
		cur := m.current[name]
		if cur == 0 && target > 0 {
			if err := m.engine.PlayLoop(name); err != nil {
				continue
			}
		}
		cur = approach(cur, clamp01(target), step)
		m.current[name] = cur
		m.engine.SetLoopVolume(name, cur)
	}

	for name, cur := range m.current {
		if _, wanted := m.target[name]; wanted {
			continue
		}
		cur = approach(cur, 0, step)
		if cur <= 0.001 {
			m.engine.StopLoop(name)
			delete(m.current, name)
			continue
		}
		m.current[name] = cur
		m.engine.SetLoopVolume(name, cur)
	}
}

// Weight returns the current (post-fade) volume of a track, mostly useful
// for tests/debugging.
func (m *Mixer) Weight(name string) float32 {
	return m.current[name]
}

func approach(current, target, step float32) float32 {
	if step >= 1 {
		return target
	}
	diff := target - current
	return current + diff*step
}
