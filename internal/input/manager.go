package input

// Source produces an input Frame for the current tick given the active
// Config. Implementations should be cheap to call every frame.
type Source interface {
	Poll(cfg Config) Frame
}

// Manager polls every registered Source once per tick and merges the
// results into a single Frame, so gameplay code never has to care whether
// the player is on a keyboard, a gamepad, or switching between both
// mid-session (common on Steam Deck-style setups).
type Manager struct {
	cfg     Config
	sources []Source
}

// NewManager creates a Manager with the given config and input sources.
// Sources are polled in order; see Frame.Merge for how conflicting axis
// values are resolved.
func NewManager(cfg Config, sources ...Source) *Manager {
	return &Manager{cfg: cfg, sources: sources}
}

// Config returns the manager's current input configuration.
func (m *Manager) Config() Config {
	return m.cfg
}

// SetConfig replaces the input configuration (e.g. after the player rebinds
// a key or loads a new config file).
func (m *Manager) SetConfig(cfg Config) {
	m.cfg = cfg
}

// Poll gathers a merged Frame from all sources for the current tick.
func (m *Manager) Poll() Frame {
	var result Frame
	for _, src := range m.sources {
		result = result.Merge(src.Poll(m.cfg))
	}
	return result
}
