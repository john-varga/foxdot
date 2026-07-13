package audio

import "testing"

func TestEventQueuePushAndDrain(t *testing.T) {
	var q EventQueue
	q.Push("jump", 1)
	q.Push("ui_click", 0)

	got := q.Drain()
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
	if got[0].Name != "jump" || got[0].Volume != 1 {
		t.Errorf("got %+v", got[0])
	}
	if got[1].Name != "ui_click" || got[1].Volume != 1 {
		t.Errorf("expected default volume 1 for zero input, got %+v", got[1])
	}
}

func TestEventQueueDrainEmptiesQueue(t *testing.T) {
	var q EventQueue
	q.Push("jump", 1)
	q.Drain()

	if got := q.Drain(); got != nil {
		t.Fatalf("expected nil after draining an already-empty queue, got %v", got)
	}
}

func TestEventQueueDrainOnEmptyQueueIsNil(t *testing.T) {
	var q EventQueue
	if got := q.Drain(); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}
