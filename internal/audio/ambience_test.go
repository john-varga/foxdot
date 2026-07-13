package audio

import (
	"testing"

	"foxdot/internal/worldtime"
)

func TestAmbienceWeightsDayHasBirdsNoCrickets(t *testing.T) {
	w := AmbienceWeights(worldtime.Day, 1)
	if w["birds"] <= 0 {
		t.Fatalf("expected birds during the day, got %v", w)
	}
	if w["crickets"] > 0 {
		t.Fatalf("did not expect crickets during the day, got %v", w)
	}
}

func TestAmbienceWeightsNightHasCricketsNoBirds(t *testing.T) {
	w := AmbienceWeights(worldtime.Night, 1)
	if w["crickets"] <= 0 || w["frogs"] <= 0 {
		t.Fatalf("expected crickets and frogs at night, got %v", w)
	}
	if _, ok := w["birds"]; ok {
		t.Fatalf("did not expect birds at night, got %v", w)
	}
}

func TestAmbienceWeightsAlwaysHasWindBed(t *testing.T) {
	for _, phase := range []worldtime.Phase{worldtime.Night, worldtime.Dawn, worldtime.Day, worldtime.Dusk} {
		w := AmbienceWeights(phase, 1)
		if w["wind"] <= 0 {
			t.Errorf("phase %v: expected a wind bed, got %v", phase, w)
		}
	}
}

func TestCampfireWeightFadesWithProximity(t *testing.T) {
	far := AmbienceWeights(worldtime.Day, 1)
	if _, ok := far["campfire"]; ok {
		t.Fatalf("expected no campfire loop far away, got %v", far)
	}

	near := AmbienceWeights(worldtime.Day, 0)
	mid := AmbienceWeights(worldtime.Day, 0.5)

	if near["campfire"] <= mid["campfire"] {
		t.Fatalf("expected campfire weight to decrease with distance: near=%v mid=%v", near["campfire"], mid["campfire"])
	}
	if mid["campfire"] <= 0 {
		t.Fatalf("expected some campfire weight at mid distance, got %v", mid["campfire"])
	}
}

func TestChooseMusicTrackFollowsTimeOfDayWhenExploring(t *testing.T) {
	if got := ChooseMusicTrack(SituationExplore, worldtime.Day); got != "forest_day" {
		t.Fatalf("got %q, want forest_day", got)
	}
	if got := ChooseMusicTrack(SituationExplore, worldtime.Night); got != "forest_night" {
		t.Fatalf("got %q, want forest_night", got)
	}
	// Dawn/dusk aren't full night; still counts as day music.
	if got := ChooseMusicTrack(SituationExplore, worldtime.Dawn); got != "forest_day" {
		t.Fatalf("got %q, want forest_day at dawn", got)
	}
}

func TestChooseMusicTrackSituationOverridesTimeOfDay(t *testing.T) {
	cases := []struct {
		situation Situation
		want      string
	}{
		{SituationMainMenu, "main_menu"},
		{SituationDiscovery, "discovery"},
		{SituationDanger, "quiet_danger"},
		{SituationCredits, "credits"},
	}
	for _, tc := range cases {
		if got := ChooseMusicTrack(tc.situation, worldtime.Night); got != tc.want {
			t.Errorf("situation %v: got %q, want %q", tc.situation, got, tc.want)
		}
	}
}
