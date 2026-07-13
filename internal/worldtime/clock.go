// Package worldtime is a small, pure (no raylib) day/night clock for the
// simulation. It tracks time as a fraction of a configurable-length day and
// exposes a Phase (Night/Dawn/Day/Dusk) plus a continuous sun-height value,
// so both gameplay (creature schedules) and the audio subsystem (ambience
// mix, music selection) can react to time of day without depending on each
// other.
package worldtime

import "math"

// Phase buckets TimeOfDay into a handful of named ranges.
type Phase int

const (
	Night Phase = iota
	Dawn
	Day
	Dusk
)

func (p Phase) String() string {
	switch p {
	case Night:
		return "night"
	case Dawn:
		return "dawn"
	case Day:
		return "day"
	case Dusk:
		return "dusk"
	default:
		return "unknown"
	}
}

// Settings configures the length of a day and where dawn/dusk fall within
// it, all as fractions of a full day in [0, 1). Zero value is invalid; use
// DefaultSettings.
type Settings struct {
	// DaySeconds is how many real seconds a full day/night cycle takes.
	DaySeconds float64
	// DawnStart/DawnEnd and DuskStart/DuskEnd mark the transition windows.
	// Night covers everything outside [DawnStart, DuskEnd).
	DawnStart, DawnEnd float64
	DuskStart, DuskEnd float64
}

// DefaultSettings gives a 6-minute day/night cycle: a short dawn and dusk
// and roughly equal day/night, tuned for a small demo scene rather than
// realism.
func DefaultSettings() Settings {
	return Settings{
		DaySeconds: 360,
		DawnStart:  0.22,
		DawnEnd:    0.28,
		DuskStart:  0.72,
		DuskEnd:    0.78,
	}
}

// Clock is a resource (see internal/ecs) tracking elapsed time and deriving
// time-of-day/phase/sun-height from it. All methods are pure and safe to
// unit test without a window or a real frame clock.
type Clock struct {
	Settings Settings
	Elapsed  float64
	Day      int
}

// NewClock creates a Clock starting at midnight (TimeOfDay 0) on day 0.
func NewClock(s Settings) *Clock {
	return &Clock{Settings: s}
}

// Advance moves the clock forward by dt seconds, rolling Day over whenever
// it crosses midnight.
func (c *Clock) Advance(dt float64) {
	if dt <= 0 || c.Settings.DaySeconds <= 0 {
		return
	}
	c.Elapsed += dt
	for c.Elapsed >= c.Settings.DaySeconds {
		c.Elapsed -= c.Settings.DaySeconds
		c.Day++
	}
}

// TimeOfDay returns the current time as a fraction of a day in [0, 1), where
// 0 is midnight and 0.5 is noon.
func (c *Clock) TimeOfDay() float64 {
	if c.Settings.DaySeconds <= 0 {
		return 0
	}
	return c.Elapsed / c.Settings.DaySeconds
}

// SetTimeOfDay jumps directly to a given fraction of the day (for saves,
// tests, or debug skip-to-time controls). frac is wrapped into [0, 1).
func (c *Clock) SetTimeOfDay(frac float64) {
	frac = math.Mod(frac, 1)
	if frac < 0 {
		frac++
	}
	c.Elapsed = frac * c.Settings.DaySeconds
}

// Phase buckets the current TimeOfDay into Night/Dawn/Day/Dusk per Settings.
func (c *Clock) Phase() Phase {
	t := c.TimeOfDay()
	s := c.Settings
	switch {
	case t >= s.DawnStart && t < s.DawnEnd:
		return Dawn
	case t >= s.DawnEnd && t < s.DuskStart:
		return Day
	case t >= s.DuskStart && t < s.DuskEnd:
		return Dusk
	default:
		return Night
	}
}

// IsNight reports whether the current phase is Night.
func (c *Clock) IsNight() bool { return c.Phase() == Night }

// IsDay reports whether the current phase is Day.
func (c *Clock) IsDay() bool { return c.Phase() == Day }

// SunHeight returns a smooth value in [-1, 1] for the sun's height above/
// below the horizon: -1 at midnight, +1 at noon, 0 near dawn/dusk. It's
// continuous (unlike Phase) so lighting or ambience volume can fade smoothly
// instead of snapping at phase boundaries.
func (c *Clock) SunHeight() float64 {
	return -math.Cos(2 * math.Pi * c.TimeOfDay())
}
