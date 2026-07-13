package worldtime

import (
	"math"
	"testing"
)

func TestAdvanceWrapsDay(t *testing.T) {
	c := NewClock(Settings{DaySeconds: 100})
	c.Advance(150)
	if c.Day != 1 {
		t.Fatalf("expected Day=1 after 1.5 days, got %d", c.Day)
	}
	if math.Abs(c.TimeOfDay()-0.5) > 1e-9 {
		t.Fatalf("expected TimeOfDay=0.5, got %v", c.TimeOfDay())
	}

	c.Advance(500)
	if c.Day != 6 {
		t.Fatalf("expected Day=6 after 5 more days, got %d", c.Day)
	}
}

func TestAdvanceIgnoresNonPositiveDelta(t *testing.T) {
	c := NewClock(DefaultSettings())
	c.Advance(10)
	before := c.Elapsed
	c.Advance(0)
	c.Advance(-5)
	if c.Elapsed != before {
		t.Fatalf("expected no change from non-positive dt, got %v want %v", c.Elapsed, before)
	}
}

func TestSetTimeOfDayWrapsAndRoundTrips(t *testing.T) {
	c := NewClock(Settings{DaySeconds: 100})
	c.SetTimeOfDay(0.75)
	if math.Abs(c.TimeOfDay()-0.75) > 1e-9 {
		t.Fatalf("got %v, want 0.75", c.TimeOfDay())
	}

	c.SetTimeOfDay(-0.25)
	if math.Abs(c.TimeOfDay()-0.75) > 1e-9 {
		t.Fatalf("negative wrap: got %v, want 0.75", c.TimeOfDay())
	}

	c.SetTimeOfDay(1.25)
	if math.Abs(c.TimeOfDay()-0.25) > 1e-9 {
		t.Fatalf(">1 wrap: got %v, want 0.25", c.TimeOfDay())
	}
}

func TestPhaseTransitions(t *testing.T) {
	s := DefaultSettings()
	c := NewClock(s)

	cases := []struct {
		frac float64
		want Phase
	}{
		{0.0, Night},
		{s.DawnStart, Dawn},
		{(s.DawnStart + s.DawnEnd) / 2, Dawn},
		{s.DawnEnd, Day},
		{0.5, Day},
		{s.DuskStart, Dusk},
		{(s.DuskStart + s.DuskEnd) / 2, Dusk},
		{s.DuskEnd, Night},
		{0.99, Night},
	}
	for _, tc := range cases {
		c.SetTimeOfDay(tc.frac)
		if got := c.Phase(); got != tc.want {
			t.Errorf("frac=%v: got phase %v, want %v", tc.frac, got, tc.want)
		}
	}
}

func TestIsNightIsDay(t *testing.T) {
	c := NewClock(DefaultSettings())

	c.SetTimeOfDay(0.5)
	if !c.IsDay() || c.IsNight() {
		t.Fatalf("expected noon to be day, not night")
	}

	c.SetTimeOfDay(0.0)
	if !c.IsNight() || c.IsDay() {
		t.Fatalf("expected midnight to be night, not day")
	}
}

func TestSunHeightPeaksAtNoonAndTroughsAtMidnight(t *testing.T) {
	c := NewClock(DefaultSettings())

	c.SetTimeOfDay(0.5)
	if got := c.SunHeight(); math.Abs(got-1) > 1e-9 {
		t.Fatalf("expected sun height 1 at noon, got %v", got)
	}

	c.SetTimeOfDay(0)
	if got := c.SunHeight(); math.Abs(got-(-1)) > 1e-9 {
		t.Fatalf("expected sun height -1 at midnight, got %v", got)
	}

	c.SetTimeOfDay(0.25)
	if got := c.SunHeight(); math.Abs(got) > 1e-9 {
		t.Fatalf("expected sun height ~0 at 6am, got %v", got)
	}
}

func TestTimeOfDayZeroDaySecondsIsSafe(t *testing.T) {
	c := NewClock(Settings{})
	c.Advance(10)
	if c.TimeOfDay() != 0 {
		t.Fatalf("expected 0 with zero-length day, got %v", c.TimeOfDay())
	}
}
