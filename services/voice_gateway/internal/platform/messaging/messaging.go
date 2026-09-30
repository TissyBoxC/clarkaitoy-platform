// Package messaging owns asynchronous message publishing.
package messaging

// Publisher publishes service events.
type Publisher interface {
	Publish(topic string, payload []byte) error
}
