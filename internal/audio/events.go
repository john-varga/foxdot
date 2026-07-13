package audio

// SoundEvent is a request to play a one-shot sound by catalog name, queued
// by gameplay/UI/creature code and drained by the audio system once per
// frame. Decoupling "something happened that makes noise" from "how the
// engine plays it" means the sim/world/UI never need to import raylib.
type SoundEvent struct {
	Name   string
	Volume float32 // 0..1; treated as 1 if left at the zero value
}

// EventQueue buffers SoundEvents for one frame. Register it as an ecs
// resource (ecs.SetResource) so any system can push to it.
type EventQueue struct {
	events []SoundEvent
}

// Push queues a sound by catalog name at volume (0..1; pass 0 for the
// default of full volume).
func (q *EventQueue) Push(name string, volume float32) {
	if volume <= 0 {
		volume = 1
	}
	q.events = append(q.events, SoundEvent{Name: name, Volume: volume})
}

// Drain returns every queued event and empties the queue.
func (q *EventQueue) Drain() []SoundEvent {
	if len(q.events) == 0 {
		return nil
	}
	out := q.events
	q.events = nil
	return out
}
