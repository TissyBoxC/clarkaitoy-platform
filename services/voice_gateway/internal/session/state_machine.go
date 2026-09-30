package session

// State identifies the current voice session phase.
type State string

const (
	StateIdle      State = "idle"
	StateListening State = "listening"
	StateThinking  State = "thinking"
	StateSpeaking  State = "speaking"
)
