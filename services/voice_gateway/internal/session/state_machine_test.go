package session

import "testing"

func TestConversationLifecycle(t *testing.T) {
	state := StateIdle

	steps := []struct {
		event Event
		want  State
	}{
		{event: EventStartListening, want: StateListening},
		{event: EventStartThinking, want: StateThinking},
		{event: EventStartSpeaking, want: StateSpeaking},
		{event: EventReset, want: StateIdle},
	}

	for _, step := range steps {
		next, err := state.Transition(step.event)
		if err != nil {
			t.Fatalf("transition %s from %s failed: %v", step.event, state, err)
		}
		if next != step.want {
			t.Fatalf("transition %s: expected %s, got %s", step.event, step.want, next)
		}
		state = next
	}
}

func TestInvalidTransitionIsRejected(t *testing.T) {
	if _, err := StateIdle.Transition(EventStartThinking); err == nil {
		t.Fatal("expected idle -> thinking to be rejected")
	}
}
